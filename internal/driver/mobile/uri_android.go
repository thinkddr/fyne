//go:build android

package mobile

/*
#cgo LDFLAGS: -landroid -llog

#include <stdlib.h>
#include <stdint.h>

char* contentURIGetFileName(uintptr_t jni_env, uintptr_t ctx, char* uriCstr);
int64_t contentURIGetSize(uintptr_t jni_env, uintptr_t ctx, char* uriCstr);
*/
import "C"

import (
	"path/filepath"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/driver/mobile/app"
	"fyne.io/fyne/v2/storage"
)

type androidURI struct {
	systemURI string
	fyne.URI
}

// Override Name on android for content://
func (a *androidURI) Name() string {
	if a.Scheme() == "content" {
		result := contentURIGetFileName(a.systemURI)
		if result != "" {
			return result
		}
	}
	return a.URI.Name()
}

// Size is what the content provider says the file measures (OpenableColumns.SIZE), or
// -1 when it does not say or the URI is not content://. Not part of fyne.URI: callers
// ask for it with a type assertion on interface{ Size() int64 }.
func (a *androidURI) Size() int64 {
	if a.Scheme() != "content" {
		return -1
	}
	uriStr := C.CString(a.systemURI)
	defer C.free(unsafe.Pointer(uriStr))

	size := int64(-1)
	app.RunOnJVM(func(_, env, ctx uintptr) error {
		size = int64(C.contentURIGetSize(C.uintptr_t(env), C.uintptr_t(ctx), uriStr))
		return nil
	})
	return size
}

func (a *androidURI) Extension() string {
	return filepath.Ext(a.Name())
}

func contentURIGetFileName(uri string) string {
	uriStr := C.CString(uri)
	defer C.free(unsafe.Pointer(uriStr))

	var filename string
	app.RunOnJVM(func(_, env, ctx uintptr) error {
		fnamePtr := C.contentURIGetFileName(C.uintptr_t(env), C.uintptr_t(ctx), uriStr)
		vPtr := unsafe.Pointer(fnamePtr)
		if vPtr == C.NULL {
			return nil
		}
		filename = C.GoString(fnamePtr)
		C.free(vPtr)

		return nil
	})
	return filename
}

func nativeURI(uri string) fyne.URI {
	return &androidURI{URI: storage.NewURI(uri), systemURI: uri}
}
