package blobadapter

import "time"

// Option is a function that sets options on the adapter.
type Option func(*Adapter)

// WithTimeout sets the timeout on the adapter.
func WithTimeout(d time.Duration) Option {
	return func(a *Adapter) {
		a.timeout = d
	}
}
