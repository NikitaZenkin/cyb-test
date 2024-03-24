package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DSN            string
	MigrationsPath string
	HTTPServer     *HTTPServer
	DSNServerAddr  string
}

type HTTPServer struct {
	Host     string
	Port     string
	BasePath string
}

func Configure() (*Config, error) {
	r := &envReader{}

	result := &Config{
		DSN:            r.readEnv("DSN"),
		MigrationsPath: r.readEnv("MIGRATION_PATH"),
		HTTPServer: &HTTPServer{
			Host:     r.readEnv("SERV_HOST"),
			Port:     r.readEnv("SERV_PORT"),
			BasePath: r.readEnv("BASE_PATH"),
		},
		DSNServerAddr: r.readEnv("DNS_SERV_ADDR"),
	}

	if len(r.emptyParams) > 0 {
		return nil, fmt.Errorf("config parameters not set \n%s", strings.Join(r.emptyParams, "\n"))
	}

	return result, nil
}

type envReader struct {
	emptyParams []string
}

func (r *envReader) readEnv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		r.emptyParams = append(r.emptyParams, name)
	}

	return val
}
