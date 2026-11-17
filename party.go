package parliament

import "image/color"

type Party struct {
	Name   string
	Symbol string
	Abbrev string
	Seats  int
	Color  color.Color
}
