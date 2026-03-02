package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"btcaggregation/internal/config"
	"btcaggregation/internal/provider"
	"btcaggregation/internal/repository"
	"btcaggregation/internal/server"
	"btcaggregation/internal/service/fetcher"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	if err := config.Init(config.DotEnv); err != nil {
		logger.Fatalf("Config error: %v", err)
	}

	logger.WithFields(logrus.Fields{
		"pair":           config.Pair(),
		"fetch_interval": config.FetchInterval(),
		"port":           config.Port(),
	}).Info("btcaggregation: starting")

	httpClient := &http.Client{Timeout: config.RequestTimeout()}
	providers := []provider.PriceProvider{
		provider.NewCoinbase(httpClient, config.CoinbaseURL()),
		provider.NewKraken(httpClient, config.KrakenURL()),
		provider.NewCoinDesk(httpClient, config.CoindeskURL()),
	}

	priceRepo := repository.New()

	priceFetcher := fetcher.New(providers, priceRepo, logger)
	echoServer := server.New(priceRepo, logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go priceFetcher.Run(ctx)

	if err := server.Start(ctx, echoServer, ":"+config.Port(), logger); err != nil {
		logger.WithField("error", err).Error("btcaggregation: server error")
		os.Exit(1)
	}

	logger.Info("btcaggregation: shutdown complete")
}
