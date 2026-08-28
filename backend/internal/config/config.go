// Package config reads runtime settings from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	AllowedOrigins []string
	CookieName     string
	CookieSecure   bool
	ShutdownGrace  time.Duration

	// Optional: without these the guided tour narrates using the browser's
	// own speech synthesis instead.
	ElevenLabsKey   string
	ElevenLabsVoice string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:          env("PORT", "8080"),
		DatabaseURL:   env("DATABASE_URL", ""),
		CookieName:    env("SESSION_COOKIE", "brag_session"),
		CookieSecure:  envBool("COOKIE_SECURE", false),
		ShutdownGrace: 15 * time.Second,

		ElevenLabsKey:   env("ELEVENLABS_API_KEY", ""),
		ElevenLabsVoice: env("ELEVENLABS_VOICE_ID", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// The frontend is a separate origin in development, so credentials-carrying
	// requests need it listed explicitly — "*" is not allowed with cookies.
	origins := env("ALLOWED_ORIGINS", "http://localhost:5173")
	for _, o := range strings.Split(origins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			cfg.AllowedOrigins = append(cfg.AllowedOrigins, o)
		}
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v, err := strconv.ParseBool(env(key, ""))
	if err != nil {
		return fallback
	}
	return v
}
