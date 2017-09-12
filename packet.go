package retranslator

import "encoding/json"

type Packet struct {
	StatusCode int
	Status     string
	Header     map[string][]string
	Body       []byte
}

func (p Packet) GetBytes() ([]byte, error) {
	return json.Marshal(p)
}
