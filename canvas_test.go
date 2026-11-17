package parliament_test

import (
	"testing"

	"github.com/mikkelstb/parliament"
)

func TestDrawing(t *testing.T) {

	p := parliament.NewParliament(150, 10)

	p.Seats = 179
	p.Rows = 7

	p.SetAngle(180)
	if err := p.Draw(); err != nil {
		t.Fatal(err)
	}
}
