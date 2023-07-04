package main

import (
	"fmt"
	"log"
	"my-redis-go/operations"
	"my-redis-go/resp"
	"net"
)

//TODO: use config to pass host and port and other params
// func setupFlags() {
// 	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the dice server")
// 	flag.IntVar(&config.Port, "port", 7379, "port for the dice server")
// 	flag.Parse()
// }

type Server struct {
	listenAddr string
	listener   net.Listener
	quitCh     chan struct{}
}

func spinServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitCh:     make(chan struct{}),
	}
}

func (server *Server) Start() error {
	listener, err := net.Listen("tcp", server.listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()
	server.listener = listener

	go server.establishConnection()

	<-server.quitCh

	return nil
}

func (server *Server) establishConnection() {
	for {
		fmt.Println("Accepting New Connections\n")
		conn, err := server.listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("New Connection:", conn.RemoteAddr())
		go server.requestHandler(conn)
	}
}

func (server *Server) requestHandler(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 2048)
		n, err := conn.Read(buf)
		if err != nil {
			print("hello error")
			log.Fatal(err)
		}
		msg := buf[:n]
		// print("hello")
		// print(string(msg))
		// echo := "+OK\r\n"
		commands, _ := resp.ParseRequest(msg)
		for _, command := range commands {
			dataType, err := operations.ExecuteCommand(command)
			fmt.Println(dataType.ToString())
			if err != resp.EmptyRedisError {
				fmt.Println("else error")
				conn.Write([]byte(err.ToString() + "\n"))
			} else {
				fmt.Println("else")
				if dataType == nil {
					fmt.Println("else nil")
					conn.Write([]byte("(nil)" + "\n"))
				} else {
					fmt.Println("elsewow")
					out := fmt.Sprintf("$%d\r\n%s\r\n", len(dataType.ToString()), dataType.ToString())
					fmt.Println(out)
					conn.Write([]byte(out))
				}
			}
		}

		// conn.Write([]byte(echo))
		// log.Println("Done Writing!")
		fmt.Println(string(msg))
	}
}

func main() {
	server := spinServer(":4000")
	server.Start()
}
