package keepaliveconn

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func fileServer(tcp string) {
	listener, err := net.Listen("tcp", tcp)
	if err != nil {
		panic(fmt.Errorf("usbmuxd: fail to listen on: %v, error:%v", tcp, err))
	}

	for {
		_conn, err := listener.Accept()
		if err != nil {
			fmt.Println(fmt.Errorf("usbmuxd: fail to listen accept: %v", err))
			continue
		}

		conn2, ok := _conn.(*net.TCPConn)
		if !ok {
			panic("conn convert to TCPConn failed")
		}
		conn := New(conn2, time.Duration(10))
		go func() {
			defer conn.Close()
			f, err := os.Create("file.bin")
			if err != nil {
				panic(err)
			}
			defer f.Close()

			for {
				n, err := io.Copy(f, conn)
				fmt.Printf("file write: %v bytes, err:%v\n", n, err)
				return
			}
		}()
	}
}

func fileClient(server, filePath string) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		fmt.Printf("dial failed at %v by err:%v\n", server, err)
		return
	}
	conn = New(conn, time.Duration(10))
	defer conn.Close()

	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// b := make([]byte, 1024)
	_, err = io.Copy(conn, f)
	if err != nil {
		panic(err)
	}
}
