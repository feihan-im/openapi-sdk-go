package fhcore

import "encoding/json"

var (
	JsonMarshal   func(v interface{}) ([]byte, error)    = json.Marshal
	JsonUnmarshal func(data []byte, v interface{}) error = json.Unmarshal
)

func Pretty(obj interface{}) string {
	s, _ := json.MarshalIndent(obj, "", "  ")
	return string(s)
}
