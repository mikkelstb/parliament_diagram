# Parliament

Parliament is a small Go library for generating semicircular seating charts. Seats are arranged in rows, colored by party, and can be exported as PNG or SVG.

## Usage

Create a parliament with an inner row radius and a seat radius. The default renderer is PNG; use a different renderer in the constructor to select another format. The default output width is 800 pixels; use the width variants to choose another fixed width.

```go
package main

import (
	"io"
	"os"

	"github.com/mikkelstb/parliament"
	"golang.org/x/image/colornames"
)

func main() {
	p := parliament.NewParliamentWithWidthAndRenderer(140, 10, 1000, parliament.NewSVGRenderer())
	p.Seats = 349
	p.Rows = 10
	p.SetAngle(180)

	p.AddParty(parliament.Party{
		Name:  "Example Party",
		Seats: 120,
		Color: colornames.Red,
	})
	p.AddParty(parliament.Party{
		Name:  "Another Party",
		Seats: 90,
		Color: colornames.Blue,
	})

	if err := p.AsFile("parliament.svg"); err != nil {
		panic(err)
	}

	reader, err := p.AsReader()
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(os.Stdout, reader); err != nil {
		panic(err)
	}
}
```

`NewParliament(innerRadius, seatRadius)` is equivalent to using `NewParliamentWithWidth(innerRadius, seatRadius, 800)`. Use `NewParliamentWithRenderer` or `NewParliamentWithWidthAndRenderer` to select an output format. The default angle is 180 degrees; call `SetAngle` to change it.

## Rendering

- `AsFile(filename)` renders with the selected renderer and writes the result to a file.
- `AsReader()` renders with the selected renderer and returns an `io.Reader`.
- Choose PNG or SVG at construction with `NewPNGRenderer()` or `NewSVGRenderer()`. A custom implementation of the `Renderer` interface can also be selected.

The drawing uses the configured width and scales the seating chart proportionally to fit, including a margin around the seats. Its height varies with the chart dimensions.

Each row receives a share of the total seats proportional to its arc length. Within a row, party seat counts are allocated using the Sainte-Laguë method and colored in the order parties were added. Any seats remaining after the listed parties' seat totals are counted as grey “Others” seats.