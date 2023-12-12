//go:build ignore
// +build ignore

package main

import (
	"bmp/constants"
	"bmp/server"
	"fmt"

	"github.com/jfcarter2358/go-logger"
)

func BufferPull(endpoint string) (string, string, error) {
	logger.Infof("", "Grabbing data from buffer %s", endpoint)
	return constants.CONTENT_TYPE_TEXT, "buffer data", nil
}

func QueuePull(endpoint string) (string, string, error) {
	logger.Infof("", "Grabbing data from queue %s", endpoint)
	return constants.CONTENT_TYPE_TEXT, "queue data", nil
}

func BufferPush(endpoint, data string) error {
	logger.Infof("", "Pushing data to buffer %s", endpoint)
	logger.Debugf("", "Data: %s", data)
	return nil
}

func QueuePush(endpoint, data string) error {
	logger.Infof("", "Pushing data to queue %s", endpoint)
	logger.Debugf("", "Data: %s", data)
	return nil
}

func main() {
	server.BufferPull = BufferPull
	server.BufferPush = BufferPush
	server.QueuePull = QueuePull
	server.QueuePush = QueuePush
	server.Version = fmt.Sprintf("%s/%s", constants.VERSION_1, constants.PROTOCOL_PLAIN)
	server.Start(logger.LOG_LEVEL_DEBUG, logger.LOG_FORMAT_CONSOLE, "localhost:8081")
}
