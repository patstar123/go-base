package utils

import (
	"encoding/json"
	"fmt"
)

const maxStrBytes = 100 * 1000 * 1000 // 100M

func UnmarshalJson(data []byte, v any) error {
	if len(data) > maxStrBytes {
		return fmt.Errorf("UnmarshalJson failed, input is too large %v bytes", len(data))
	}

	return json.Unmarshal(data, v)
}
