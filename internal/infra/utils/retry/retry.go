package utils

import "time"

func Retry(f func() error, retryTimes int, delay time.Duration) error {
	for retryTimes > 0 {
		if err := f(); err != nil {
			time.Sleep(delay)
			retryTimes--

			continue
		}
		return nil
	}

	return nil
}
