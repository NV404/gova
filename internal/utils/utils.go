package utils

import (
	"errors"
	"os"
	"time"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func Debounce(fn func(), delay time.Duration) func() {
	var timer *time.Timer
	ch := make(chan struct{}, 1)

	go func() {
		for range ch {
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(delay, fn)
		}
	}()

	return func() {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
