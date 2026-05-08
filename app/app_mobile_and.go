//go:build !ci && android

package app

/*
#cgo LDFLAGS: -landroid -llog

#include <stdlib.h>

void openURL(uintptr_t java_vm, uintptr_t jni_env, uintptr_t ctx, char *url);
void sendNotification(uintptr_t java_vm, uintptr_t jni_env, uintptr_t ctx, char *title, char *content);
*/
import "C"

import (
	"net/url"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/driver/mobile/app"
)

func (a *fyneApp) OpenURL(url *url.URL) error {
	urlStr := C.CString(url.String())
	defer C.free(unsafe.Pointer(urlStr))

	app.RunOnJVM(func(vm, env, ctx uintptr) error {
		C.openURL(C.uintptr_t(vm), C.uintptr_t(env), C.uintptr_t(ctx), urlStr)
		return nil
	})
	return nil
}

func (a *fyneApp) SendNotification(n *fyne.Notification) {
	titleStr := C.CString(n.Title)
	defer C.free(unsafe.Pointer(titleStr))
	contentStr := C.CString(n.Content)
	defer C.free(unsafe.Pointer(contentStr))

	app.RunOnJVM(func(vm, env, ctx uintptr) error {
		C.sendNotification(C.uintptr_t(vm), C.uintptr_t(env), C.uintptr_t(ctx), titleStr, contentStr)
		return nil
	})
}

// Native AlarmManager-based scheduling on Android requires registering a
// BroadcastReceiver in AndroidManifest.xml, which is owned by the Fyne packaging
// tool. Until that wiring lands the in-process scheduler with cache persistence
// and replay-on-launch is used; this gives correct timing while the process is
// alive and recovers any past-due deliveries on next launch.
func (a *fyneApp) ScheduleNotification(n *fyne.Notification, when time.Time) (*fyne.ScheduledNotification, error) {
	return a.scheduleViaScheduler(n, when)
}

func (a *fyneApp) CancelScheduledNotification(id string) error {
	return a.cancelViaScheduler(id)
}
