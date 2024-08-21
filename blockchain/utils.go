package blockchain

import (
	"bytes"
	"encoding/binary"
)

func IntToBytes(n int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, int64(n))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
