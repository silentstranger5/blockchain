package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func (h *Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%x", *h))
}

func (h *Hash) UnmarshalJSON(data []byte) error {
	trimmed := string(bytes.Trim(data, "\""))
	decoded, err := hex.DecodeString(trimmed)
	if err != nil {
		return err
	}
	*h = (Hash)(decoded)
	return nil
}
