package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Socketsj/lcp/common"
	"io"
	"net"
	"os"
)

func handshake(conn net.Conn, f *os.File, args []string) (uint32, error) {
	s, err := f.Stat()
	if err != nil {
		return 0, err
	}
	args[2] = s.Name()
	bs := make([]byte, common.HS_MSG_SIZE)
	common.Endian.PutUint64(bs, uint64(s.Size()))
	err = common.HandShakeSend(conn, bs, args[2:]...)
	if err != nil {
		return 0, err
	}
	bs, err = common.HandShakeRecv(conn)
	if err != nil {
		return 0, err
	}
	if bs[0] != 0 {
		return 0, errors.New("failed to handshake")
	}
	blockSize := binary.LittleEndian.Uint32(bs[common.HS_STATUS_SIZE:common.HS_STATUS_SIZE+common.HS_HEAD_SIZE])
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
