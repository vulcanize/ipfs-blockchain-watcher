package main

import (
	"flag"
	"log"

	"fmt"

	"github.com/8thlight/vulcanizedb/pkg/config"
	"github.com/8thlight/vulcanizedb/pkg/geth"
	"github.com/8thlight/vulcanizedb/pkg/history"
	"github.com/8thlight/vulcanizedb/pkg/repositories"
)

func main() {
	environment := flag.String("environment", "", "Environment name")
	startingBlockNumber := flag.Int("starting-number", -1, "First block to fill from")
	flag.Parse()
	cfg, err := config.NewConfig(*environment)
	if err != nil {
		log.Fatalf("Error loading config\n%v", err)
	}

	blockchain := geth.NewGethBlockchain(cfg.Client.IPCPath)
	repository := repositories.NewPostgres(cfg.Database)
	numberOfBlocksCreated := history.PopulateBlocks(blockchain, repository, int64(*startingBlockNumber))
	fmt.Printf("Populated %d blocks", numberOfBlocksCreated)
}
