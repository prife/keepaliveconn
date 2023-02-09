package keepaliveconn

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
)

const (
	DataTypePayload      = 0
	DataTypeHeartBeat    = 1
	DataTypeHeartBeatAck = 2
)

type KeepaliveConn struct {
	c                 net.Conn
	buf               bytes.Buffer
	header            []byte
	writeBuf          bytes.Buffer
	writeHeader       []byte
	left              int // 上次包剩余未读取的数据长度
	keepaliveInterval time.Duration
}

func New(c net.Conn, keepaliveInterval time.Duration) *KeepaliveConn {
	return &KeepaliveConn{
		c:                 c,
		header:            make([]byte, 6),
		writeHeader:       make([]byte, 6),
		left:              0,
		keepaliveInterval: keepaliveInterval,
	}
}

func (c *KeepaliveConn) Heartbeat(p []byte) (err error) {
	buf := make([]byte, 6)
	binary.BigEndian.PutUint32(buf, uint32(len(p)))                // 4
	binary.BigEndian.PutUint16(buf[4:], uint16(DataTypeHeartBeat)) // 2
	_, err = c.c.Write(buf)
	if len(p) > 0 {
		_, err = c.c.Write(p)
	}

	return err
}

func (c *KeepaliveConn) heartbeatAck(p []byte) (err error) {
	buf := make([]byte, 6)
	binary.BigEndian.PutUint32(buf, uint32(len(p)))                   // 4
	binary.BigEndian.PutUint16(buf[4:], uint16(DataTypeHeartBeatAck)) // 2
	_, err = c.c.Write(buf)
	if len(p) > 0 {
		_, err = c.c.Write(p)
	}

	return err
}

func (c *KeepaliveConn) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}

	for {
		// 船(c.c) <--> 码头(c.buf) <--> 货车(b)
		if c.left == 0 {
			// 首先确保有缓冲区有一份完整的数据包
			// c.buf 至少应该存放一个包头的大小
			for c.buf.Len() < 6 && err == nil {
				n, err = c.c.Read(b)
				c.buf.Write(b[:n])
			}

			// header都没读全，说明链接被远端关闭了，可能是EOF，也可能是异常断开
			if c.buf.Len() < 6 {
				return 0, err
			}

			// 此时c.buf至少包含了下个包的header
			// 注意err仍然可能存在，先解析包
			_, _ = c.buf.Read(c.header)
			payloadSize := int(binary.BigEndian.Uint32(c.header[:4])) //4
			payloadType := binary.BigEndian.Uint16(c.header[4:])      //2

			// c.buf中不含一个完整包裹，继续读直到含有一个或以上包或者有错误
			for c.buf.Len() < payloadSize && err == nil {
				n, err = c.c.Read(b)
				c.buf.Write(b[:n])
			}

			// 心跳保活模式，此时不需要判断是否写成功
			// 在tcp socket中，读和写是两条不同的链路，可以关闭写，只读；若heartbeat写失败，后续读一定也会失败
			if payloadType != DataTypePayload {
				// 消耗payloadSize
				if payloadSize > 0 {
					b2 := make([]byte, payloadSize)
					c.buf.Read(b2)
				}

				if payloadType == DataTypeHeartBeat {
					c.heartbeatAck(nil)
				}
				continue
			}

			c.left = payloadSize
			// 此时buf中已经有了部分有效数据，但err可能非空
		}

		// 此时码头(c.buf)包含了1个或者以上的完整包裹
		// 注意此时err可能非空
		var toRead int
		if c.left <= len(b) {
			// 2.1 货车(b)很大，可以拉完本次包裹，这里可以优化，继续拆分下一个包裹，为了简单，这里直接返回
			toRead = c.left
		} else {
			// 2.2 货车(b)太小，拉不完本次包裹
			toRead = len(b)
		}

		// 注意这里的返回的Reader的第二个参数表示错误必须丢弃，使用前面传下来的err
		n, _ = c.buf.Read(b[:toRead])
		c.left -= toRead

		return
	}
}

func (c *KeepaliveConn) Write(b []byte) (int, error) {
	binary.BigEndian.PutUint32(c.writeHeader, uint32(len(b)))              // 4
	binary.BigEndian.PutUint16(c.writeHeader[4:], uint16(DataTypePayload)) // 2
	c.writeBuf.Write(c.writeHeader)
	c.writeBuf.Write(b)
	n, err := c.writeBuf.WriteTo(c.c)
	if n >= 6 {
		return int(n - 6), err
	}
	return 0, err
}

func (c *KeepaliveConn) Close() error {
	return c.c.Close()
}

func (c *KeepaliveConn) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}

func (c *KeepaliveConn) RemoteAddr() net.Addr {
	return c.c.RemoteAddr()
}

func (c *KeepaliveConn) SetDeadline(t time.Time) error {
	return c.c.SetDeadline(t)
}

func (c *KeepaliveConn) SetReadDeadline(t time.Time) error {
	return c.c.SetReadDeadline(t)
}

func (c *KeepaliveConn) SetWriteDeadline(t time.Time) error {
	return c.c.SetReadDeadline(t)
}

func Copy(dst net.Conn, src KeepaliveConn) (written int64, err error) {
	for {
		buf := make([]byte, 32*1024)
		src.SetReadDeadline(time.Now().Add(src.keepaliveInterval))
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = errors.New("short write")
				break
			}
		}
		if er != nil {
			if nerr, ok := er.(net.Error); ok && nerr.Timeout() {
				src.Heartbeat([]byte("heart beat"))
				continue
			}
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
