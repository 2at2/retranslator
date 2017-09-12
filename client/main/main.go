package main

import (
	"encoding/json"
	"flag"
	"github.com/2at2/retranslator"
	"github.com/2at2/retranslator/client/target"
	"github.com/gorilla/websocket"
	"github.com/mono83/slf/recievers/ansi"
	"github.com/mono83/slf/wd"
	"net/url"
	"os"
	"os/signal"
	"strings"
)

func main() {
	var serverAddr, path, targetAddr string
	var port int

	flag.StringVar(&serverAddr, "serverAddr", "127.0.0.1:8080", "127.0.0.1:8080")
	flag.StringVar(&path, "path", "/api", "callback path")
	flag.IntVar(&port, "port", 80, "callback path")
	flag.StringVar(&targetAddr, "targetAddr", "127.0.0.1:8080", "127.0.0.1:8080")
	flag.Parse()

	if len(serverAddr) == 0 {
		panic("Empty receiver addr")
	}
	if len(path) == 0 {
		panic("Empty path")
	}
	if port < 0 {
		panic("Invalid port")
	}
	if len(targetAddr) == 0 {
		panic("Empty target addr")
	}
	if !strings.HasPrefix(targetAddr, "http") {
		targetAddr = "http://" + targetAddr
	}
	serverAddr = strings.TrimPrefix(serverAddr, "http://")
	serverAddr = strings.TrimPrefix(serverAddr, "https://")
	path = strings.TrimPrefix(path, "/")

	wd.AddReceiver(ansi.New(true, true, false))

	var log = wd.NewLogger("receiver")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)

	log.Info("Target :addr", wd.StringParam("addr", targetAddr))
	log.Info("Server :addr", wd.StringParam("addr", serverAddr))
	log.Info("Serve path *::port/:path", wd.IntParam("port", port), wd.StringParam("path", path))

	receiverUrl := url.URL{Scheme: "ws", Host: serverAddr, Path: "/"}

	log.Info("Connecting to :url", wd.StringParam("url", receiverUrl.String()))

	c, _, err := websocket.DefaultDialer.Dial(receiverUrl.String(), nil)

	if err != nil {
		log.Error("Dial error - :err", wd.ErrParam(err))
		panic(err)
	}

	defer c.Close()

	targetUrl, err := url.Parse(targetAddr)

	if err != nil {
		panic(err)
	}

	terminate := make(chan bool, 1)

	deliver, err := target.NewDeliver(targetAddr)

	if err != nil {
		panic(err)
	}

	// First init
	initClient := retranslator.ClientInitialization{
		Path: path,
		Port: port,
	}

	bt, err := initClient.GetBytes()

	if err != nil {
		panic(err)
	}

	if err := c.WriteMessage(websocket.TextMessage, bt); err != nil {
		log.Error("Failed init - :err", wd.ErrParam(err))
		panic(err)
	}

	message := make(chan []byte, 1)
	closeMessage := make(chan bool)

	// Listening
	go func() {
		for {
			messageType, body, err := c.ReadMessage()

			if err != nil {
				log.Error("read error - :err", wd.ErrParam(err))
				panic(err)
			}

			log.Info("Received message")

			switch messageType {
			case websocket.TextMessage:
				message <- body
				break
			case websocket.CloseMessage:
				closeMessage <- true
				break
			}
		}
	}()

	for {
		select {
		case body := <-message:
			var log = wd.NewLogger("receiver")

			var pack retranslator.Packet

			err := json.Unmarshal(body, &pack)

			if err != nil {
				log.Error("Failed to unmarshal packet - :err", wd.ErrParam(err))
				terminate <- true
				continue
			}

			log.Info("Transferring to :url", wd.StringParam("url", targetUrl.String()))

			packet, err := deliver.Send(pack.Body, pack.Header)

			if err != nil {
				log.Error("Unable to transfer message - :err", wd.ErrParam(err))
				continue
			}

			jsn, err := json.Marshal(packet)

			if err != nil {
				log.Error("Unable to encode body to json- :err", wd.ErrParam(err))
				continue
			}

			if err := c.WriteMessage(websocket.TextMessage, jsn); err != nil {
				log.Error("Failed write - :err", wd.ErrParam(err))
			}
			break

		case <-sigterm:
			log.Warning("Sigterm")
			if err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				log.Error("Unable to close socket - :err", wd.ErrParam(err))
			}
			return
		case <-terminate:
			log.Warning("Terminate")
			if err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				log.Error("Unable to close socket - :err", wd.ErrParam(err))
			}
			return
		}
	}
}
