package fhcore

import "time"

type Config struct {
	AppId      string
	AppSecret  string
	BackendUrl string

	HttpClient       HttpClient
	ApiClient        ApiClient
	EnableEncryption bool
	RequestTimeout   time.Duration

	Logger Logger
}
