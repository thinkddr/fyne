package urlhandler

import (
	"net/url"
	"testing"
)

func reset() {
	state.Lock()
	state.handler, state.dispatch, state.pending = nil, nil, nil
	state.Unlock()
}

func TestQueuesURLUntilAHandlerIsInstalled(t *testing.T) {
	reset()
	t.Cleanup(reset)
	Deliver("sytue-mail://auth/callback?code=one")

	var got string
	Set(func(u *url.URL) { got = u.String() }, func(f func()) { f() })
	if got != "sytue-mail://auth/callback?code=one" {
		t.Fatalf("URL entregada = %q", got)
	}
}

func TestRejectsArgumentsThatAreNotURLs(t *testing.T) {
	reset()
	t.Cleanup(reset)
	Deliver("-psn_0_12345")
	called := false
	Set(func(*url.URL) { called = true }, func(f func()) { f() })
	if called {
		t.Fatal("un argumento normal no puede llegar al handler")
	}
}
