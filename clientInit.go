package retranslator

import "encoding/json"

type ClientInitialization struct {
	Path string
	Port int
}

func (c ClientInitialization) GetBytes() ([]byte, error) {
	return json.Marshal(c)
}
