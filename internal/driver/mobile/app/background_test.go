// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package app

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestFirstSceneRunsAtOnceWithoutScenes(t *testing.T) {
	b := &backgroundState{}
	ran := false
	b.onFirstScene(func() { ran = true })
	if !ran {
		t.Fatal("a platform without scenes must run the function at once")
	}
	if b.launchedWithoutScene() {
		t.Fatal("a platform without scenes is never launched without one")
	}
}

func TestFirstSceneWaitsForTheSceneAndKeepsTheOrder(t *testing.T) {
	b := &backgroundState{waitsForScene: true}
	var order []string
	b.onFirstScene(func() { order = append(order, "first") })
	b.onFirstScene(func() {
		order = append(order, "second")
		// Registered while connecting: the scene is already there.
		b.onFirstScene(func() { order = append(order, "nested") })
	})
	b.launchFinished(true, true)
	if len(order) != 0 {
		t.Fatalf("ran before any scene: %v", order)
	}
	if !b.launchedWithoutScene() {
		t.Fatal("a background launch without a scene is not reported")
	}
	b.connectScene()
	if want := []string{"first", "second", "nested"}; !slices.Equal(order, want) {
		t.Fatalf("got %v, want %v", order, want)
	}
	if b.launchedWithoutScene() {
		t.Fatal("still reported as without scene after one connected")
	}
	b.connectScene()
	b.onFirstScene(func() { order = append(order, "late") })
	if want := []string{"first", "second", "nested", "late"}; !slices.Equal(order, want) {
		t.Fatalf("a second scene or a late registration: got %v, want %v", order, want)
	}
}

func TestProtectedDataFollowsThePlatform(t *testing.T) {
	b := &backgroundState{waitsForScene: true}
	if !b.protectedDataAvailable() {
		t.Fatal("protected data must be assumed available until the platform says otherwise")
	}
	b.launchFinished(true, false)
	if b.protectedDataAvailable() {
		t.Fatal("a launch with the device locked reports protected data")
	}
	b.setProtectedData(true)
	if !b.protectedDataAvailable() {
		t.Fatal("unlocking is not seen")
	}
}

type completion struct {
	ok   bool
	next time.Duration
}

func TestRefreshWithoutHandlerFailsWithoutRescheduling(t *testing.T) {
	b := &backgroundState{waitsForScene: true}
	got := make(chan completion, 2)
	b.receiveRefresh(1, func(ok bool, next time.Duration) { got <- completion{ok, next} })
	b.startRefresh(1)
	if c := <-got; c.ok || c.next != 0 {
		t.Fatalf("got %+v, want a failure that requests nothing", c)
	}
	if b.refreshInterval() != 0 {
		t.Fatal("an app without a handler must not request refreshes")
	}
}

func TestRefreshCompletesWithTheHandlerResult(t *testing.T) {
	b := &backgroundState{waitsForScene: true}
	retry := time.Duration(0)
	b.onBackgroundRefresh(30*time.Minute, func(context.Context) (bool, time.Duration) { return true, retry })
	if b.refreshInterval() != 30*time.Minute {
		t.Fatal("the interval is not the one registered")
	}
	got := make(chan completion, 2)
	done := func(ok bool, next time.Duration) { got <- completion{ok, next} }
	b.receiveRefresh(1, done)
	b.startRefresh(1)
	if c := <-got; !c.ok || c.next != 30*time.Minute {
		t.Fatalf("got %+v, want success and the regular interval", c)
	}
	retry = 5 * time.Minute
	b.receiveRefresh(2, done)
	b.startRefresh(2)
	if c := <-got; !c.ok || c.next != 5*time.Minute {
		t.Fatalf("got %+v, want the handler's retry", c)
	}
}

func TestExpiredRefreshCancelsAndCompletesOnce(t *testing.T) {
	b := &backgroundState{waitsForScene: true}
	started, released := make(chan struct{}), make(chan struct{})
	b.onBackgroundRefresh(time.Hour, func(ctx context.Context) (bool, time.Duration) {
		close(started)
		<-ctx.Done()
		<-released
		return true, time.Minute
	})
	got := make(chan completion, 2)
	b.receiveRefresh(7, func(ok bool, next time.Duration) { got <- completion{ok, next} })
	b.startRefresh(7)
	<-started
	b.expireRefresh(7)
	if c := <-got; c.ok || c.next != time.Hour {
		t.Fatalf("got %+v, want a failure now, with the regular interval", c)
	}
	close(released) // the handler returns late: nothing else is reported
	time.Sleep(20 * time.Millisecond)
	select {
	case c := <-got:
		t.Fatalf("the task completed twice: %+v", c)
	default:
	}
	b.expireRefresh(7) // an expiration after completion is ignored
	select {
	case c := <-got:
		t.Fatalf("an expiration after completion was reported: %+v", c)
	default:
	}
}

func TestExpiredRefreshWaitsForTheHandlerToLetGo(t *testing.T) {
	b := &backgroundState{waitsForScene: true}
	started := make(chan struct{})
	var letGo time.Time
	b.onBackgroundRefresh(time.Hour, func(ctx context.Context) (bool, time.Duration) {
		close(started)
		<-ctx.Done()
		time.Sleep(50 * time.Millisecond) // releasing a lock in the shared container
		letGo = time.Now()
		return true, time.Minute
	})
	got := make(chan time.Time, 2)
	var c completion
	b.receiveRefresh(8, func(ok bool, next time.Duration) {
		c = completion{ok, next}
		got <- time.Now()
	})
	b.startRefresh(8)
	<-started
	b.expireRefresh(8)
	done := <-got
	if done.Before(letGo) || letGo.IsZero() {
		t.Fatal("the task completed before the handler let go")
	}
	if c.ok || c.next != time.Hour {
		t.Fatalf("got %+v, want a failure with the regular interval, whatever the handler says", c)
	}
	select {
	case <-got:
		t.Fatal("the task completed twice")
	case <-time.After(20 * time.Millisecond):
	}
}

func TestExpiredRefreshDoesNotWaitForeverForTheHandler(t *testing.T) {
	defer func(g time.Duration) { expireGrace = g }(expireGrace)
	expireGrace = 30 * time.Millisecond
	b := &backgroundState{waitsForScene: true}
	stuck := make(chan struct{})
	defer close(stuck)
	b.onBackgroundRefresh(time.Hour, func(context.Context) (bool, time.Duration) {
		<-stuck
		return true, 0
	})
	got := make(chan completion, 1)
	b.receiveRefresh(9, func(ok bool, next time.Duration) { got <- completion{ok, next} })
	b.startRefresh(9)
	b.expireRefresh(9)
	select {
	case c := <-got:
		if c.ok {
			t.Fatalf("got %+v, want a failure", c)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("an expired task waited for a handler that never returns")
	}
}

// The expiration handler is set before the handler starts on the main thread: an
// expiration that comes first completes the task at once, and the handler never runs.
func TestRefreshExpiredBeforeItStarts(t *testing.T) {
	b := &backgroundState{waitsForScene: true}
	ran := false
	b.onBackgroundRefresh(time.Hour, func(context.Context) (bool, time.Duration) {
		ran = true
		return true, 0
	})
	got := make(chan completion, 2)
	b.receiveRefresh(10, func(ok bool, next time.Duration) { got <- completion{ok, next} })
	b.expireRefresh(10)
	select {
	case c := <-got:
		if c.ok || c.next != time.Hour {
			t.Fatalf("got %+v, want a failure with the regular interval", c)
		}
	case <-time.After(time.Second):
		t.Fatal("a task that expired before its handler started was not completed")
	}
	b.startRefresh(10)
	time.Sleep(20 * time.Millisecond)
	if ran {
		t.Fatal("the handler ran for a task that had already expired")
	}
	select {
	case c := <-got:
		t.Fatalf("the task completed twice: %+v", c)
	default:
	}
}

func TestClearingTheRefreshHandler(t *testing.T) {
	b := &backgroundState{}
	b.onBackgroundRefresh(time.Hour, func(context.Context) (bool, time.Duration) { return true, 0 })
	b.onBackgroundRefresh(time.Hour, nil)
	if b.refreshInterval() != 0 {
		t.Fatal("a nil handler must stop requesting refreshes")
	}
}

// The native side cannot be compiled on Linux: these contracts pin what it must do.
func TestIOSBackgroundRefreshBridge(t *testing.T) {
	m := string(iosGeometrySource(t, "darwin_ios.m"))
	for _, required := range []string{
		"#import <BackgroundTasks/BackgroundTasks.h>",
		"objectForInfoDictionaryKey:@\"BGTaskSchedulerPermittedIdentifiers\"",
		"registerForTaskWithIdentifier:identifier usingQueue:nil launchHandler:",
		"task.expirationHandler = ^{",
		"backgroundRefreshReceived(handle);",
		"backgroundRefreshStarted(handle)",
		"backgroundRefreshExpired(handle)",
		"[task setTaskCompleted:ok ? YES : NO]",
		"BGAppRefreshTaskRequest alloc",
		"submitTaskRequest:request error:&error",
		"launchFinished(application.applicationState == UIApplicationStateBackground, application.isProtectedDataAvailable)",
		"applicationProtectedDataWillBecomeUnavailable",
		"applicationProtectedDataDidBecomeAvailable",
		"scheduleBackgroundRefresh(backgroundRefreshInterval())",
	} {
		if !strings.Contains(m, required) {
			t.Errorf("darwin_ios.m is missing %q", required)
		}
	}
	// Each task asks UIKit whether protected data is available before the handler runs:
	// the delegate notifications can reach a suspended app late.
	handler := m[strings.Index(m, "launchHandler:"):]
	handler = handler[:strings.Index(handler, "}];")]
	if p, s := strings.Index(handler, "protectedDataChanged([UIApplication sharedApplication].isProtectedDataAvailable);"), strings.Index(handler, "backgroundRefreshStarted(handle);"); p < 0 || s < 0 || p > s {
		t.Error("the launch handler must refresh protected data availability before starting the handler")
	}
	// Go knows the task before it can expire: an expiration that beats the main thread
	// completes it instead of being lost.
	if r, e := strings.Index(handler, "backgroundRefreshReceived(handle);"), strings.Index(handler, "task.expirationHandler = ^{"); r < 0 || e < 0 || r > e {
		t.Error("the launch handler must record the task before setting its expiration handler")
	}
	// Registration must happen before launching finishes, or BGTaskScheduler throws.
	launch := m[strings.Index(m, "didFinishLaunchingWithOptions"):]
	launch = launch[:strings.Index(launch, "return YES;")]
	if !strings.Contains(launch, "registerBackgroundRefresh();") {
		t.Error("the refresh task is not registered during didFinishLaunching")
	}
	// The first-scene functions run before the scene's URLs reach the app.
	connect := m[strings.Index(m, "willConnectToSession"):]
	if s, u := strings.Index(connect, "sceneWillConnect();"), strings.Index(connect, "openURLContexts:connectionOptions.URLContexts"); s < 0 || u < 0 || s > u {
		t.Error("sceneWillConnect must run before the connection URLs are delivered")
	}
	// A scene connected outside the foreground does not run them until it enters it.
	will := connect[:strings.Index(connect, "connectWithOptions:(UISceneConnectionOptions *)")]
	if !strings.Contains(will, "state != UISceneActivationStateForegroundActive && state != UISceneActivationStateForegroundInactive") ||
		!strings.Contains(will, "self.pendingConnection = connectionOptions;") || strings.Contains(will, "sceneWillConnect();") {
		t.Error("a scene connected outside the foreground must leave sceneWillConnect for its first foreground")
	}
	fg := m[strings.Index(m, "- (void)sceneWillEnterForeground:"):]
	fg = fg[:strings.Index(fg, "\n}\n")]
	if !strings.Contains(fg, "[self scene:scene connectWithOptions:pending];") {
		t.Error("sceneWillEnterForeground must connect a scene that connected in the background")
	}
	if strings.Count(m, "scheduleBackgroundRefresh(backgroundRefreshInterval())") < 2 {
		t.Error("both the scene and the application background transitions must request a refresh")
	}
	g := string(iosGeometrySource(t, "darwin_ios.go"))
	for _, required := range []string{
		"-framework BackgroundTasks", "background.waitsForScene = true",
		"//export sceneWillConnect", "//export launchFinished", "//export protectedDataChanged",
		"//export backgroundRefreshInterval", "//export backgroundRefreshReceived", "//export backgroundRefreshStarted", "//export backgroundRefreshExpired",
		"C.completeBackgroundRefresh(task",
	} {
		if !strings.Contains(g, required) {
			t.Errorf("darwin_ios.go is missing %q", required)
		}
	}
}
