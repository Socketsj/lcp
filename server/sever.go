package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Socketsj/lcp/common"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

const (
	HEAD_SIZE = 4
	CONTENT_SIZE = 8
	BLOCK_SIZE = 32768
)

const (
	status_init = 0
	status_active = 1
	status_stop = 2
)

const (
	task_add = 1
	task_del = 2
)

type SeverTask struct {
	sc *SeverConn
	s  int
}

type SeverConn struct {
	id   string
	size uint64
	conn net.Conn
	path string
	file *bufio.Writer
	status int32
}

func NewConn(conn net.Conn, dir string) *SeverConn {
	return &SeverConn{
		id: RandId(),
		conn: conn,
		path: dir,
	}
}

func (s *SeverConn) Server(	taskCh chan SeverTask) {
	f, err := s.handshake()
	if err != nil {
		s.conn.Close()
		fmt.Println("fail to handshake, err:", err)
		return
	}
	defer f.Close()
	defer s.close()
	defer func() {
		taskCh <- SeverTask{sc: s, s: task_del}
	}()
	taskCh <- SeverTask{sc: s, s: task_add}

	s.file = bufio.NewWriterSize(f, BLOCK_SIZE * 4)
	s.status = status_active
	block := make([]byte, BLOCK_SIZE)
	var size uint64
	for {
		n, err := s.conn.Read(block)
		if err != nil && err != io.EOF {
			fmt.Println("fail to accept, err:", err)
			return
		}
		if n == 0 {
			break
		}
		s.file.Write(block[:n])
		size += uint64(n)
		if size >= s.size{
			break
		}
	}

}

func (s *SeverConn) closeConn() {
	if atomic.CompareAndSwapInt32(&s.status, status_active, status_stop) {
		s.conn.Close()
	}
}

func (s *SeverConn) close() {
	if atomic.CompareAndSwapInt32(&s.status, status_active, status_stop) {
		s.conn.Close()
		if s.file != nil {
			s.file.Flush()
		}
	}
}

func (s *SeverConn) handshake() (*os.File, error) {
	content, err := common.HandShakeRecv(s.conn)
	if err != nil {
		s.sendError()
		return nil, err
	}
	size := len(content)
	if size < common.HS_HEAD_SIZE + common.HS_MSG_SIZE {
		s.sendError()
		return nil, errors.New("size too small")
	}
	s.size = common.Endian.Uint64(content[:common.HS_MSG_SIZE])
	args := common.ParserHsBytes(content[common.HS_MSG_SIZE:])
	if len(args) == 0 {
		s.sendError()
		return nil, errors.New("size too small")
	}
	name := args[0]
	dir := s.path
	if len(args) >= 2 {
		dir = args[1]
	}
	s.path = filepath.Join(dir, name)
	f, err := os.Create(s.path)
	if err != nil {
		s.sendError()
		return nil, err
	}
	bs := make([]byte, 9)
	binary.LittleEndian.PutUint32(bs[:4], 5)
	binary.LittleEndian.PutUint32(bs[5:], BLOCK_SIZE)
	_, err = s.conn.Write(bs)
	if err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func (s *SeverConn) sendError() {
	bs := make([]byte, 5)
	binary.LittleEndian.PutUint32(bs[:4], 1)
	bs[4] = 1
	s.conn.Write(bs)
}

func (s *SeverConn) read() ([]byte, error) {
	head := make([]byte, HEAD_SIZE)
	_, err := io.ReadFull(s.conn, head)
	if err != nil {
		return nil, err
	}
	headSize := binary.LittleEndian.Uint32(head)
	content := make([]byte, headSize)
	_, err = io.ReadFull(s.conn, content)
	return content, err
}

func init()  {
	rand.Seed(time.Now().Unix())
}

func RandId() string {
	ts := time.Now().Unix()
	id := rand.Int31n(9999)
	return fmt.Sprintf("%d%4d", ts, id)
}
