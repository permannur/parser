package config

import (
	"encoding/json"
	"os"
	"time"
)

type IServer interface {
	GetHost() string
	GetPort() string
}

type IConf interface {
	GetServer() IServer
	GetPerUrlTimeout() time.Duration
	GetContextTimeout() time.Duration
	GetDebug() bool
}

type server struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type conf struct {
	Server         *server       `json:"server"`
	PerUrlTimeout  time.Duration `json:"per_url_timeout"`
	ContextTimeout time.Duration `json:"context_timeout"`
	Debug          bool          `json:"debug"`
}

var p *conf

func Init() (err error) {
	var f *os.File
	f, err = os.Open("config.json")
	if err != nil {
		return
	}
	decoder := json.NewDecoder(f)
	p = &conf{}
	err = decoder.Decode(p)
	if err != nil {
		return
	}
	return
}

func (s *server) GetHost() string {
	return s.Host
}

func (s *server) GetPort() string {
	return s.Port
}

func (c *conf) GetServer() IServer {
	return c.Server
}

func (c *conf) GetPerUrlTimeout() time.Duration {
	return c.PerUrlTimeout
}

func (c *conf) GetContextTimeout() time.Duration {
	return c.ContextTimeout
}

func (c *conf) GetDebug() bool {
	return c.Debug
}

func Values() IConf {
	return p
}
