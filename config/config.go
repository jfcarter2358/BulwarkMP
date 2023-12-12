package config

type Config struct {
	Type       string
	Version    string
	Auth       string
	Secret     string
	PubKey     string
	BufferPull func(string) (string, string, error)
	QueuePull  func(string) (string, string, error)
	BufferPush func(string, string) error
	QueuePush  func(string, string) error
}
