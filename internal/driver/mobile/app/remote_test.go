// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package app

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func withRemoteTime(t *testing.T, d, grace time.Duration) {
	t.Helper()
	old, oldGrace := RemoteNotificationTime, expireGrace
	RemoteNotificationTime, expireGrace = d, grace
	t.Cleanup(func() { RemoteNotificationTime, expireGrace = old, oldGrace })
}

type completions struct {
	mu   sync.Mutex
	got  []RemoteResult
	done chan struct{}
}

func newCompletions() *completions { return &completions{done: make(chan struct{}, 4)} }

func (c *completions) complete(r RemoteResult) {
	c.mu.Lock()
	c.got = append(c.got, r)
	c.mu.Unlock()
	c.done <- struct{}{}
}

func (c *completions) wait(t *testing.T) []RemoteResult {
	t.Helper()
	select {
	case <-c.done:
	case <-time.After(2 * time.Second):
		t.Fatal("the notification was never completed")
	}
	time.Sleep(50 * time.Millisecond) // a second completion would land here
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]RemoteResult(nil), c.got...)
}

func TestRemoteNotificationReportsWhatTheHandlerReturned(t *testing.T) {
	for _, tc := range []struct {
		newData, ok bool
		want        RemoteResult
	}{{true, true, RemoteNewData}, {false, true, RemoteNoData}, {true, false, RemoteFailed}} {
		r := &remoteState{handler: func(context.Context) (bool, bool) { return tc.newData, tc.ok }}
		c := newCompletions()
		r.receiveNotification(c.complete)
		if got := c.wait(t); len(got) != 1 || got[0] != tc.want {
			t.Fatalf("(%v, %v) completed %v, want once %v", tc.newData, tc.ok, got, tc.want)
		}
	}
}

func TestRemoteNotificationWithoutHandlerFailsAtOnce(t *testing.T) {
	c := newCompletions()
	(&remoteState{}).receiveNotification(c.complete)
	if got := c.wait(t); len(got) != 1 || got[0] != RemoteFailed {
		t.Fatalf("completed %v, want once failed", got)
	}
}

// A handler that ignores its context still lets the notification complete, after the
// grace: iOS kills an app that never calls the completion handler.
func TestRemoteNotificationThatOverrunsCompletesOnceAsFailed(t *testing.T) {
	withRemoteTime(t, 20*time.Millisecond, 20*time.Millisecond)
	release := make(chan struct{})
	r := &remoteState{handler: func(context.Context) (bool, bool) { <-release; return true, true }}
	c := newCompletions()
	r.receiveNotification(c.complete)
	select {
	case <-c.done:
	case <-time.After(2 * time.Second):
		t.Fatal("the notification was never completed")
	}
	close(release) // the handler returns late: it must not complete a second time
	c.done <- struct{}{}
	if got := c.wait(t); len(got) != 1 || got[0] != RemoteFailed {
		t.Fatalf("completed %v, want once failed", got)
	}
}

// One that notices the cancellation and returns in time is failed too, but only once.
func TestRemoteNotificationCancelledReturnsFailedOnce(t *testing.T) {
	withRemoteTime(t, 20*time.Millisecond, time.Second)
	r := &remoteState{handler: func(ctx context.Context) (bool, bool) { <-ctx.Done(); return true, true }}
	c := newCompletions()
	r.receiveNotification(c.complete)
	if got := c.wait(t); len(got) != 1 || got[0] != RemoteFailed {
		t.Fatalf("completed %v, want once failed", got)
	}
}

func TestRemoteTokenGoesToEveryoneWaitingAndAsksOnce(t *testing.T) {
	asked := 0
	r := &remoteState{declared: func() bool { return true }, register: func() { asked++ }}
	got := make(chan string, 2)
	for range 2 {
		r.registerToken(func(token string, err error) { got <- token })
	}
	if asked != 1 {
		t.Fatalf("asked the system %d times, want once", asked)
	}
	r.tokenReceived("ab01", nil)
	for range 2 {
		select {
		case tok := <-got:
			if tok != "ab01" {
				t.Fatalf("token %q", tok)
			}
		case <-time.After(time.Second):
			t.Fatal("a caller never got the token")
		}
	}
	r.registerToken(func(string, error) {})
	if asked != 2 {
		t.Fatal("a later registration asks again: the token can change")
	}
}

func TestRemoteNotificationsUnavailableFailRegistration(t *testing.T) {
	for _, r := range []*remoteState{{}, {declared: func() bool { return false }}} {
		errs := make(chan error, 1)
		r.registerToken(func(_ string, err error) { errs <- err })
		select {
		case err := <-errs:
			if !errors.Is(err, ErrRemoteNotificationsUnavailable) {
				t.Fatalf("err = %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("registration never answered")
		}
	}
}

// The native side cannot be compiled on Linux: these contracts pin what it must do.
func TestIOSRemoteNotificationBridge(t *testing.T) {
	m := string(iosGeometrySource(t, "darwin_ios.m"))
	for _, required := range []string{
		`objectForInfoDictionaryKey:@"UIBackgroundModes"`,
		`containsObject:@"remote-notification"`,
		"[[UIApplication sharedApplication] registerForRemoteNotifications];",
		"didRegisterForRemoteNotificationsWithDeviceToken:(NSData *)deviceToken",
		`[hex appendFormat:@"%02x", bytes[i]];`,
		"remoteTokenReceived((char *)[hex UTF8String], NULL);",
		"didFailToRegisterForRemoteNotificationsWithError:(NSError *)error",
		"remoteTokenReceived(NULL, (char *)[[error localizedDescription] UTF8String]);",
		"didReceiveRemoteNotification:(NSDictionary *)userInfo fetchCompletionHandler:(void (^)(UIBackgroundFetchResult))completionHandler",
		"CFBridgingRetain([completionHandler copy])",
		"CFBridgingRelease((CFTypeRef)handle)",
		"UIBackgroundFetchResultNewData", "UIBackgroundFetchResultNoData", "UIBackgroundFetchResultFailed",
	} {
		if !strings.Contains(m, required) {
			t.Errorf("darwin_ios.m is missing %q", required)
		}
	}
	// The payload is never handed to the app, and protected data is asked for before
	// the handler runs.
	recv := m[strings.Index(m, "didReceiveRemoteNotification:"):]
	recv = recv[:strings.Index(recv, "\n}\n")]
	if strings.Contains(recv, "userInfo[") || strings.Contains(recv, "objectForKey") {
		t.Error("the notification payload must not be read")
	}
	if p, r := strings.Index(recv, "protectedDataChanged(application.isProtectedDataAvailable);"), strings.Index(recv, "remoteNotificationReceived("); p < 0 || r < 0 || p > r {
		t.Error("protected data availability must be refreshed before the handler starts")
	}
	// The results Go sends are the order of RemoteResult.
	if RemoteNewData != 0 || RemoteNoData != 1 || RemoteFailed != 2 {
		t.Fatal("RemoteResult changed: completeRemoteNotification maps 0, 1 and 2")
	}
	g := string(iosGeometrySource(t, "darwin_ios.go"))
	for _, required := range []string{
		"//export remoteTokenReceived", "//export remoteNotificationReceived",
		"remote.declared = func() bool { return bool(C.remoteNotificationsDeclared()) }",
		"remote.register = func() { C.registerForRemoteNotifications() }",
		"C.completeRemoteNotification(handler, C.int(r))",
	} {
		if !strings.Contains(g, required) {
			t.Errorf("darwin_ios.go is missing %q", required)
		}
	}
}
