package client

import (
	"net/url"

	"bmp/config"
	"bmp/constants"
	"bmp/frame"

	"github.com/gorilla/websocket"
	"github.com/jfcarter2358/go-logger"
)

type Client struct {
	Endpoint string
	Conf     config.Config
	Previous frame.Frame
	Conn     *websocket.Conn
}

func (cl *Client) Run(addr string) {

	mt := websocket.TextMessage

	u := url.URL{Scheme: "ws", Host: addr, Path: constants.ENDPOINT}
	logger.Infof("", "Connecting to %s", u.String())

	var err error
	cl.Conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Errorf("", "Error on dial: %s", err.Error())
		return
	}
	defer cl.Conn.Close()

	fInit := frame.NewConnect(cl.Conf)
	if err := fInit.WriteFrame(cl.Conn, mt, cl.Conf); err != nil {
		logger.Errorf("", "Error on connect: %s", err.Error())
		return
	}

	done := make(chan struct{})

	defer close(done)
	for {
		_, message, err := cl.Conn.ReadMessage()
		if err != nil {
			logger.Error("", "Unable to read incoming message")
			break
		}
		logger.Tracef("", "Received inbound message: %s", message)
		f, err := frame.ParseFrame(message)
		if err != nil {
			logger.Errorf("", "Unable to parse incoming frame")
			fOut := frame.NewError(cl.Conf, err)
			if err := fOut.WriteFrame(cl.Conn, mt, cl.Conf); err != nil {
				logger.Errorf("", "Unable to reach client with error %s", err.Error())
				return
			}
		}
		logger.Debugf("", "Processing frame %v", f)
		if err := f.Do(cl.Conn, mt, &cl.Conf, cl.Previous); err != nil {
			logger.Fatalf("", "Unable to reach client with error %s", err.Error())
		}
	}
}

func (cl *Client) Push(contentType string, data interface{}) error {
	// switch contentType {
	// case constants.CONTENT_TYPE_BINARY:

	// case constants.CONTENT_TYPE_JSON:

	// case constants.CONTENT_TYPE_TEXT:

	// case constants.CONTENT_TYPE_YAML:

	// case constants.CONTENT_TYPE_XML:

	// }
	strData := data.(string)
	f := frame.NewPush(cl.Conf, cl.Endpoint, strData, contentType)
	return f.WriteFrame(cl.Conn, websocket.TextMessage, cl.Conf)
}

func (cl *Client) Pull() error {
	f := frame.NewPull(cl.Conf, cl.Endpoint)
	return f.WriteFrame(cl.Conn, websocket.TextMessage, cl.Conf)
}

func (cl *Client) Close() error {
	f := frame.NewClose(cl.Conf)
	return f.WriteFrame(cl.Conn, websocket.TextMessage, cl.Conf)
}

func (cl *Client) Start(logLevel, logFormat, addr string) {
	logger.SetFormat(logFormat)
	logger.SetLevel(logLevel)

	cl.Run(addr)
}

func (cl *Client) New(version, endpoint string, bufferFunc, queueFunc func(string, string) error) {
	cl.Endpoint = endpoint
	cl.Conf = config.Config{
		Version:    version,
		Auth:       "abdefghijklmnopqrstuvwxyz1234567890",
		BufferPush: bufferFunc,
		QueuePush:  queueFunc,
		Type:       constants.TYPE_CLIENT,
	}
	cl.Previous = frame.Frame{}
}
