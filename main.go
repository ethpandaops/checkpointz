package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/samcm/checkpointz/cmd"
)

func main() {
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, syscall.SIGTERM, syscall.SIGINT)

	go cmd.Execute()

	sig := <-cancel
	log.Printf("Caught signal: %v", sig)

	os.Exit(0)
}
