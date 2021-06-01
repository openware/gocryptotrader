package main

import (
	"encoding/json"
	irixCfg "github.com/openware/irix/config"
	"log"
	"sync"

	"github.com/openware/gocryptotrader/engine"
	exchange "github.com/openware/irix"
)

func main() {
	var err error
	engine.Bot, err = engine.New()
	if err != nil {
		log.Fatalf("Failed to initialise engine. Err: %s", err)
	}

	log.Printf("Loading exchanges..")
	var wg sync.WaitGroup
	for x := range exchange.Exchanges {
		name := exchange.Exchanges[x]
		err = engine.Bot.LoadExchange(name, true, &wg)
		if err != nil {
			log.Printf("Failed to load exchange %s. Err: %s", name, err)
			continue
		}
	}
	wg.Wait()
	log.Println("Done.")

	var cfgs []irixCfg.ExchangeConfig
	exchanges := engine.Bot.GetExchanges()
	for x := range exchanges {
		var cfg *irixCfg.ExchangeConfig
		cfg, err = exchanges[x].GetDefaultConfig()
		if err != nil {
			log.Printf("Failed to get exchanges default config. Err: %s", err)
			continue
		}
		log.Printf("Adding %s", exchanges[x].GetName())
		cfgs = append(cfgs, *cfg)
	}

	data, err := json.MarshalIndent(cfgs, "", " ")
	if err != nil {
		log.Fatalf("Unable to marshal cfgs. Err: %s", err)
	}

	log.Println(string(data))
}
