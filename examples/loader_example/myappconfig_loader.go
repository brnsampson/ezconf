package main

import (
	"flag"
	"fmt"
	"github.com/brnsampson/ezconf/file"
	"github.com/brnsampson/ezconf/httpconf"
	"github.com/brnsampson/optional"
	"sync"
)

// Default values for MyServiceConfig
const (
	DefaultMyServiceConfigDescription   = ""
	DefaultMyServiceConfigNodeID        = 1
	DefaultMyServiceConfigPriority      = 1
	DefaultMyServiceConfigSecretKey     = "/etc/myapp/secretkey.txt"
	DefaultMyServiceConfigAddress       = "127.0.0.1"
	DefaultMyServiceConfigPort          = 443
	DefaultMyServiceConfigTlsCert       = "/etc/myapp/tls/cert.pem"
	DefaultMyServiceConfigTlsPrivateKey = "/etc/myapp/tls/key.pem"
	DefaultMyServiceConfigTlsEnabled    = true
	DefaultMyServiceConfigTlsSkipVerify = false
)

// Default values for MyDBConfig
const (
	DefaultMyDBConfigAddress = "127.0.0.1"
	DefaultMyDBConfigPort    = 8080
)

// flag variables
var (
	flagSetupper      sync.Once
	myServiceNodeFlag optional.Uint32
	myDBAddressFlag   optional.Str
	myDBPortFlag      optional.Uint16
)

type loader[T any] interface {
	Update() (T, error)
	Prev() T
}

// SetupMyAppConfigFlags adds command line flags to support this config. It is suggested that
// you just call NewLoader() which does this for you, but you may do this yourself if you want more
// control. As an example:
//
// SetupMyAppConfigFlags()
// l := MyAppConfigLoader{}
// myappconf, err := l.Update(flag.CommandLine)
func SetupMyAppConfigFlags() {
	onceBody := func() {
		flag.Var(&myServiceNodeFlag, "myServiceNode", "MyServiceConfig Node Value. Type: uint32, Required: true")
		flag.Var(&myDBAddressFlag, "myDBAddress", "MyDBConfig Address Value. Type: String, Default: '127.0.0.1'")
		flag.Var(&myDBPortFlag, "myDBPort", "MyDBConfig Port Value. Type: uint16, Default: 8080")
	}
	flagSetupper.Do(onceBody)
}

// NewLoader sets up required flags, creates a new loader, updates it, and returns the loaded loader.
func NewLoader() (MyAppConfigLoader, error) {
	SetupMyAppConfigFlags()

	l := MyAppConfigLoader{}
	_, err := l.Update()
	return l, err
}

type MyAppConfigLoader struct {
	MyService MyServiceConfigLoader
	MyDB      MyDBConfigLoader
	Flags     *flag.FlagSet
	previous  MyAppConfig
}

func (l *MyAppConfigLoader) Prev() MyAppConfig {
	return l.previous
}

func (l *MyAppConfigLoader) Update() (config MyAppConfig, err error) {
	// TODO: check myAppConfigPath for the value of the -config flag and use that as the config file to load.
	// TODO: load l from env vars.
	myService, err := l.MyService.Update()
	if err != nil {
		return
	}

	myDB, err := l.MyDB.Update()
	if err != nil {
		return
	}

	l.previous = MyAppConfig{MyService: myService, MyDB: myDB}

	return l.previous, err
}

// Loader for MyServiceConfig type
type MyServiceConfigLoader struct {
	Name         optional.Str    `env:"MY_APP_MY_SERVICE_NAME"`
	Description  optional.Str    `env:"MY_APP_MY_SERVICE_DESCRIPTION"`
	NodeID       optional.Uint32 `json:"node" toml:"node" yaml:"node" env:"MY_APP_MY_SERVICE_NODE"`
	Priority     optional.Uint16 `env:"MY_APP_MY_SERVICE_PRIORITY"`
	SecretKey    file.SecretFile `env:"MY_APP_MY_SERVICE_SECRET_KEY"`
	ServerConfig httpconf.HttpServerLoader
	previous     MyServiceConfig
}

func (l *MyServiceConfigLoader) Update() (c MyServiceConfig, err error) {
	var ok bool
	// Flags are defined as package variables above. Flags override all other config sources.
	l.NodeID = optional.Or(myServiceNodeFlag, l.NodeID)

	// Update sub-loaders
	if l.SecretKey.IsNone() {
		l.SecretKey.Set(DefaultMyServiceConfigSecretKey)
	}

	// Read values from file types
	secretKey, ok := l.SecretKey.ReadFile()
	if !ok {
		return c, fmt.Errorf("MyServiceConfig missing required field: SecretKey")
	}

	serverConfig, err := l.ServerConfig.Update()

	newConfig := l.previous
	newConfig.Name, ok = l.Name.Get()
	if !ok {
		return c, fmt.Errorf("MyServiceConfig missing required field: Name")
	}
	newConfig.Description = optional.GetOr(l.Description, DefaultMyServiceConfigDescription)
	newConfig.NodeID = optional.GetOr(l.NodeID, DefaultMyServiceConfigNodeID)
	newConfig.Priority = optional.GetOr(l.Priority, DefaultMyServiceConfigPriority)
	newConfig.SecretKey = secretKey
	newConfig.ServerConfig = serverConfig

	l.previous = newConfig
	return newConfig, nil
}

func (l MyServiceConfigLoader) Prev() MyServiceConfig {
	return l.previous
}

// Loader for MyDBConfig type
type MyDBConfigLoader struct {
	Address  optional.Str    `env:"MY_APP_MY_DB_ADDRESS"`
	Port     optional.Uint16 `env:"MY_APP_MY_DB_PORT"`
	previous MyDBConfig
}

func (l *MyDBConfigLoader) Update() (c MyDBConfig, err error) {
	// Flags are defined as package variables above. Flags override all other config sources.
	l.Address = optional.Or(myDBAddressFlag, l.Address)
	l.Port = optional.Or(myDBPortFlag, l.Port)

	newConfig := l.previous

	newConfig.Address = optional.GetOr(l.Address, DefaultMyDBConfigAddress)
	newConfig.Port = optional.GetOr(l.Port, DefaultMyDBConfigPort)

	l.previous = newConfig
	return l.previous, nil
}

func (l MyDBConfigLoader) Prev() MyDBConfig {
	return l.previous
}
