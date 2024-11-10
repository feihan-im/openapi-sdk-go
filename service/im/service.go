package fhserviceim

import (
	fhcore "github.com/feihan-im/openapi-sdk-go/core"
	v1 "github.com/feihan-im/openapi-sdk-go/service/im/v1"
)

type Service struct {
	*v1.V1
}

func New(config *fhcore.Config) *Service {
	return &Service{
		V1: v1.New(config),
	}
}
