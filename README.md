# gomlib

go mobile library

## go-keepalive-conn

## Features

- implement net.Conn interface
- inner keepalive supported

## Examples

server

```go
func server(tcp string) {
	listener, err := net.Listen("tcp", tcp)
	if err != nil {
		panic(fmt.Errorf("usbmuxd: fail to listen on: %v, error:%v", tcp, err))
	}

	b := make([]byte, 1024)
	for {
		_conn, err := listener.Accept()
		if err != nil {
			fmt.Println(fmt.Errorf("usbmuxd: fail to listen accept: %v", err))
			continue
		}

		conn, ok := _conn.(*net.TCPConn)
		if !ok {
			panic("conn convert to TCPConn failed")
		}

        // wrap the conn to create a keepalive-conn
		conn = NewKeepaliveConn(conn)
		go func() {
			defer conn.Close()
			for {
				conn.SetReadDeadline(time.Now().Add(time.Second * 3))
				n, err := conn.Read(b)

                // Heartbeat invoked by server, the client will response automaticly
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
                    conn.Heartbeat([]byte("heart beat"))
					fmt.Printf("<-- heart beat --> err:%v\n", err)
					continue
				}

				if err != nil {
					fmt.Printf("<-- read failed: %v", err)
					return
				}

				fmt.Printf("<-- read: %v --> write back", string(b[:n]))
				_, err = conn.Write(b[:n])

				if err != nil {
					fmt.Printf("--> write failed: %v err:%v", err)
					return
				}
			}
		}()
	}
}

func main() {
    server("127.0.0.1:2000")
}
```

client

```go
func client(server string) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		fmt.Errorf("dial failed at %v by err:%v", server, err)
		return
	}
	conn = NewKeepaliveConn(conn)
	defer conn.Close()

	b := make([]byte, 1024)
	_, err = conn.Read(b)
	if err != nil {
		fmt.Errorf("conn addr %v err:%v", conn.RemoteAddr().String(), err)
	}
}

func main() {
    client("127.0.0.1:2000")
}

