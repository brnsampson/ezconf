package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

const (
	DEFAULT_FLAG_CONFIG = "conf.toml"
)

var (
	help bool
	h    bool
)

func main() {
	// See config.go for the configuration specific flags being defined.
	setupFlags()

	flag.BoolVar(&help, "help", false, "Get usage message")
	flag.BoolVar(&h, "h", false, "Get usage message")
	flag.Parse()

	if help || h {
		flag.Usage()
		os.Exit(0)
	}

	// Set up logging in debug so that we can see if anything goes wrong while loading
	programLevel := new(slog.LevelVar)
	programLevel.Set(slog.LevelDebug)
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	logger := slog.New(h)
	slog.SetDefault(logger)

	loader := NewLoader()
	slog.Info("Initialized loader", "loader", loader)
	slog.Info("Initialized sub-loader", "loader", loader.SubLoader)

	err := loader.Update()
	if err != nil {
		slog.Error("Error loading config", "error", err)
		os.Exit(2)
	} else {
		slog.Info("Loading completed")
	}
	config := loader.Current()

	// Reset log level to user defined value
	programLevel.Set(config.LogLevel)

	slog.Debug("Loaded sub-config", "config", config.SubConfig)
	slog.Debug("Loaded config", "config", config)

	// We have a reloadable config! Easy, right?!?! ...right?
	fmt.Println("")
	fmt.Println("I loaded a config and my host is: ", config.SubConfig.Host)
	fmt.Println("")
	fmt.Println("My address is: ", config.SubConfig.GetAddr())
	fmt.Println("")
	fmt.Println("My logging level was set to: ", config.LogLevel)
	fmt.Println("")
	fmt.Println("My dev mode was enabled: ", config.DevMode)
	fmt.Println("")
	fmt.Println("Any my config file was set to: ", config.ConfigFile)
	fmt.Println("")
	fmt.Println("Plus, I have a TLS config, but you probably don't want me to print out a long byte string.")
}
