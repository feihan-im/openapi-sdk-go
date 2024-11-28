package fhcore

type Config struct {
	AppId      string
	AppSecret  string
	BackendUrl string

	EnableEncryption bool
	HttpClient       HttpClient
	ApiClient        ApiClient

	Logger Logger
}
