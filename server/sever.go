package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
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
	content, err := s.read()
	if err != nil {
		return nil, err
	}
	size := len(content)
	if size < CONTENT_SIZE {
		return nil, errors.New("size too small")
	}
	s.size = binary.LittleEndian.Uint64(content[0:CONTENT_SIZE])
	s.path = filepath.Join(s.path, string(content[CONTENT_SIZE:]))
	f, err := os.Create(s.path)
	if err != nil {
		return nil, err
	}
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint32(bs[:4], 4)
	binary.LittleEndian.PutUint32(bs[4:], BLOCK_SIZE)
	_, err = s.conn.Write(bs)
	if err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
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