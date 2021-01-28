package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)


func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	counter := int64(0)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	ticker := time.NewTicker(1 * time.Second)

	logger.Info("Start logging app")
	for {
		select {
		case <-ticker.C:
			counter++
			logger.Info("Logging a line", zap.Int64("counter", counter))
			break
		case <-c:
			ticker.Stop()
			logger.Info("Stop logging app")
			return
		}
	}
}
