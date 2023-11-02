package blobadapter

import "time"

// WithTimeout sets the timeout on the adapter.
func WithTimeout(d time.Duration) Option {
	return func(a *Adapter) {
		a.timeout = d
	}
}
