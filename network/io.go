package network

import (
	"errors"
	"io"
	"net"
	"os"
	"time"
)

func Copy(dst net.Conn, src *KeepaliveConn) (written int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		src.SetReadDeadline(time.Now().Add(src.keepaliveInterval))
		nr, er := src.Read(buf)
		if errors.Is(er, os.ErrDeadlineExceeded) {
			src.Heartbeat([]byte("heart beat"))
			continue
		}

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
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
