package utils

import (
	"math"
	"time"
)

func Retry(logger *Logger, fn func() error) error {
	return retry(logger, 1*time.Minute, 1*time.Second, fn)
}

func RetrySlow(logger *Logger, fn func() error) error {
	return retry(logger, 5*time.Minute, 5*time.Second, fn)
}

func retry(logger *Logger, maxTime time.Duration, baseDelay time.Duration, fn func() error) error {
	var err error
	start := time.Now()
	attempt := 0
	for (time.Now().UnixMilli() - start.UnixMilli()) < maxTime.Milliseconds() {
		err = fn()
		if err == nil {
			break
		}
		logger.Debug.Printf("Attempt failed: %v", err)
		delayFactor := math.Pow(1.1, float64(attempt))
		delay := time.Duration(int64(float64(baseDelay.Milliseconds())*delayFactor) * int64(time.Millisecond))
		time.Sleep(delay)
		attempt = attempt + 1
	}
	return err
}
