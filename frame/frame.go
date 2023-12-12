package frame

import (
	"bmp/config"
	"bmp/constants"
	"bmp/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/jfcarter2358/go-logger"
)

type Frame struct {
	Endpoint      string
	Version       string
	Protocol      string
	Auth          string
	Kind          string
	ContentType   string
	ContentLength int
	Data          string
}

func ParseFrame(bytes []byte) (Frame, error) {
	logger.Debugf("", "---- Processing new frame ----")
	f := Frame{}

	data := string(bytes)
	lines := strings.Split(data, "\n")

	logger.Tracef("", "Got %d lines: %v", len(lines), lines)

	for _, line := range lines {
		logger.Tracef("", "Processing line %s", line)
		parts := strings.Split(line, ": ")
		logger.Tracef("", "Got parts %v", parts)
		if len(parts) == 1 {
			if err := f.ParseField("", parts[0]); err != nil {
				return f, err
			}
		} else {
			if err := f.ParseField(parts[0], parts[1]); err != nil {
				return f, err
			}
		}
		logger.Tracef("", "Frame state: %v", f)
	}

	return f, nil
}

func (f *Frame) ParseField(field, value string) error {
	logger.Tracef("", "Parsing field: %s", field)
	switch field {
	case constants.FIELD_VERSION:
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			return fmt.Errorf("value for field %s does not conform to format <version>/<%s>", field, strings.Join(constants.PROTOCOLS, "|"))
		}
		if !utils.Contains(constants.VERSIONS, parts[0]) {
			return fmt.Errorf("invalid version number: %s, valid versions are: %s", value, strings.Join(constants.VERSIONS, ", "))
		}
		if !utils.Contains(constants.PROTOCOLS, parts[1]) {
			return fmt.Errorf("invalid protocol: %s, valid protocols are: %s", value, strings.Join(constants.PROTOCOLS, ", "))
		}
		logger.Tracef("", "Version parse: %s", value)
		f.Version = parts[0]
		f.Protocol = parts[1]
	case constants.FIELD_AUTH:
		logger.Tracef("", "Auth parse: %s", value)
		f.Auth = value
	case constants.FIELD_KIND:
		if !utils.Contains(constants.KINDS, value) {
			return fmt.Errorf("invalid kind: %s, valid kinds are: %s", value, strings.Join(constants.KINDS, ", "))
		}
		logger.Tracef("", "Kind parse: %s", value)
		f.Kind = value
	case constants.FIELD_CONTENT_TYPE:
		if !utils.Contains(constants.CONTENT_TYPES, value) {
			return fmt.Errorf("invalid content type: %s, valid types are: %s", value, strings.Join(constants.CONTENT_TYPES, ", "))
		}
		logger.Tracef("", "Content type parse: %s", value)
		f.ContentType = value
	case constants.FIELD_CONTENT_LENGTH:
		val, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid content length: %s, must be integer", value)
		}
		logger.Tracef("", "Content length parse: %d", val)
		f.ContentLength = val
	case constants.FIELD_ENDPOINT:
		logger.Tracef("", "Endpoint parse: %s", value)
		f.Endpoint = value
	default:
		logger.Tracef("", "Data parse: %s", value)
		if len(f.Data) < f.ContentLength {
			f.Data += value
		}
		if len(f.Data) > f.ContentLength {
			f.Data = f.Data[:f.ContentLength]
		}
	}
	return nil
}

func (f *Frame) Do(c *websocket.Conn, mt int, conf *config.Config, pervious Frame) error {
	switch f.Kind {
	case constants.KIND_ACCEPT:

	case constants.KIND_ACKNOWLEDGE:

	case constants.KIND_CLOSE:
		switch conf.Type {
		case constants.TYPE_SERVER:
			out := NewClose(*conf)
			err := out.WriteFrame(c, mt, *conf)
			return err
		}
	case constants.KIND_CONNECT:
		switch conf.Type {
		case constants.TYPE_CLIENT:
			out := NewAcknowledge(*conf)
			err := out.WriteFrame(c, mt, *conf)
			return err
		case constants.TYPE_SERVER:
			conf.Version = fmt.Sprintf("%s/%s", f.Version, f.Protocol)
			out := NewConnect(*conf)
			err := out.WriteFrame(c, mt, *conf)
			return err
		}
	case constants.KIND_DATA:
		switch conf.Type {
		case constants.TYPE_CLIENT:
			parts := strings.Split(f.Endpoint, "/")
			switch parts[0] {
			case constants.ENDPOINT_TYPE_BUFFER:
				if err := conf.BufferPush(parts[1], f.Data); err != nil {
					return err
				}
			case constants.ENDPOINT_TYPE_QUEUE:
				if err := conf.QueuePush(parts[1], f.Data); err != nil {
					return err
				}
			}
			out := NewAcknowledge(*conf)
			err := out.WriteFrame(c, mt, *conf)
			return err
		}
	case constants.KIND_ERROR:
		logger.Errorf("", "Got frame error: %s", f.Data)
	case constants.KIND_PULL:
		switch conf.Type {
		case constants.TYPE_SERVER:
			parts := strings.Split(f.Endpoint, "/")
			switch parts[0] {
			case constants.ENDPOINT_TYPE_BUFFER:
				contentType, content, err := conf.BufferPull(parts[1])
				if err != nil {
					return err
				}
				out := NewData(*conf, content, contentType)
				return out.WriteFrame(c, mt, *conf)
			case constants.ENDPOINT_TYPE_QUEUE:
				contentType, content, err := conf.QueuePull(parts[1])
				if err != nil {
					return err
				}
				out := NewData(*conf, content, contentType)
				return out.WriteFrame(c, mt, *conf)
			}
		}
	case constants.KIND_PUSH:
		switch conf.Type {
		case constants.TYPE_SERVER:
			parts := strings.Split(f.Endpoint, "/")
			switch parts[0] {
			case constants.ENDPOINT_TYPE_BUFFER:
				if err := conf.BufferPush(parts[1], f.Data); err != nil {
					out := NewError(*conf, err)
					err := out.WriteFrame(c, mt, *conf)
					return err
				}
				out := NewAcknowledge(*conf)
				err := out.WriteFrame(c, mt, *conf)
				return err
			case constants.ENDPOINT_TYPE_QUEUE:
				if err := conf.QueuePush(parts[1], f.Data); err != nil {
					out := NewError(*conf, err)
					err := out.WriteFrame(c, mt, *conf)
					return err
				}
				out := NewAcknowledge(*conf)
				err := out.WriteFrame(c, mt, *conf)
				return err
			}
		}
	case constants.KIND_REJECT:

	case constants.KIND_RETRY:
	}
	return nil
	// return fmt.Errorf("invalid frame kind: %s, valid kinds are %v", f.Kind, constants.KINDS)
}

func (f *Frame) WriteFrame(c *websocket.Conn, mt int, conf config.Config) error {
	message := ""
	message += fmt.Sprintf("%s: %s\n", constants.FIELD_VERSION, conf.Version)
	message += fmt.Sprintf("%s: %s\n", constants.FIELD_AUTH, conf.Auth)
	message += fmt.Sprintf("%s: %s\n", constants.FIELD_KIND, f.Kind)
	if f.Endpoint != "" {
		message += fmt.Sprintf("%s: %s\n", constants.FIELD_ENDPOINT, f.Endpoint)
	}
	if f.Data != "" {
		message += fmt.Sprintf("%s: %s\n", constants.FIELD_CONTENT_TYPE, f.ContentType)
		message += fmt.Sprintf("%s: %d\n", constants.FIELD_CONTENT_LENGTH, f.ContentLength)
		message += f.Data
	}
	err := c.WriteMessage(mt, []byte(message))
	return err
}

func NewAccept(conf config.Config) Frame {
	f := Frame{
		Version: conf.Version,
		Auth:    conf.Auth,
		Kind:    constants.KIND_ACCEPT,
	}

	return f
}

func NewAcknowledge(conf config.Config) Frame {
	f := Frame{
		Version: conf.Version,
		Auth:    conf.Auth,
		Kind:    constants.KIND_ACKNOWLEDGE,
	}

	return f
}

func NewClose(conf config.Config) Frame {
	f := Frame{
		Version: conf.Version,
		Auth:    conf.Auth,
		Kind:    constants.KIND_CLOSE,
	}

	return f
}

func NewConnect(conf config.Config) Frame {
	f := Frame{
		Version: conf.Version,
		Auth:    conf.Auth,
		Kind:    constants.KIND_CONNECT,
	}

	return f
}

func NewData(conf config.Config, data, contentType string) Frame {
	f := Frame{
		Version:       conf.Version,
		Auth:          conf.Auth,
		Kind:          constants.KIND_DATA,
		ContentType:   contentType,
		ContentLength: len(data),
		Data:          data,
	}

	return f
}

func NewError(conf config.Config, err error) Frame {
	errS := err.Error()
	f := Frame{
		Version:       conf.Version,
		Auth:          conf.Auth,
		Kind:          constants.KIND_ERROR,
		ContentType:   constants.CONTENT_TYPE_TEXT,
		ContentLength: len(errS),
		Data:          errS,
	}

	return f
}

func NewPull(conf config.Config, endpoint string) Frame {
	f := Frame{
		Version:  conf.Version,
		Auth:     conf.Auth,
		Kind:     constants.KIND_PULL,
		Endpoint: endpoint,
	}

	return f
}

func NewPush(conf config.Config, endpoint, data, contentType string) Frame {
	f := Frame{
		Version:       conf.Version,
		Auth:          conf.Auth,
		Kind:          constants.KIND_PUSH,
		Endpoint:      endpoint,
		ContentType:   contentType,
		ContentLength: len(data),
		Data:          data,
	}

	return f
}

func NewReject(conf config.Config) Frame {
	f := Frame{
		Version: conf.Version,
		Auth:    conf.Auth,
		Kind:    constants.KIND_REJECT,
	}

	return f
}

func NewRetry(conf config.Config) Frame {
	f := Frame{
		Version: conf.Version,
		Auth:    conf.Auth,
		Kind:    constants.KIND_RETRY,
	}

	return f
}
