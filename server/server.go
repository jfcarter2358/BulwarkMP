package server

import (
	"bmp/config"
	"bmp/constants"
	"bmp/frame"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jfcarter2358/go-logger"
)

var upgrader = websocket.Upgrader{} // use default options

var BufferPull func(string) (string, string, error)
var QueuePull func(string) (string, string, error)
var BufferPush func(string, string) error
var QueuePush func(string, string) error
var Version string

func connect(w http.ResponseWriter, r *http.Request) {
	conf := config.Config{
		Type:       constants.TYPE_SERVER,
		Version:    Version,
		BufferPull: BufferPull,
		BufferPush: BufferPush,
		QueuePull:  QueuePull,
		QueuePush:  QueuePush,
	}
	previous := frame.Frame{}
	mt := websocket.TextMessage

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("", "Error on upgrading to websocket connection: %s", err.Error())
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logger.Error("", "Unable to read incoming message")
			break
		}
		logger.Tracef("", "Received inbound message: %s", message)
		f, err := frame.ParseFrame(message)
		if err != nil {
			logger.Errorf("", "Unable to parse incoming frame")
			fOut := frame.NewError(conf, err)
			if err := fOut.WriteFrame(c, mt, conf); err != nil {
				logger.Errorf("", "Unable to reach client with error %s", err.Error())
				return
			}
		}
		logger.Debugf("", "Processing frame %v", f)
		if err := f.Do(c, mt, &conf, previous); err != nil {
			logger.Fatalf("", "Unable to reach client with error %s", err.Error())
		}
		if f.Kind == constants.KIND_CLOSE {
			logger.Infof("", "Closing connection")
			return
		}
	}
}

func Start(logLevel, logFormat, addr string) {
	logger.SetFormat(logFormat)
	logger.SetLevel(logLevel)
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc(constants.ENDPOINT, connect)
	logger.Infof("", "Starting Bulwark server on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
