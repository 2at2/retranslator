package module

import (
	"encoding/json"
	"github.com/2at2/retranslator"
	"github.com/gorilla/websocket"
	"github.com/mono83/slf/wd"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/2at2/retranslator/receiver"
)

var upgrader = websocket.Upgrader{} // use default options
var log = wd.NewLogger("handler")

type Handler struct {
	Hub *receiver.Hub
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info("Receiver connected")

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("Upgrader error - :err", wd.ErrParam(err))
		return
	}

	transport := &WebsocketTransport{c: conn}
	defer transport.Close()
	defer log.Warning("Transport is closed")

	h.Hub.Register(transport)

	// Initialize
	log.Info("Waiting for init message")

	bts, err := transport.Read()

	if err != nil {
		log.Error("Init error - :err", wd.ErrParam(err))
		return
	}

	var init retranslator.ClientInitialization

	err = json.Unmarshal(bts, &init)

	if err != nil {
		log.Error("Unable to unmarshal init - :err", wd.ErrParam(err))
		return
	}

	if len(init.Path) == 0 {
		log.Error("Empty path")
		return
	}

	addr := ":" + strconv.Itoa(init.Port)

	log.Info("Serve :addr/:path", wd.StringParam("addr", addr), wd.StringParam("path", init.Path))

	// New server listener
	server := &http.Server{Addr: addr, Handler: InnerHandler{
		transport: transport,
	}}

	defer server.Close()

	// Ping
	go func() {
		for transport.IsAlive() {
			if err := transport.Ping(); err != nil {
				log.Error("Failed ping")
			}

			time.Sleep(time.Millisecond * 500)
		}

		server.Close()
		return
	}()

	// Listen
	if err := server.ListenAndServe(); err != nil {
		log.Error("HTTP server error - :err", wd.ErrParam(err))
	}
}

type InnerHandler struct {
	transport *WebsocketTransport
}

// ServeHTTP handler of http request
func (h InnerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info("Received callback from :addr", wd.StringParam("addr", r.RemoteAddr))

	err := h.handleCallback(w, r)

	if err != nil {
		log.Error("Callback handler error - :err", wd.ErrParam(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// handleCallback process callback request
func (h *InnerHandler) handleCallback(w http.ResponseWriter, r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Error("Unable to read body - :err", wd.ErrParam(err))
		return err
	}

	defer r.Body.Close()

	p := retranslator.Packet{
		Header: r.Header,
		Body:   body,
	}

	jsn, err := p.GetBytes()

	if err != nil {
		log.Error("Failed to marshal packet - :err", wd.ErrParam(err))
	}

	if err := h.transport.Write(jsn); err != nil {
		log.Error("Failed write - :err", wd.ErrParam(err))
		return err
	}

	bts, err := h.transport.Read()

	if err != nil {
		log.Error("Response error - :err", wd.ErrParam(err))
		return err
	}

	var response retranslator.Packet

	err = json.Unmarshal(bts, &response)

	if err != nil {
		log.Error("Failed to unmarshal response - :err", wd.ErrParam(err))
		return err
	}

	w.WriteHeader(response.StatusCode)

	for key, values := range response.Header {
		w.Header().Add(key, strings.Join(values, "\n"))
	}

	w.Write(response.Body)

	log.Info("Callback processed")

	return nil
}
