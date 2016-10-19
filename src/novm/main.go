package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rocksolidlabs/novm/src/novm/cmd"
	"github.com/rocksolidlabs/novm/src/novm/utils"
)

func main() {

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		utils.Cleanup()
		os.Exit(0)
	}()

	cmd.Execute()
}
