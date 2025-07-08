package jsonx

import (
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// MarshalJSON 序列化
func MarshalJSON(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// UnmarshalJSON 反序列化
func UnmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
