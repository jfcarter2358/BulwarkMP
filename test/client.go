//go:build ignore
// +build ignore

package main

import (
	"bmp/client"
	"bmp/constants.go"
	"fmt"
	"time"

	"github.com/jfcarter2358/go-logger"
)

func bufferPush(endpoint, data string) error {
	logger.Infof("", "Got data from buffer %s", endpoint)
	logger.Debugf("", "Data: %s", data)
	return nil
}

func queuePush(endpoint, data string) error {
	logger.Infof("", "Got data from queue %s", endpoint)
	logger.Debugf("", "Data: %s", data)
	return nil
}

func main() {
	c := client.Client{}
	c.New(fmt.Sprintf("%s/%s", constants.VERSION_1, constants.PROTOCOL_PLAIN), "queue/foobar", bufferPush, queuePush)
	go c.Start(logger.LOG_LEVEL_DEBUG, logger.LOG_FORMAT_CONSOLE, "localhost:8081")

	logger.Infof("", "Sleeping for 2...")
	time.Sleep(2 * time.Second)

	logger.Debugf("", "Pushing data")
	c.Push(constants.CONTENT_TYPE_TEXT, "This is some test data")

	logger.Infof("", "Sleeping for 2...")
	time.Sleep(2 * time.Second)

	logger.Debugf("", "Pulling data")
	c.Pull()

	logger.Infof("", "Sleeping for 2...")
	time.Sleep(2 * time.Second)

	logger.Debugf("", "Closing connection")
	c.Close()
}
