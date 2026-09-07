package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ListenAddr    string
	DataDir       string
	SecureCookies bool
}

func FromEnv() (Config, error) {
	c := Config{
		ListenAddr: envOr("GALLERY_LISTEN_ADDR", ":8080"),
		DataDir:    envOr("GALLERY_DATA_DIR", "./data"),
	}
	if strings.TrimSpace(c.ListenAddr) == "" {
		return Config{}, fmt.Errorf("GALLERY_LISTEN_ADDR must not be empty")
	}
	if strings.TrimSpace(c.DataDir) == "" {
		return Config{}, fmt.Errorf("GALLERY_DATA_DIR must not be empty")
	}
	if raw, ok := os.LookupEnv("GALLERY_SECURE_COOKIES"); ok {
		secure, err := strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("GALLERY_SECURE_COOKIES must be true or false")
		}
		c.SecureCookies = secure
	}
	return c, nil
}

func envOr(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return fallback
}
