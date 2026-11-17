# Parliament

Parliament is a small Go library for generating semicircular seating charts. Seats are arranged in rows, colored by party, and can be exported as PNG or SVG.

## Usage

Create a parliament with an inner row radius and a seat radius. The default output width is 800 pixels; use `NewParliamentWithWidth` to choose another fixed width. The image height is calculated to fit the seating chart.

```go
package main

import (
	"io"
	"os"

	"github.com/mikkelstb/parliament"
	"golang.org/x/image/colornames"
)

func main() {
	p := parliament.NewParliamentWithWidth(140, 10, 1000)
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

	if err := p.DrawSVG("parliament.svg"); err != nil {
		panic(err)
	}

	// Reader returns the rendered data for the most recent draw.
	reader, err := p.Reader()
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(os.Stdout, reader); err != nil {
		panic(err)
	}
}
```

`NewParliament(innerRadius, seatRadius)` is equivalent to using `NewParliamentWithWidth(innerRadius, seatRadius, 800)`. The default angle is 180 degrees; call `SetAngle` to change it.

## Rendering

- `DrawPNG(filename)` writes a PNG file.
- `DrawSVG(filename)` writes an SVG file.
- `Draw()` writes a PNG file named `out.png`.
- `DrawWith(renderer, filename)` renders using a custom implementation of the `Renderer` interface.
- `Reader()` returns an `io.Reader` for the data from the most recent successful draw. Call it after a draw method.

The drawing uses the configured width and scales the seating chart proportionally to fit, including a margin around the seats. Its height varies with the chart dimensions.

Each row receives a share of the total seats proportional to its arc length. Within a row, party seat counts are allocated using the Sainte-Laguë method and colored in the order parties were added. Any seats remaining after the listed parties' seat totals are counted as grey “Others” seats.