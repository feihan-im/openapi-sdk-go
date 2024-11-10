package fhcore

type Config struct {
	AppId      string
	AppSecret  string
	BackendUrl string

	HttpClient HttpClient
	ApiClient  ApiClient

	Logger Logger
}
