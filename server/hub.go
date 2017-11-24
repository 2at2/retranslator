package server

import (
	"sync"
	"time"
)

// Receiver describes transports, that able to receive data
type Transport interface {
	Close() error
	IsAlive() bool
}

type Hub struct {
	m          sync.Mutex
	transports []Transport
}

func (h *Hub) Init() {
	h.transports = []Transport{}

	go func() {
		for {
			time.Sleep(time.Second)

			h.m.Lock()
			nl := []Transport{}
			for _, rec := range h.transports {
				if rec.IsAlive() {
					nl = append(nl, rec)
				}
			}
			h.transports = nl
			h.m.Unlock()
		}
	}()
}

func (h *Hub) Register(receiver Transport) {
	if receiver != nil {
		h.m.Lock()
		defer h.m.Unlock()

		h.transports = append(h.transports, receiver)
	}
}

func (h *Hub) Close() {
	h.m.Lock()
	defer h.m.Unlock()

	for _, rec := range h.transports {
		rec.Close()
	}
}
