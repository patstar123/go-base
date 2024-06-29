package utils

import (
	"encoding/json"
	"strconv"
	"time"
)

func GetUnmarshalU8(in any) uint8 {
	switch v := in.(type) {
	case uint8:
		return v
	case string:
		value, _ := strconv.ParseUint(v, 10, 8)
		return uint8(value)
	case float64:
		return uint8(v)
	default:
		panic(v)
	}
}

func GetUnmarshalU16(in any) uint16 {
	switch v := in.(type) {
	case uint16:
		return v
	case string:
		value, _ := strconv.ParseUint(v, 10, 16)
		return uint16(value)
	case float64:
		return uint16(v)
	default:
		panic(v)
	}
}

func GetUnmarshalU32(in any) uint32 {
	switch v := in.(type) {
	case uint32:
		return v
	case string:
		value, _ := strconv.ParseUint(v, 10, 32)
		return uint32(value)
	case float64:
		return uint32(v)
	default:
		panic(v)
	}
}

func GetUnmarshalI32(in any) int32 {
	switch v := in.(type) {
	case int32:
		return v
	case string:
		value, _ := strconv.ParseInt(v, 10, 32)
		return int32(value)
	case float64:
		return int32(v)
	default:
		panic(v)
	}
}

func GetUnmarshalI64(in any) int64 {
	switch v := in.(type) {
	case int64:
		return v
	case string:
		value, _ := strconv.ParseInt(v, 10, 64)
		return value
	case float64:
		return int64(v)
	default:
		panic(v)
	}
}

func GetUnmarshalBool(in any) bool {
	if in == nil {
		return false
	}
	switch v := in.(type) {
	case bool:
		return v
	case string:
		if in == "" {
			return false
		} else if in == "false" {
			return false
		} else if in == "true" {
			return true
		}
		value, _ := strconv.ParseUint(v, 10, 32)
		return uint32(value) != 0
	case float64:
		return v != 0
	default:
		panic(v)
	}
}

func GetUnmarshalTime(in any) time.Time {
	switch v := in.(type) {
	case int64:
		return time.UnixMilli(v)
	case string:
		value, _ := strconv.ParseInt(v, 10, 64)
		return time.UnixMilli(value)
	case float64:
		return time.UnixMilli(int64(v))
	default:
		panic(v)
	}
}

func GetUnmarshalDuration(in any) time.Duration {
	return time.Duration(GetUnmarshalI64(in))
}

func GetUnmarshalStringArray(in any) []string {
	switch vs := in.(type) {
	case []string:
		return vs
	case []any:
		s := make([]string, 0, len(vs))
		for _, v := range vs {
			s = append(s, v.(string))
		}
		return s
	default:
		panic(vs)
	}
}

func UnmarshalAny2Any(in any, out any) error {
	tmp, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(tmp, out)
}
