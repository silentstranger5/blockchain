package blockchain

import (
	"bytes"
	"encoding/binary"
)

func IntToBytes(n int) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, int64(n))
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}
