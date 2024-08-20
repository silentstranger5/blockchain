package blockchain

import (
	"bytes"
	"encoding/binary"
)

func IntToBytes(n int64) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, n)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
