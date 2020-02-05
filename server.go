package main

import (
	"flag"
	"fmt"
	"github.com/Socketsj/lcp/server"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	addr := flag.String("addr", "0.0.0.0:44396", "listen at TCP")
	path := flag.String("path", "", "store file path")
	flag.Parse()

	if _,err := os.Stat(*path); err != nil {
		fmt.Println(*path)
		panic(err)
	}

	listen, err := net.Listen("tcp", *addr)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("start server addr=%s, path=%s", *addr, *path))

	done := make(chan struct{})
	connMap := make(map[string]*server.Conn)
	taskConn := make(chan server.SeverTask, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go func(){
		defer wg.Done()
		for {
			conn, err := listen.Accept()
			if err != nil {
				return
			}
			sc := server.NewServerConn(conn, *path)
			go sc.Server(taskConn)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case task := <- taskConn:
				switch task.S {
				case server.Task_add:
					connMap[task.Sc.Id()] = task.Sc
				case server.Task_del:
					delete(connMap, task.Sc.Id())
				}
			case <-done:
				for _, sc := range connMap {
					sc.CloseConn()
				}
				return
			}
		}
	}()

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigs
	listen.Close()
	close(done)
	wg.Wait()
	fmt.Println("stop server")
}
