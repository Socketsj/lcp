package common

import (
	"encoding/binary"
	"io"
	"net"
)

var Endian = binary.LittleEndian

func HandShakeRecv(conn net.Conn) ([]byte, error) {
	head := make([]byte, HS_HEAD_SIZE)
	_, err := io.ReadFull(conn, head)
	if err != nil {
		return nil, err
	}
	size := Endian.Uint32(head)
	resp := make([]byte, size)
	_, err = io.ReadFull(conn, resp)
	return resp, err
}

func HandShakeSend(conn net.Conn, bs []byte, args ...string) error {
	h := make([]byte, HS_HEAD_SIZE)
	for _, arg := range args {
		line := []byte(arg)
		Endian.PutUint32(h, uint32(len(line)))
		bs = append(bs, h...)
		bs = append(bs, line...)
	}
	Endian.PutUint32(h, uint32(len(bs)))
	_, err := conn.Write(append(h, bs...))
	return err
}

func ParserHsBytes(bs []byte) []string {
	n := len(bs)
	index, next := 0, HS_HEAD_SIZE
	var result []string
	for next < n {
		s := int(Endian.Uint32(bs[index:next]))
		index = next
		next += s
		result = append(result, string(bs[index:next]))
		index = next
		next += HS_HEAD_SIZE
	}
	return result
}