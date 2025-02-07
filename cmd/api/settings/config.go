package settings

import (
	"flag"
	"os"
	"telegram/cmd/api/internal"
)

type Config struct {
	Address string
}

func NewConfig() *Config {
	c := &Config{Address: internal.DefaultServerAddr}

	return c
}

func (c *Config) WithFlag() {
	flag.StringVar(&c.Address, "a", internal.DefaultServerAddr, "server address")
	flag.Parse()
}

func (c *Config) WithEnv() {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.Address = envRunAddr
	}
}
