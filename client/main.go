package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func handshake(conn net.Conn, f *os.File, args []string) (uint32, error) {
	s, err := f.Stat()
	if err != nil {
		return 0, err
	}
	name := []byte(s.Name())
	head := make([]byte, 4)
	resp := make([]byte, 8)
	binary.LittleEndian.PutUint64(resp, uint64(s.Size()))
	binary.LittleEndian.PutUint32(head, uint32(len(name)))
	resp = append(resp, head...)
	resp = append(resp, name...)
	if len(args) >= 4 {
		dir := []byte(args[3])
		binary.LittleEndian.PutUint32(head, uint32(len(dir)))
		resp = append(resp, head...)
		resp = append(resp, dir...)
	}
	binary.LittleEndian.PutUint32(head, uint32(len(resp)))
	_, err = conn.Write(append(head, resp...))
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
	if content[0] != 0 {
		return 0, errors.New("failed to handshake")
	}
	blockSize := binary.LittleEndian.Uint32(content[1:5])
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

	size, err := handshake(conn, f, args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)	}

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
