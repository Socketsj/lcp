package main

import (
	"fmt"
	"github.com/Socketsj/lcp/client"
	"net"
	"os"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		panic("need 2 args")
	}
	addr := args[1]
	name := args[2]
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	f, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	var args1 []string
	if len(args) >= 4 {
		args1 = args[3:]
	}

	c := client.NewClientConn(conn, f)
	c.Server(args1...)
}
