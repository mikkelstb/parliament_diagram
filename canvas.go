package parliament

import (
	"image/color"
	"io"
)

type Canvas struct {
	Width    int
	Height   int
	originX  float64
	originY  float64
	scale    float64
	renderer Renderer
}

func NewCanvas(width, height int) *Canvas {
	return newCanvas(width, height, NewPNGRenderer())
}

func newCanvas(width, height int, renderer Renderer) *Canvas {
	renderer.SetSize(width, height)
	return &Canvas{
		Width:    width,
		Height:   height,
		originX:  float64(width) / 2,
		originY:  float64(height) / 2,
		scale:    1,
		renderer: renderer,
	}
}

func (c *Canvas) DrawCircle(x, y, r float64, cl color.Color) {
	c.renderer.DrawCircle(x*c.scale+c.originX, c.originY-y*c.scale, r*c.scale, cl)
}

func (c *Canvas) Save(filename string) error {
	return c.renderer.Save(filename)
}

func (c *Canvas) Reader() (io.Reader, error) {
	return c.renderer.Reader()
}
