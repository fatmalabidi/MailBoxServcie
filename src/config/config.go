package config

import (
	"github.com/jinzhu/configor"
	"os"
	"path"
	"regexp"
	"runtime"
	"time"
)

type Server struct {
	Type     string
	Host     string        `default:"localhost" env:"SERVER_HOST"`
	Port     string        `default:"50060" env:"SERVER_PORT"`
	Deadline time.Duration `default:"5" env:"GRPC_DEADLINE"`
}

type Pagination struct {
	PageSize       uint32
	MaxResultCount uint32
}

type Messages struct {
	Type       string
	TableName  string
	Pagination Pagination
}

type Database struct {
	Messages Messages
}

type Cognito struct {
	UserPoolID string
	ClientID   string
	Region     string
}

type Storage struct {
	Type            string
	Bucket          string
	CredentialsPath string
}

type Config struct {
	Tag        string // indicates the config environment prod or dev
	Server     Server
	Database   Database
	Cognito    Cognito
	Storage    Storage
}

// LoadConfig sets the application config
// uses CONFIGOR_ENV to set environment,
// if CONFIGOR_ENV not set, environment will be production by default
// and it will be test when running tests with go test
// otherwise it can be set to test manually
func LoadConfig() (*Config, error) {
	var configFilePath string
	config := configor.New(&configor.Config{})
	switch getEnvironment() {
	case "test":
		configFilePath = "../../config/config.test.yml"
	case "development":
		configFilePath = "../../config/config.dev.yml"
	default:
		configFilePath = "../../config/config.prod.yml"
	}
	_, filename, _, _ := runtime.Caller(0)
	filepath := path.Join(path.Dir(filename), configFilePath)
	conf := new(Config)
	err := config.Load(conf, filepath)
	return conf, err
}

var testRegexp = regexp.MustCompile("_test|(\\.test$)")

// getEnvironment returns the environment to run on
func getEnvironment() string {
	if env := os.Getenv("CONFIGOR_ENV"); env != "" {
		return env
	}

	if testRegexp.MatchString(os.Args[0]) {
		return "test"
	}

	return "test"
}
