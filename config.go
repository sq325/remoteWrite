package client

import "time"

type config struct {
	DialTimeout time.Duration
	Timeout     time.Duration
}

func newConfig(opts ...Option) *config {
	c := &config{}

	for _, opt := range opts {
		opt.apply(c)
	}

	return c
}

type Option interface {
	apply(*config)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*config)

func (f optionFunc) apply(c *config) {
	f(c)
}
func WithDialTimeout(d time.Duration) Option {
	return optionFunc(func(c *config) {
		c.DialTimeout = d
	})
}

func WithTimeout(d time.Duration) Option {
	return optionFunc(func(c *config) {
		c.Timeout = d
	})
}
