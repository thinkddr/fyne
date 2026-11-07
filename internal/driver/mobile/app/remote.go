// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package app

import (
	"context"
	"errors"
	"sync"
	"time"
)

// RemoteNotification is the handler of background remote notifications
// (OnRemoteNotification).
type RemoteNotification func(ctx context.Context) (newData, ok bool)

// RemoteResult is what a background remote notification reports to the system.
type RemoteResult int

const (
	// RemoteNewData: the handler fetched something new.
	RemoteNewData RemoteResult = iota
	// RemoteNoData: it worked and there was nothing new.
	RemoteNoData
	// RemoteFailed: it failed, timed out or there was no handler.
	RemoteFailed
)

// RemoteNotificationTime is how long a handler has before its context is cancelled. iOS
// gives the app about 30 seconds to call the completion handler; this leaves room for
// the grace and for the main thread.
var RemoteNotificationTime = 25 * time.Second

// ErrRemoteNotificationsUnavailable is the registration error where the app cannot
// receive background remote notifications.
var ErrRemoteNotificationsUnavailable = errors.New("remote notifications are not available")

// remoteState is what the platform tells about background remote notifications. Only
// iOS feeds it (darwin_ios.go); elsewhere declared and register stay nil.
type remoteState struct {
	mu sync.Mutex
	// declared reports whether the app declares the background mode; register asks the
	// system for a token, which arrives later through tokenReceived.
	declared func() bool
	register func()
	waiting  []func(string, error)
	handler  RemoteNotification
}

var remote = &remoteState{}

// RemoteNotificationsAvailable reports whether the app declares the remote-notification
// background mode.
func RemoteNotificationsAvailable() bool { return remote.available() }

// RegisterForRemoteNotifications asks the system for the device token.
func RegisterForRemoteNotifications(f func(token string, err error)) { remote.registerToken(f) }

// OnRemoteNotification sets the handler of background remote notifications.
func OnRemoteNotification(f RemoteNotification) { remote.onNotification(f) }

func (r *remoteState) available() bool {
	r.mu.Lock()
	declared := r.declared
	r.mu.Unlock()
	return declared != nil && declared()
}

func (r *remoteState) registerToken(f func(string, error)) {
	if f == nil {
		return
	}
	if !r.available() {
		go f("", ErrRemoteNotificationsUnavailable)
		return
	}
	r.mu.Lock()
	r.waiting = append(r.waiting, f)
	first, register := len(r.waiting) == 1, r.register
	r.mu.Unlock()
	if first && register != nil {
		register() // one request answers every caller waiting for it
	}
}

func (r *remoteState) onNotification(f RemoteNotification) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handler = f
}

// tokenReceived is the system's answer to the registration: a token or an error.
func (r *remoteState) tokenReceived(token string, err error) {
	r.mu.Lock()
	waiting := r.waiting
	r.waiting = nil
	r.mu.Unlock()
	for _, f := range waiting {
		go f(token, err)
	}
}

// receiveNotification runs the handler for one notification and calls complete exactly
// once: with what the handler returned, or failed when it times out (after the grace) or
// when there is no handler. It does not block its caller.
func (r *remoteState) receiveNotification(complete func(RemoteResult)) {
	r.mu.Lock()
	f := r.handler
	r.mu.Unlock()
	if f == nil {
		complete(RemoteFailed)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), RemoteNotificationTime)
	graceAfter := expireGrace
	var once sync.Once
	finish := func(res RemoteResult) { once.Do(func() { complete(res) }) }
	returned := make(chan struct{})
	go func() {
		defer cancel()
		newData, ok := f(ctx)
		close(returned)
		switch {
		case !ok || ctx.Err() != nil:
			finish(RemoteFailed)
		case newData:
			finish(RemoteNewData)
		default:
			finish(RemoteNoData)
		}
	}()
	go func() {
		<-ctx.Done()
		grace := time.NewTimer(graceAfter)
		defer grace.Stop()
		select {
		case <-returned:
		case <-grace.C:
			finish(RemoteFailed)
		}
	}()
}
