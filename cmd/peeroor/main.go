package main

import (
	"Peeroor/pkg"
	"github.com/spf13/pflag"
	"log"
	"os"
	"sync"
)

type CliArgs struct {
	config string
}

func main() {
	cliArgs := CliArgs{}
	flags := pflag.NewFlagSet("main", pflag.ContinueOnError)
	flags.StringVar(&cliArgs.config, "config", "", "path to config file")
	flags.Parse(os.Args)

	// Load configuration from file.
	config, err := pkg.LoadConfig(cliArgs.config)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Create a wait group to manage network goroutines.
	var wg sync.WaitGroup

	// For each network defined in the config, create a Network and start maintenance.
	for netName, rpcKeys := range config.Networks {
		network := pkg.NewNetwork(netName, rpcKeys, config)
		wg.Add(1)
		go func(nw *pkg.Network) {
			defer wg.Done()
			nw.Maintain()
		}(network)
	}

	// Wait indefinitely.
	wg.Wait()
}
