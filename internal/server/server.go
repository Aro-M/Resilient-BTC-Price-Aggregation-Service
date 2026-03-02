package server

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"btcaggregation/internal/config"
	"btcaggregation/internal/controller"
	"btcaggregation/internal/repository"

	"golang.org/x/time/rate"
)

func New(st repository.PriceRepository, logger *logrus.Logger) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true, LogURI: true, LogMethod: true, LogLatency: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.WithFields(logrus.Fields{
				"method":     v.Method,
				"uri":        v.URI,
				"status":     v.Status,
				"latency_ms": v.Latency.Milliseconds(),
				"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
			}).Info("http request")
			return nil
		},
	}))

	h := controller.New(st)
	e.GET("/price", h.Price)
	e.GET("/health", h.Health)
	e.GET("/metrics", controller.Metrics())

	return e
}

func Start(ctx context.Context, e *echo.Echo, addr string, logger *logrus.Logger) error {
	errCh := make(chan error, 1)

	go func() {
		logger.WithField("addr", addr).Info("server: starting")
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("server: shutting down gracefully")

		shutCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout())
		defer cancel()

		if err := e.Shutdown(shutCtx); err != nil {
			logger.WithField("error", err).Error("server: shutdown failed")
			return err
		}
		return <-errCh
	}
}
