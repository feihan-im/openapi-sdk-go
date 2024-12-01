package fhcore

type Marshaller = func(v interface{}) ([]byte, error)
type Unmarshaller = func(data []byte, v interface{}) error
