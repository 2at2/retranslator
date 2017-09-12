package main

import (
	"flag"
	"github.com/2at2/retranslator/receiver"
	"github.com/2at2/retranslator/receiver/module"
	"github.com/mono83/slf/recievers/ansi"
	"github.com/mono83/slf/wd"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

func main() {
	var port int

	flag.IntVar(&port, "port", 8029, "")
	flag.Parse()

	wd.AddReceiver(ansi.New(true, true, false))

	var log = wd.NewLogger("receiver")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Hub
	hub := &receiver.Hub{}
	hub.Init()

	// Client http
	go func() {
		log.Info("Http client on port :port", wd.IntParam("port", port))

		err := http.ListenAndServe(":"+strconv.Itoa(port), module.Handler{Hub: hub})

		if err != nil {
			log.Alert("Client error - :err", wd.ErrParam(err))
		}
	}()

	for {
		select {
		case <-interrupt:
			log.Warning("Interrupt")
			hub.Close()
			return
		}
	}
}
