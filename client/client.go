package client

import (
	"errors"
	"fmt"
	"github.com/Socketsj/lcp/common"
	"io"
	"net"
	"os"
	"time"
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
	stat, err := c.file.Stat()
	if err != nil {
		fmt.Println("failed to File ")
		return
	}

	total := stat.Size()
	var sum uint64
	name := stat.Name()
	var collect, diff, start int64
	var sspeed string
	for {
		if start == 0 || diff != 0 {
			start = time.Now().UnixNano()
			collect = 0
		}

		n, err := c.send(block)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("failed to send err:", err)
			return
		}
		collect += int64(n)
		diff = time.Now().UnixNano() - start
		if diff > 0 {
			sspeed = fmt.Sprintf("%.2fKB/s", float64(collect) / float64(diff) * 1e9 / 1024)
		}
		sum += uint64(n)
		fmt.Fprint(os.Stdout, fmt.Sprintf("%s\t\t\t\t\t%s\t\t\t\t\t%.2f%%\r", name, sspeed, 100*float64(sum)/float64(total)))
	}

}

func(c *Conn) send(block [] byte) (int, error) {
	n, err := c.file.Read(block)
	if err != nil && err != io.EOF {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}
	_, err = c.conn.Write(block[:n])
	return n, err
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