package parliament

import (
	"bytes"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/image/colornames"
)

func TestCalculateSeatsProportionalToRowLength(t *testing.T) {
	p := NewParliament(100, 4)
	p.Rows = 3
	p.Seats = 60

	got := p.calculateSeats()
	want := []int{18, 20, 22}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("calculateSeats() = %v, want %v", got, want)
	}
}

func TestAllocatePartySeatsUsesSainteLague(t *testing.T) {
	p := NewParliament(100, 4)
	p.Parties = []Party{
		{Name: "A", Seats: 1000},
		{Name: "B", Seats: 600},
		{Name: "C", Seats: 300},
	}

	got := p.allocatePartySeats(7)
	want := []int{4, 2, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("allocatePartySeats(7) = %v, want %v", got, want)
	}
}

func TestAllocatePartySeatsFillsRemainderWithOthers(t *testing.T) {
	p := NewParliament(100, 4)
	p.Seats = 10
	p.Parties = []Party{
		{Name: "A", Seats: 3},
		{Name: "B", Seats: 3},
	}

	got := p.allocatePartySeats(10)
	want := []int{3, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("allocatePartySeats(10) = %v, want %v", got, want)
	}
}

func TestDrawColorsSeatsLeftToRightByPartyOrder(t *testing.T) {
	p := NewParliament(40, 2)
	p.Rows = 1
	p.Seats = 4
	p.AddParty(Party{Name: "A", Seats: 1, Color: colornames.Red})
	p.AddParty(Party{Name: "B", Seats: 1, Color: colornames.Blue})
	if err := p.AsFile(filepath.Join(t.TempDir(), "parliament.png")); err != nil {
		t.Fatal(err)
	}

	image := p.canvas.renderer.(*PNGRenderer).context.Image()
	scale := p.canvas.scale
	y := int(math.Round(p.canvas.originY))
	leftX := int(math.Round(p.canvas.originX - p.inner_radius*scale))
	assertPixelColor(t, image.At(leftX, y), colornames.Red)

	secondX := int(math.Round(p.canvas.originX - p.inner_radius*scale/2))
	secondY := int(math.Round(p.canvas.originY - p.inner_radius*scale*math.Sin(2*math.Pi/3)))
	assertPixelColor(t, image.At(secondX, secondY), colornames.Blue)

	rightX := int(math.Round(p.canvas.originX + p.inner_radius*scale))
	assertPixelColor(t, image.At(rightX, y), colornames.Grey)
}

func TestDrawColorsFollowArcOrderAbove180Degrees(t *testing.T) {
	p := NewParliament(40, 2)
	p.Rows = 1
	p.Seats = 4
	p.SetAngle(240)
	p.AddParty(Party{Name: "A", Seats: 1, Color: colornames.Red})
	p.AddParty(Party{Name: "B", Seats: 1, Color: colornames.Blue})
	if err := p.AsFile(filepath.Join(t.TempDir(), "parliament.png")); err != nil {
		t.Fatal(err)
	}

	image := p.canvas.renderer.(*PNGRenderer).context.Image()
	scale := p.canvas.scale
	leftX := int(math.Round(p.canvas.originX + p.inner_radius*scale*math.Cos(7*math.Pi/6)))
	leftY := int(math.Round(p.canvas.originY - p.inner_radius*scale*math.Sin(7*math.Pi/6)))
	assertPixelColor(t, image.At(leftX, leftY), colornames.Red)

	secondX := int(math.Round(p.canvas.originX + p.inner_radius*scale*math.Cos(13*math.Pi/18)))
	secondY := int(math.Round(p.canvas.originY - p.inner_radius*scale*math.Sin(13*math.Pi/18)))
	assertPixelColor(t, image.At(secondX, secondY), colornames.Blue)

	rightX := int(math.Round(p.canvas.originX + p.inner_radius*scale*math.Cos(-math.Pi/6)))
	rightY := int(math.Round(p.canvas.originY - p.inner_radius*scale*math.Sin(-math.Pi/6)))
	assertPixelColor(t, image.At(rightX, rightY), colornames.Grey)
}

func TestDrawUsesConfiguredWidth(t *testing.T) {
	p := NewParliamentWithWidth(40, 2, 320)
	p.Rows = 1
	p.Seats = 4
	if err := p.AsFile(filepath.Join(t.TempDir(), "parliament.png")); err != nil {
		t.Fatal(err)
	}

	if p.canvas.Width != 320 {
		t.Fatalf("canvas width = %d, want 320", p.canvas.Width)
	}
}

func TestDrawLeavesSeatRadiusMargin(t *testing.T) {
	p := NewParliament(140, 10)
	p.Rows = 10
	p.Seats = 349
	p.SetAngle(180)

	p.AddParty(Party{Name: "Vänsterpartiet", Symbol: "V", Abbrev: "V", Color: colornames.Darkred, Seats: 30})
	p.AddParty(Party{Name: "Miljöpartiet", Symbol: "MP", Abbrev: "MP", Color: colornames.Lightgreen, Seats: 22})
	p.AddParty(Party{Name: "Socialdemokraterna", Symbol: "S", Abbrev: "S", Color: colornames.Red, Seats: 99})
	p.AddParty(Party{Name: "Centerpartiet", Symbol: "C", Abbrev: "C", Color: colornames.Darkolivegreen, Seats: 25})
	p.AddParty(Party{Name: "Libdem", Symbol: "C", Abbrev: "C", Color: colornames.Lightblue, Seats: 22})
	p.AddParty(Party{Name: "Moderaterna", Symbol: "M", Abbrev: "M", Color: colornames.Blue, Seats: 70})
	p.AddParty(Party{Name: "Kristdemokraterna", Symbol: "M", Abbrev: "M", Color: colornames.Navy, Seats: 22})
	p.AddParty(Party{Name: "Sverigedemokraterna", Symbol: "SD", Abbrev: "SD", Color: colornames.Yellow, Seats: 62})

	filename := filepath.Join(t.TempDir(), "parliament.png")
	if err := p.AsFile(filename); err != nil {
		t.Fatal(err)
	}
	reader, err := p.AsReader()
	if err != nil {
		t.Fatal(err)
	}
	readerData, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	fileData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readerData, fileData) {
		t.Fatal("returned PNG reader does not match the saved file")
	}

	image := p.canvas.renderer.(*PNGRenderer).context.Image()
	bounds := image.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y
	drawn := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := image.At(x, y).RGBA()
			if alpha == 0 {
				continue
			}
			drawn = true
			minX = min(minX, x)
			minY = min(minY, y)
			maxX = max(maxX, x)
			maxY = max(maxY, y)
		}
	}

	if !drawn {
		t.Fatal("Draw() produced no visible circles")
	}
	if minX < 10 || minY < 10 || bounds.Max.X-1-maxX < 10 || bounds.Max.Y-1-maxY < 10 {
		t.Fatalf("drawn circles don't have a 10px margin: image bounds %v, drawn bounds (%d,%d)-(%d,%d)", bounds, minX, minY, maxX, maxY)
	}

	left, top := minX, minY
	right, bottom := bounds.Max.X-1-maxX, bounds.Max.Y-1-maxY
	if left-right > 1 || right-left > 1 || top-bottom > 1 || bottom-top > 1 {
		t.Fatalf("margins are not equal: left=%d top=%d right=%d bottom=%d", left, top, right, bottom)
	}
}

func TestDrawSVG(t *testing.T) {
	p := NewParliamentWithRenderer(40, 2, NewSVGRenderer())
	p.Rows = 1
	p.Seats = 1
	p.AddParty(Party{Name: "A", Seats: 1, Color: colornames.Red})
	filename := filepath.Join(t.TempDir(), "parliament.svg")

	if err := p.AsFile(filename); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `<svg `) || !strings.Contains(string(data), `fill="#ff0000"`) {
		t.Fatalf("SVG output does not contain expected root and party color: %s", data)
	}

	reader, err := p.AsReader()
	if err != nil {
		t.Fatal(err)
	}
	readerData, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readerData, data) {
		t.Fatal("returned SVG reader does not match the saved file")
	}
}

func TestAsReaderRendersPNGWithoutFileOutput(t *testing.T) {
	p := NewParliament(40, 2)
	p.Rows = 1
	p.Seats = 1
	p.AddParty(Party{Name: "A", Seats: 1, Color: colornames.Red})
	reader, err := p.AsReader()
	if err != nil {
		t.Fatal(err)
	}
	readerData, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	image, err := png.Decode(bytes.NewReader(readerData))
	if err != nil {
		t.Fatalf("AsReader() did not return a PNG: %v", err)
	}
	if image.Bounds().Dx() != p.canvasWidth {
		t.Fatalf("PNG width = %d, want %d", image.Bounds().Dx(), p.canvasWidth)
	}
}

func assertPixelColor(t *testing.T, got, want color.Color) {
	t.Helper()
	gotNRGBA := color.NRGBAModel.Convert(got).(color.NRGBA)
	wantNRGBA := color.NRGBAModel.Convert(want).(color.NRGBA)
	if gotNRGBA != wantNRGBA {
		t.Fatalf("pixel color = %#v, want %#v", gotNRGBA, wantNRGBA)
	}
}
