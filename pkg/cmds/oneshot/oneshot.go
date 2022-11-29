package oneshot

import (
	"os"

	"github.com/rs/zerolog"
)

type OneshotOptions func(config *OneshotConfig)

type OneshotConfig struct {
	log *zerolog.Logger
}

func NewOneshot(opts ...OneshotOptions) (*OneshotConfig, error) {
	cfg := &OneshotConfig{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.log == nil {
		log := zerolog.New(os.Stderr).With().Timestamp().Logger()
		cfg.log = &log
	}

	return cfg, nil
}

func WithLogger(log *zerolog.Logger) OneshotOptions {
	return func(config *OneshotConfig) {
		config.log = log
	}
}

func (c *OneshotConfig) Run() error {
	c.log.Info().Msg("Running oneshot")
	return nil
}
