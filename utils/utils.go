package utils

import "time"

func DoEvery(d time.Duration, f func(time.Time)) {
	go func() {
		for x := range time.Tick(d) {
			f(x)
		}
	}()
}
