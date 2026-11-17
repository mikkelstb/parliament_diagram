package parliament_test

import (
	"path/filepath"
	"testing"

	parliament "github.com/mikkelstb/parliament_diagram"
)

func TestDrawing(t *testing.T) {

	p := parliament.NewParliament(150, 10)

	p.Seats = 179
	p.Rows = 7

	p.SetAngle(180)
	if err := p.AsFile(filepath.Join(t.TempDir(), "parliament.png")); err != nil {
		t.Fatal(err)
	}
}
