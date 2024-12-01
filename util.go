package fhsdk

import "encoding/json"

func Pretty(obj interface{}) string {
	s, _ := json.MarshalIndent(obj, "", "  ")
	return string(s)
}
