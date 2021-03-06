package services

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Config gathers all the necessary data to run the app
type Config struct {
	URL                   string `json:"url"`
	Port                  string `json:"port"`
	StaticContentPath     string `json:"static_content_path"`
	RedisAddress          string `json:"redis_address"`
	SecretKey             string `json:"secret_key"`
	SessionName           string `json:"session_name"`
	DatabaseUrl           string `json:"database_url"`
	DatabaseName          string `json:"database_name"`
	Debug                 bool   `json:"debug"`
	SecureCookies         bool   `json:"secure_cookies"`
	StorePath             string `json:"store_path"`
	ThumbnailStorePath    string `json:"thumbnail_store_path"`
	WebStorePath          string `json:"web_store_path"`
	WebThumbnailStorePath string `json:"web_thumbnail_store_path"`
	LogsPath              string `json:"logs_path"`
	UseHTTPS              bool   `json:"use_https"`
	SSLCert               string `json:"ssl_cert"`
	SSLKey                string `json:"ssl_key"`
}

// NewConfig creates a new config struct
func NewConfig(configPath string) (*Config, error) {
	var config = new(Config)

	file, err := os.Open(configPath)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
