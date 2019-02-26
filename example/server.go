package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/chukeer/httpjsonrpc"
)

type Args struct {
	X, Y int
}

type Calculator struct{}

func (t *Calculator) Add(args *Args, reply *int) error {
	*reply = args.X + args.Y
	return nil
}

func main() {
	cal := new(Calculator)
	server := rpc.NewServer()
	server.Register(cal)
	listener, e := net.Listen("tcp", ":12345")
	if e != nil {
		log.Fatal("listen error:", e)

	}
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())

		} else {
			log.Printf("new connection established\n")
			go server.ServeCodec(httpjsonrpc.NewServerCodec(conn))

		}

	}

}
