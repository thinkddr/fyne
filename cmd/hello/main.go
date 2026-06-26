// Package main loads a very basic Hello World graphical application.
package main

import (
	"log"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome 😀")

			w.(driver.NativeWindow).RunNative(func(ctx any) {
				log.Println("CTX", ctx)
			})
		}),
	))

	w.ShowAndRun()
}
