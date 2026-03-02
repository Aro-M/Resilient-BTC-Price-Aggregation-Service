package fetcher

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"btcaggregation/internal/config"
	"btcaggregation/internal/connection"
	"btcaggregation/internal/metrics"
	"btcaggregation/internal/provider"
	"btcaggregation/internal/repository"
	"btcaggregation/internal/retry"
	"btcaggregation/internal/service/aggregator"
)

const (
	maxConnectionFailures = 3
	connectionResetTime   = 60 * time.Second
)

type fetchResult struct {
	source string
	price  float64
	err    error
}

type Fetcher struct {
	providers []provider.PriceProvider
	states    map[string]*connection.State
	store     repository.PriceRepository
	logger    *logrus.Logger
}

func New(providers []provider.PriceProvider, st repository.PriceRepository, logger *logrus.Logger) *Fetcher {
	states := make(map[string]*connection.State, len(providers))
	for _, p := range providers {
		states[p.Name()] = connection.New(maxConnectionFailures, connectionResetTime)
		metrics.SourceStatus.WithLabelValues(p.Name()).Set(1)
		metrics.FetchSuccess.WithLabelValues(p.Name()).Add(0)
		metrics.FetchFailure.WithLabelValues(p.Name()).Add(0)
	}
	return &Fetcher{
		providers: providers,
		states:    states,
		store:     st,
		logger:    logger,
	}
}

func (f *Fetcher) Run(ctx context.Context) {
	ticker := time.NewTicker(config.FetchInterval())
	defer ticker.Stop()

	f.fetchAll(ctx)
	for {
		select {
		case <-ctx.Done():
			f.logger.WithField("pair", config.Pair()).Info("fetcher: stopping")
			return
		case <-ticker.C:
			f.fetchAll(ctx)
		}
	}
}

func (f *Fetcher) fetchAll(ctx context.Context) {
	results := make(chan fetchResult, len(f.providers))
	var wg sync.WaitGroup
	for _, p := range f.providers {
		wg.Add(1)
		go func(p provider.PriceProvider) {
			defer wg.Done()
			results <- f.fetchOne(ctx, p)
		}(p)
	}
	go func() { wg.Wait(); close(results) }()

	var prices []float64
	for r := range results {
		if r.err != nil {
			metrics.FetchFailure.WithLabelValues(r.source).Inc()
			metrics.SourceStatus.WithLabelValues(r.source).Set(0)

			f.logger.WithFields(logrus.Fields{
				"source": r.source,
				"error":  r.err,
			}).Warn("fetcher: source failed")
		} else {
			metrics.FetchSuccess.WithLabelValues(r.source).Inc()
			metrics.SourceStatus.WithLabelValues(r.source).Set(1)
			prices = append(prices, r.price)
		}
	}

	agg := aggregator.Aggregate(prices)
	currentPair := config.Pair()

	if agg.OK {
		f.store.Update(currentPair, agg.Price, agg.SourcesUsed)
		metrics.CurrentPrice.Set(agg.Price)

		f.logger.WithFields(logrus.Fields{
			"pair":         currentPair,
			"price":        agg.Price,
			"sources_used": agg.SourcesUsed,
		}).Info("fetcher: price updated")
	} else {
		f.store.MarkStale(currentPair)
		f.logger.WithField("pair", currentPair).Warn("fetcher: all sources failed, marking stale")
	}
}

func (f *Fetcher) fetchOne(ctx context.Context, p provider.PriceProvider) fetchResult {
	name := p.Name()
	state := f.states[name]
	if !state.IsAllowed() {
		return fetchResult{source: name, err: connection.ErrConnectionBroken}
	}

	reqCtx, cancel := context.WithTimeout(ctx, config.RequestTimeout())
	defer cancel()

	start := time.Now()
	var price float64
	attempts, err := retry.Do(reqCtx, config.MaxRetries(), func() error {
		var ferr error
		price, ferr = p.FetchPrice(reqCtx)
		return ferr
	})

	latency := time.Since(start)
	if err != nil {
		state.RecordFailure()
		f.logger.WithFields(logrus.Fields{
			"source":     name,
			"latency_ms": latency.Milliseconds(),
			"attempts":   attempts,
			"error":      err,
		}).Error("fetcher: fetch failed")
		return fetchResult{source: name, err: err}
	}

	state.RecordSuccess()
	f.logger.WithFields(logrus.Fields{
		"source":     name,
		"price":      price,
		"latency_ms": latency.Milliseconds(),
		"attempts":   attempts,
	}).Info("fetcher: fetch ok")

	return fetchResult{source: name, price: price}
}
