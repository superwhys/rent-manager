package config

import (
	"errors"
	"time"
)

type JwtConfig struct {
	JwtSecret         string
	JwtTimeout        time.Duration
	JwtRefreshTimeout time.Duration
}

func (c *JwtConfig) Validate() error {
	if c.JwtSecret == "" {
		return errors.New("jwt secret is required")
	}
	if c.JwtTimeout <= 0 {
		return errors.New("jwt timeout is required")
	}
	if c.JwtRefreshTimeout <= 0 {
		c.JwtRefreshTimeout = c.JwtTimeout * 24
	}
	return nil
}
