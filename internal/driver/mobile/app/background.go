// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package app

import (
	"context"
	"sync"
	"time"
)

// BackgroundRefresh is the handler of the app refresh task (OnBackgroundRefresh).
type BackgroundRefresh func(ctx context.Context) (ok bool, retry time.Duration)

// backgroundState is what the platform tells about launches without a user
// interface. Only iOS feeds it (darwin_ios.go); elsewhere the first scene is the
// application itself and no refresh is ever requested.
type backgroundState struct {
	mu sync.Mutex
	// waitsForScene: the platform can launch without a scene and connects one later.
	waitsForScene        bool
	sceneConnected       bool
	launchedInBackground bool
	firstScene           []func()
	protectedUnavailable bool
	every                time.Duration
	refresh              BackgroundRefresh
	running              map[uintptr]*refreshRun
}

// refreshRun is one refresh task the system handed to the app.
type refreshRun struct {
	// cancel is nil until the handler starts.
	cancel   context.CancelFunc
	complete func(ok bool, next time.Duration)
	// returned is closed when the handler returns.
	returned chan struct{}
}

// expireGrace is how long an expired task waits for its handler to return after the
// handler's context is cancelled. iOS suspends the app as soon as the task completes,
// and kills one suspended while holding a file lock in a shared container
// (0xdead10cc): this gives the handler time to notice and let go.
var expireGrace = 500 * time.Millisecond

var background = &backgroundState{}

// LaunchedInBackground reports whether the system launched the process without a
// user interface and no scene has connected since.
func LaunchedInBackground() bool { return background.launchedWithoutScene() }

// OnFirstScene runs f when the first scene connects, or now if there is no scene to
// wait for. See mobile.BackgroundDriver.
func OnFirstScene(f func()) { background.onFirstScene(f) }

// ProtectedDataAvailable reports whether files with complete data protection can be read.
func ProtectedDataAvailable() bool { return background.protectedDataAvailable() }

// OnBackgroundRefresh sets the handler of the app refresh task.
func OnBackgroundRefresh(every time.Duration, f BackgroundRefresh) {
	background.onBackgroundRefresh(every, f)
}

func (b *backgroundState) launchedWithoutScene() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.launchedInBackground && !b.sceneConnected
}

func (b *backgroundState) onFirstScene(f func()) {
	if f == nil {
		return
	}
	b.mu.Lock()
	if b.waitsForScene && !b.sceneConnected {
		b.firstScene = append(b.firstScene, f)
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()
	f()
}

func (b *backgroundState) protectedDataAvailable() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return !b.protectedUnavailable
}

func (b *backgroundState) onBackgroundRefresh(every time.Duration, f BackgroundRefresh) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if f == nil || every <= 0 {
		b.every, b.refresh = 0, nil
		return
	}
	b.every, b.refresh = every, f
}

// launchFinished records how the application finished launching.
func (b *backgroundState) launchFinished(inBackground, protectedData bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.launchedInBackground = inBackground
	b.protectedUnavailable = !protectedData
}

func (b *backgroundState) setProtectedData(available bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.protectedUnavailable = !available
}

// connectScene runs what waited for the first scene, in order. A function registered
// by one of them runs at once: the scene is already there.
func (b *backgroundState) connectScene() {
	b.mu.Lock()
	if b.sceneConnected {
		b.mu.Unlock()
		return
	}
	b.sceneConnected = true
	waiting := b.firstScene
	b.firstScene = nil
	b.mu.Unlock()
	for _, f := range waiting {
		f()
	}
}

// refreshInterval is how far ahead to request the next refresh when the app moves to
// the background: zero without a handler.
func (b *backgroundState) refreshInterval() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.every
}

// receiveRefresh records the system task `task` as soon as the system hands it over,
// before it can expire and before startRefresh. complete is called exactly once: when
// the handler returns, or when the task expires first (at once if the handler had not
// started). next is when to request the following refresh; zero requests none.
func (b *backgroundState) receiveRefresh(task uintptr, complete func(ok bool, next time.Duration)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.running == nil {
		b.running = map[uintptr]*refreshRun{}
	}
	b.running[task] = &refreshRun{complete: complete, returned: make(chan struct{})}
}

// startRefresh runs the handler for the task receiveRefresh recorded. If the task
// already expired, it was completed then and the handler does not run.
func (b *backgroundState) startRefresh(task uintptr) {
	b.mu.Lock()
	run := b.running[task]
	if run == nil || run.cancel != nil {
		b.mu.Unlock()
		return // expired before it started, or already running
	}
	f, every := b.refresh, b.every
	if f == nil {
		delete(b.running, task)
		b.mu.Unlock()
		run.complete(false, 0)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	run.cancel = cancel
	b.mu.Unlock()
	go func() {
		ok, retry := f(ctx)
		close(run.returned)
		next := every
		if retry > 0 {
			next = retry
		}
		b.finishRefresh(task, ok, next)
	}()
}

// expireRefresh is the system ending the task: the handler's context is cancelled and
// the task fails, with the regular interval for the next one, once the handler returns
// or expireGrace passes. It does not block its caller.
func (b *backgroundState) expireRefresh(task uintptr) {
	b.mu.Lock()
	run := b.running[task]
	delete(b.running, task)
	var cancel context.CancelFunc
	if run != nil {
		cancel = run.cancel
	}
	next := b.every
	b.mu.Unlock()
	if run == nil {
		return // already completed
	}
	if cancel == nil {
		run.complete(false, next) // the handler never started: nothing to wait for
		return
	}
	cancel()
	go func() {
		grace := time.NewTimer(expireGrace)
		defer grace.Stop()
		select {
		case <-run.returned:
		case <-grace.C:
		}
		run.complete(false, next)
	}()
}

func (b *backgroundState) finishRefresh(task uintptr, ok bool, next time.Duration) {
	b.mu.Lock()
	run := b.running[task]
	delete(b.running, task)
	b.mu.Unlock()
	if run == nil {
		return // already completed: expired before the handler returned
	}
	run.cancel()
	run.complete(ok, next)
}
