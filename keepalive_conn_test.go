package main

import (
	"fmt"
	"io"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
)

func unix_to_tcp(tcp, unix string) error {
	listener, err := net.Listen("unix", unix)
	if err != nil {
		return fmt.Errorf("usbmuxd: fail to listen on: %v, error:%v", unix, err)
	}

	os.Chmod(unix, 0777)
	log.Infoln("listen on: ")
	for {
		src, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("usbmuxd: fail to listen accept: %v", err)
		}

		dstConn, err := net.Dial("tcp", tcp)

		src2 := NewKeepaliveConn(src)

		go func() {
			io.Copy(dstConn, src)
			dstConn.Close()
		}()
		go func() {
			io.Copy(src, dstConn)
			src.Close()
		}()
	}
}

func startAsUsbmuxd() {
	unix_to_tcp("127.0.0.1:27015", "/var/run/usbmuxd")
}
