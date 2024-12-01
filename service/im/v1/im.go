package fhim

import fhcore "github.com/feihan-im/openapi-sdk-go/core"

type V1 struct {
	Message *v1Message
}

func New(config *fhcore.Config) *V1 {
	return &V1{
		Message: &v1Message{config: config, Event: &v1MessageEvent{config: config}},
	}
}
