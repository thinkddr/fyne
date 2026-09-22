// Package urlhandler bridges incoming OS URLs to the application that registered
// for them. It intentionally has no dependency on the public fyne package so the
// mobile and desktop drivers can use it without creating an import cycle.
package urlhandler

import (
	"net/url"
	"sync"
)

var state struct {
	sync.Mutex
	handler  func(*url.URL)
	dispatch func(func())
	pending  []*url.URL
}

// Set installs the receiver and delivers URLs which reached the process before
// the application had finished initialising. The dispatcher must run callbacks in
// the driver's graphical context.
func Set(handler func(*url.URL), dispatch func(func())) {
	state.Lock()
	state.handler, state.dispatch = handler, dispatch
	pending := state.pending
	state.pending = nil
	state.Unlock()

	if handler == nil || dispatch == nil {
		return
	}
	for _, incoming := range pending {
		u := incoming
		dispatch(func() { handler(u) })
	}
}

// Deliver parses and queues an URL from the operating system. Invalid strings and
// ordinary command-line flags are not URLs and are deliberately ignored.
func Deliver(raw string) {
	incoming, err := url.Parse(raw)
	if err != nil || incoming.Scheme == "" {
		return
	}

	state.Lock()
	handler, dispatch := state.handler, state.dispatch
	if handler == nil || dispatch == nil {
		state.pending = append(state.pending, incoming)
		state.Unlock()
		return
	}
	state.Unlock()
	dispatch(func() { handler(incoming) })
}
