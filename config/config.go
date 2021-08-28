package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug bool `envconfig:"DEBUG" default:"false"`
}

// NewConfig inits a new config
func NewConfig() Config {
	s := Config{}
	if err := envconfig.Process("UPTODATE", &s); err != nil {
		log.Fatal(err)
	}
	return s
}
