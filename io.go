package keepaliveconn

import (
	"errors"
	"io"
	"net"
	"time"
)

func Copy(dst net.Conn, src *KeepaliveConn) (written int64, err error) {
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
