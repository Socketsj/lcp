package client

import (
	"errors"
	"fmt"
	"github.com/Socketsj/lcp/common"
	"io"
	"net"
	"os"
)

type Conn struct {
	conn  net.Conn
	file  *os.File
}

func NewClientConn(conn net.Conn, file *os.File) *Conn {
	return &Conn{
		conn: conn,
		file: file,
	}
}

func(c *Conn) Server(args ...string)  {
	size, err := c.handshake(args...)
	if err != nil {
		fmt.Println("failed to handshake err:", err)
		return
	}

	block := make([]byte, size)
	for {
		n, err := c.file.Read(block)
		if err != nil && err != io.EOF {
			fmt.Println("failed to read file err", err)
			return
		}
		if n == 0 {
			break
		}
		_, err = c.conn.Write(block[:n])
		if err != nil {
			fmt.Println(err)
			return
		}
	}

}

func(c *Conn) handshake(args ...string) (uint32, error) {
	stat, err := c.file.Stat()
	if err != nil {
		return 0, err
	}
	args = append([]string{stat.Name()}, args...)
	bs := make([]byte, common.HS_MSG_SIZE)
	common.Endian.PutUint64(bs, uint64(stat.Size()))
	err = common.HandShakeSend(c.conn, bs, args...)
	if err != nil {
		return 0, err
	}
	bs, err = common.HandShakeRecv(c.conn)
	if err != nil {
		return 0, err
	}
	if bs[0] != common.HS_SUC {
		return 0, errors.New("failed to handshake")
	}
	blockSize := common.Endian.Uint32(bs[common.HS_STATUS_SIZE:common.HS_STATUS_SIZE+common.HS_HEAD_SIZE])
	return blockSize, nil
}