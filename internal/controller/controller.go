package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"btcaggregation/internal/config"
	"btcaggregation/internal/repository"
)

type PriceResponse struct {
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	SourcesUsed int     `json:"sources_used"`
	LastUpdated string  `json:"last_updated"`
	Stale       bool    `json:"stale"`
}

type Controller struct{ store repository.PriceRepository }

func New(st repository.PriceRepository) *Controller { return &Controller{store: st} }

func (h *Controller) Price(c echo.Context) error {
	pair := c.QueryParam("pair")
	if pair == "" {
		pair = config.Pair()
	}
	data, ok := h.store.GetResponseData(pair)
	if !ok {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "no price data available yet for " + pair,
		})
	}
	return c.JSON(http.StatusOK, PriceResponse{
		Price:       data.Price,
		Currency:    data.Currency,
		SourcesUsed: data.SourcesUsed,
		LastUpdated: data.LastUpdated.Format("2006-01-02T15:04:05Z"),
		Stale:       data.Stale,
	})
}

func (h *Controller) Health(c echo.Context) error {
	pair := c.QueryParam("pair")
	if pair == "" {
		pair = config.Pair()
	}
	data, ok := h.store.GetResponseData(pair)
	if !ok {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "no data yet"})
	}
	if data.Stale {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "all sources failing"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func Metrics() echo.HandlerFunc {
	h := promhttp.Handler()
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
