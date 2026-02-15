package fyne

// Cache is used to manage cache storage inside an application sandbox.
// The files managed by this interface are unique to the current application
// and may be deleted by the operating system to clear space
//
// Since: 2.8
type Cache interface {
	RootURI() URI

	Exists(name string) bool
	Get(name string) ([]byte, error)
	Set(name string, data []byte) error
	Remove(name string) error
}
