package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Moranilt/config-keeper/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		<-exit
		cancel()
	}()

	server.Run(ctx)
}
