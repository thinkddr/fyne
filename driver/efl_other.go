// +build !linux,!darwin,!windows

package driver

func oSEngineName() string {
	return oSEngineOther
}

func oSWindowInit(w *window) {
}
