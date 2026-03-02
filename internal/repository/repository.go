package repository

import (
	"fmt"
	"time"

	"btcaggregation/internal/cache"
)

type PriceRepository interface {
	Update(pair string, price float64, sourcesUsed int)
	MarkStale(pair string)
	GetResponseData(pair string) (cache.ResponseData, bool)
}

type priceRepository struct {
	c cache.PriceCache
}

func New() PriceRepository {
	return &priceRepository{c: cache.New()}
}

func (r *priceRepository) Update(pair string, price float64, sourcesUsed int) {
	r.c.Set(pair, cache.ResponseData{
		Price:       price,
		Currency:    parseCurrency(pair),
		SourcesUsed: sourcesUsed,
		LastUpdated: time.Now().UTC(),
		Stale:       false,
	})
}

func (r *priceRepository) MarkStale(pair string) {
	if data, ok := r.c.Get(pair); ok {
		data.Stale = true
		r.c.Set(pair, data)
	}
}

func (r *priceRepository) GetResponseData(pair string) (cache.ResponseData, bool) {
	return r.c.Get(pair)
}

func parseCurrency(pair string) string {
	for i, ch := range pair {
		if ch == '/' {
			return fmt.Sprintf("%s", pair[i+1:])
		}
	}
	return "USD"
}
