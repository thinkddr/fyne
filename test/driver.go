package test

import (
	"bytes"
	"runtime"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	fynedriver "fyne.io/fyne/v2/driver"
	intdriver "fyne.io/fyne/v2/internal/driver"
	"fyne.io/fyne/v2/internal/painter"
	"fyne.io/fyne/v2/internal/painter/software"
	intRepo "fyne.io/fyne/v2/internal/repository"
	"fyne.io/fyne/v2/storage/repository"
)

// SoftwarePainter describes a simple type that can render canvases
//
// Deprecated: Use driver.Painter instead.
type SoftwarePainter = fynedriver.Painter

type driver struct {
	device       device
	painter      fynedriver.Painter
	windows      []fyne.Window
	windowsMutex sync.RWMutex

	queue     []queuedCall
	queueLock sync.Mutex
}

type queuedCall struct {
	f    func()
	done chan struct{}
}

// Declare conformity with Driver
var _ fyne.Driver = (*driver)(nil)

// NewDriver sets up and registers a new dummy driver for test purpose
func NewDriver() fyne.Driver {
	drv := &driver{windowsMutex: sync.RWMutex{}}
	repository.Register("file", intRepo.NewFileRepository())

	httpHandler := intRepo.NewHTTPRepository()
	repository.Register("http", httpHandler)
	repository.Register("https", httpHandler)

	// make a single dummy window for rendering tests
	drv.CreateWindow("")

	return drv
}

// NewDriverWithPainter creates a new dummy driver that will pass the given
// painter to all canvases created
func NewDriverWithPainter(painter fynedriver.Painter) fyne.Driver {
	return &driver{painter: painter}
}

// DoFromGoroutine runs f straight away when it is called from the test itself, which is
// what tests have always relied on. From any other goroutine (a timer, a background load)
// f is queued instead, and the test goroutine runs the queue at its next synchronisation
// point: a Capture, an input helper of this package or a fyne.DoAndWait of its own. That
// is the one-goroutine model of the real drivers. Running f on the calling goroutine let
// a timer left behind by one test touch the renderer cache while the next test was
// capturing, a data race whose victim changed with every -shuffle seed.
// A new test App gets a new driver, so it drops whatever was still queued.
func (d *driver) DoFromGoroutine(f func(), wait bool) {
	if onTestGoroutine() {
		if wait {
			d.runQueue()
		}
		f()
		return
	}

	call := queuedCall{f: f}
	if wait {
		call.done = make(chan struct{})
	}
	d.queueLock.Lock()
	d.queue = append(d.queue, call)
	d.queueLock.Unlock()
	if wait {
		<-call.done
	}
}

// runQueue runs, in order, what other goroutines had queued when it was called; what
// they queue meanwhile waits for the next call, so a goroutine that keeps queueing
// cannot hold the test forever. Only for the test goroutine.
func (d *driver) runQueue() {
	d.queueLock.Lock()
	n := len(d.queue)
	d.queueLock.Unlock()
	for ; n > 0; n-- {
		d.queueLock.Lock()
		if len(d.queue) == 0 { // a nested call, from one of the functions run here, got there first
			d.queueLock.Unlock()
			return
		}
		call := d.queue[0]
		d.queue = d.queue[1:]
		d.queueLock.Unlock()

		call.f()
		if call.done != nil {
			close(call.done)
		}
	}
}

// synchronise runs the calls queued by other goroutines, if we are on the test goroutine.
func synchronise() {
	if d, ok := fyne.CurrentApp().Driver().(*driver); ok && onTestGoroutine() {
		d.runQueue()
	}
}

// onTestGoroutine reports whether the caller runs on a goroutine of the testing package
// (a test, a subtest, a benchmark, TestMain) rather than on one the code under test started.
func onTestGoroutine() bool {
	buf := make([]byte, 8<<10)
	n := runtime.Stack(buf, false)
	for n == len(buf) { // the testing frames are at the bottom: never look at a truncated stack
		buf = make([]byte, 2*len(buf))
		n = runtime.Stack(buf, false)
	}
	return bytes.Contains(buf[:n], []byte("\ntesting.")) || bytes.Contains(buf[:n], []byte("\nmain.main()"))
}

func (d *driver) AbsolutePositionForObject(co fyne.CanvasObject) fyne.Position {
	c := d.CanvasForObject(co)
	if c == nil {
		return fyne.NewPos(0, 0)
	}

	overlays := c.Overlays().List()
	trees := make([]fyne.CanvasObject, 0, len(overlays)+1)
	if content := c.Content(); content != nil {
		trees = append(trees, content)
	}
	trees = append(trees, overlays...)
	pos := intdriver.AbsolutePositionForObject(co, trees)
	inset, _ := c.InteractiveArea()
	return pos.Subtract(inset)
}

func (d *driver) AllWindows() []fyne.Window {
	d.windowsMutex.RLock()
	defer d.windowsMutex.RUnlock()
	return d.windows
}

func (d *driver) CanvasForObject(fyne.CanvasObject) fyne.Canvas {
	d.windowsMutex.RLock()
	defer d.windowsMutex.RUnlock()
	// cheating: probably the last created window is meant
	return d.windows[len(d.windows)-1].Canvas()
}

func (d *driver) CreateWindow(title string) fyne.Window {
	p := d.painter
	if p == nil {
		p = software.NewPainter()
	}
	c := NewCanvasWithPainter(p)
	w := &window{canvas: c, driver: d, title: title}

	d.windowsMutex.Lock()
	d.windows = append(d.windows, w)
	d.windowsMutex.Unlock()
	return w
}

func (d *driver) Device() fyne.Device {
	return &d.device
}

// RenderedTextSize looks up how bit a string would be if drawn on screen
func (d *driver) RenderedTextSize(text string, size float32, style fyne.TextStyle, source fyne.Resource) (fyne.Size, float32) {
	return painter.RenderedTextSize(text, size, style, source)
}

func (d *driver) Run() {
	// no-op
}

func (d *driver) StartAnimation(a *fyne.Animation) {
	// currently no animations in test app, we just initialise it and leave
	if a.AutoReverse {
		a.Tick(0.0)
	} else {
		a.Tick(1.0)
	}
}

func (d *driver) StopAnimation(a *fyne.Animation) {
	// currently no animations in test app, do nothing
}

func (d *driver) Quit() {
	// no-op
}

func (d *driver) Clipboard() fyne.Clipboard {
	return nil
}

func (d *driver) removeWindow(w *window) {
	d.windowsMutex.Lock()
	i := 0
	for _, win := range d.windows {
		if win == w {
			break
		}
		i++
	}

	copy(d.windows[i:], d.windows[i+1:])
	d.windows[len(d.windows)-1] = nil // Allow the garbage collector to reclaim the memory.
	d.windows = d.windows[:len(d.windows)-1]

	d.windowsMutex.Unlock()
}

func (d *driver) DoubleTapDelay() time.Duration {
	return 300 * time.Millisecond
}

func (d *driver) SetDisableScreenBlanking(_ bool) {
	// no-op for test
}
