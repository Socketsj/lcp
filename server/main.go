package main

import (
	"flag"
	"fmt"
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
	connMap := make(map[string]*SeverConn)
	taskConn := make(chan SeverTask, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go func(){
		defer wg.Done()
		for {
			conn, err := listen.Accept()
			if err != nil {
				return
			}
			sc := NewConn(conn, *path)
			go sc.Server(taskConn)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case task := <- taskConn:
				switch task.s {
				case task_add:
					connMap[task.sc.id] = task.sc
				case task_del:
					delete(connMap, task.sc.id)
				}
			case <-done:
				for _, sc := range connMap {
					sc.closeConn()
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
