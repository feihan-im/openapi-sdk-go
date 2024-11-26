package fhcore

import "encoding/json"

var (
	JsonMarshal   func(v any) ([]byte, error)    = json.Marshal
	JsonUnmarshal func(data []byte, v any) error = json.Unmarshal
)

func Pretty(obj any) string {
	s, _ := json.MarshalIndent(obj, "", "  ")
	return string(s)
}
