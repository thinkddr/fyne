package mobile

import "testing"

func TestNmREEscapesHostnameDot(t *testing.T) {
	if got := nmRE.FindStringSubmatch("00000000 t fyne.io/fyne/v2/internal/driver/mobile/app"); got == nil {
		t.Fatal("expected fyne.io package path to match")
	}
	if got := nmRE.FindStringSubmatch("00000000 t fyneXio/fyne/v2/internal/driver/mobile/app"); got != nil {
		t.Fatalf("unexpected match for hostname with an unescaped separator: %q", got[1])
	}
}
