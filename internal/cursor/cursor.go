package cursor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

func Encode(data any) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to encode data to JSON: %s", err.Error()))
	}
	base64 := base64.URLEncoding.EncodeToString(jsonData)
	return base64, nil
}

func Decode(in string, to any) error {
	if in == "" {
		return nil
	}
	jsonData, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to decode data: %s", err.Error()))
	}
	if err := json.Unmarshal(jsonData, to); err != nil {
		return errors.New(fmt.Sprintf("failed to decode data to JSON: %s", err.Error()))
	}
	return nil
}
