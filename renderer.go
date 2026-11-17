package parliament

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/fogleman/gg"
)

type Renderer interface {
	SetSize(width, height int)
	DrawCircle(x, y, radius float64, cl color.Color)
	Save(filename string) error
	Reader() (io.Reader, error)
}

type PNGRenderer struct {
	context *gg.Context
}

func NewPNGRenderer() *PNGRenderer {
	return &PNGRenderer{}
}

func (r *PNGRenderer) SetSize(width, height int) {
	r.context = gg.NewContext(width, height)
}

func (r *PNGRenderer) DrawCircle(x, y, radius float64, cl color.Color) {
	r.context.DrawCircle(x, y, radius)
	r.context.SetColor(cl)
	r.context.Fill()
}

func (r *PNGRenderer) encode() ([]byte, error) {
	var data bytes.Buffer
	if err := png.Encode(&data, r.context.Image()); err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func (r *PNGRenderer) Save(filename string) error {
	data, err := r.encode()
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0o644)
}

func (r *PNGRenderer) Reader() (io.Reader, error) {
	data, err := r.encode()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

type SVGRenderer struct {
	width   int
	height  int
	circles strings.Builder
}

func NewSVGRenderer() *SVGRenderer {
	return &SVGRenderer{}
}

func (r *SVGRenderer) SetSize(width, height int) {
	r.width = width
	r.height = height
	r.circles.Reset()
}

func (r *SVGRenderer) DrawCircle(x, y, radius float64, cl color.Color) {
	nrgba := color.NRGBAModel.Convert(cl).(color.NRGBA)
	fmt.Fprintf(&r.circles, `<circle cx="%g" cy="%g" r="%g" fill="#%02x%02x%02x"`, x, y, radius, nrgba.R, nrgba.G, nrgba.B)
	if nrgba.A < 255 {
		fmt.Fprintf(&r.circles, ` fill-opacity="%g"`, float64(nrgba.A)/255)
	}
	r.circles.WriteString("/>\n")
}

func (r *SVGRenderer) encode() []byte {
	var svg strings.Builder
	fmt.Fprintf(&svg, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`+"\n", r.width, r.height, r.width, r.height)
	svg.WriteString(r.circles.String())
	svg.WriteString("</svg>\n")
	return []byte(svg.String())
}

func (r *SVGRenderer) Save(filename string) error {
	return os.WriteFile(filename, r.encode(), 0o644)
}

func (r *SVGRenderer) Reader() (io.Reader, error) {
	return bytes.NewReader(r.encode()), nil
}
