package retranslator

import "encoding/json"

type RequestPacket struct {
	Headers    map[string][]string
	Body       []byte
	Method     string
	RequestUri string
	Ip         string
}
func (p RequestPacket) GetBytes() ([]byte, error) {
	return json.Marshal(p)
}

type ResponsePacket struct {
	StatusCode int
	Status     string
	Header     map[string][]string
	Body       []byte
}
func (p ResponsePacket) GetBytes() ([]byte, error) {
	return json.Marshal(p)
}

type ClientInitialization struct {
	Path string
	Port int
}
func (c ClientInitialization) GetBytes() ([]byte, error) {
	return json.Marshal(c)
}
