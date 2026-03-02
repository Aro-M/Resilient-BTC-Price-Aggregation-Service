package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Path string

const (
	DotEnv Path = ".env"
)

const (
	fetchInterval   = "FETCH_INTERVAL"
	requestTimeout  = "REQUEST_TIMEOUT"
	maxRetries      = "MAX_RETRIES"
	shutdownTimeout = "SHUTDOWN_TIMEOUT"
	port            = "PORT"
	pair            = "PAIR"
	coinbaseURL     = "COINBASE_URL"
	coindeskURL     = "COINDESK_URL"
	krakenURL       = "KRAKEN_URL"
)

func Init(path Path) error {
	if _, exists := os.LookupEnv("LOAD_FROM_DOCKER_ENV"); !exists {
		if err := godotenv.Load(string(path)); err != nil {
			logrus.WithField("path", path).Warn(".env file not found, reading from system environment")
		} else {
			logrus.WithField("path", path).Info(".env file loaded successfully")
		}
	}

	return checkENV()
}

func checkENV() error {
	vars := []string{fetchInterval, requestTimeout, maxRetries, shutdownTimeout, port, pair, coinbaseURL, coindeskURL, krakenURL}

	for _, v := range vars {
		if _, exists := os.LookupEnv(v); !exists {
			logrus.WithField("variable", v).Error("Environment variable is missing")
			return fmt.Errorf("environment variable %s is missing", v)
		}
	}

	if _, err := getDuration(fetchInterval); err != nil {
		logrus.WithFields(logrus.Fields{
			"variable": fetchInterval,
			"error":    err,
		}).Error("Invalid duration format")
		return err
	}

	if _, err := getDuration(shutdownTimeout); err != nil {
		logrus.WithFields(logrus.Fields{
			"variable": shutdownTimeout,
			"error":    err,
		}).Error("Invalid duration format")
		return err
	}

	if _, err := getInt(maxRetries); err != nil {
		logrus.WithFields(logrus.Fields{
			"variable": maxRetries,
			"error":    err,
		}).Error("Invalid integer format")
		return err
	}

	logrus.Info("All environment variables validated successfully")
	return nil
}

func FetchInterval() time.Duration {
	val, _ := getDuration(fetchInterval)
	return val
}

func RequestTimeout() time.Duration {
	val, _ := getDuration(requestTimeout)
	return val
}

func ShutdownTimeout() time.Duration {
	val, _ := getDuration(shutdownTimeout)
	return val
}

func MaxRetries() int {
	val, _ := getInt(maxRetries)
	return val
}

func Port() string {
	return os.Getenv(port)
}

func Pair() string {
	return os.Getenv(pair)
}

func CoinbaseURL() string {
	return os.Getenv(coinbaseURL)
}

func CoindeskURL() string {
	return os.Getenv(coindeskURL)
}

func KrakenURL() string {
	return os.Getenv(krakenURL)
}

func getDuration(key string) (time.Duration, error) {
	s := os.Getenv(key)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("variable %s: %w", key, err)
	}
	return time.Duration(n) * time.Second, nil
}

func getInt(key string) (int, error) {
	s := os.Getenv(key)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("variable %s: %w", key, err)
	}
	return n, nil
}
