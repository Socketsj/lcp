package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

func handshake(conn net.Conn, f *os.File) (uint32, error) {
	s, err := f.Stat()
	if err != nil {
		return 0, err
	}
	name := []byte(s.Name())
	buf := new(bytes.Buffer)
	head := make([]byte, 8)
	binary.LittleEndian.PutUint32(head[:4],8 + uint32(len(name)))
	buf.Write(head[:4])
	binary.LittleEndian.PutUint64(head, uint64(s.Size()))
	buf.Write(head)
	buf.Write(name)
	_, err = buf.WriteTo(conn)
	if err != nil {
		return 0, err
	}

	_, err = io.ReadFull(conn, head[:4])
	if err != nil {
		return 0, err
	}
	size := binary.LittleEndian.Uint32(head[:4])
	content := make([]byte, size)
	_, err = io.ReadFull(conn, content)
	if err != nil {
		return 0, err
	}
	blockSize := binary.LittleEndian.Uint32(content[:4])
	return blockSize, nil
}

func main() {
	args := os.Args
	if len(args) < 3 {
		panic("need 2 args")
	}
	addr := args[1]
	name := args[2]
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	size, err := handshake(conn, f)
	if err != nil {
		panic(err)
	}

	block := make([]byte, size)
	for {
		n, err := f.Read(block)
		if err != nil && err != io.EOF {
			fmt.Println("read file error, err:", err)
		}
		if n == 0 {
			break
		}
		_, err = conn.Write(block[:n])
		if err != nil {
			fmt.Println(err)
			return
		}
	}

}
