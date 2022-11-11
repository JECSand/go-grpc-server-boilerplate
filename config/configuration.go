package config

import (
	"encoding/json"
	"os"
	"time"
)

// ServerConfig holds config settings for Server connection
type ServerConfig struct {
	Port              string
	Registration      string
	SSL               string
	Timeout           time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	MaxConnectionIdle time.Duration
	MaxConnectionAge  time.Duration
}

// MongoDBConfig holds config settings for Mongo connection
type MongoDBConfig struct {
	URI string
	DB  string
}

// LoggerConfig holds config settings for server logger
type LoggerConfig struct {
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
}

// Configuration is a struct designed to hold the applications variable configuration settings
type Configuration struct {
	Server       ServerConfig
	MongoDB      MongoDBConfig
	Logger       LoggerConfig
	TokenSecret  string
	RootAdmin    string
	RootPassword string
	RootEmail    string
	RootGroup    string
	Cert         string
	Key          string
	ENV          string
}

// GetConfigurations is a function that reads a json configuration file and outputs a Configuration struct
func GetConfigurations() (*Configuration, error) {
	confFile := "conf.json"
	if os.Getenv("ENV") == "test" {
		confFile = "test_conf.json"
	}
	file, err := os.Open(confFile)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	configurationSettings := Configuration{}
	err = decoder.Decode(&configurationSettings)
	if err != nil {
		return nil, err
	}
	return &configurationSettings, nil
}

// InitializeEnvironmentalVars initializes the environmental variables for the application
func (c *Configuration) InitializeEnvironmentalVars() {
	os.Setenv("MONGO_URI", c.MongoDB.URI)
	os.Setenv("DATABASE", c.MongoDB.DB)
	os.Setenv("TOKEN_SECRET", c.TokenSecret)
	os.Setenv("ROOT_ADMIN", c.RootAdmin)
	os.Setenv("ROOT_PASSWORD", c.RootPassword)
	os.Setenv("ROOT_EMAIL", c.RootEmail)
	os.Setenv("ROOT_GROUP", c.RootGroup)
	os.Setenv("REGISTRATION", c.Server.Registration)
	os.Setenv("PORT", c.Server.Port)
	os.Setenv("HTTPS", c.Server.SSL)
	os.Setenv("CERT", c.Cert)
	os.Setenv("KEY", c.Key)
	os.Setenv("ENV", c.ENV)
}
