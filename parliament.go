package parliament

import (
	"fmt"
	"image/color"
	"io"
	"math"
	"sort"

	"golang.org/x/image/colornames"
)

type Parliament struct {
	canvas   Canvas
	renderer Renderer
	Seats    int
	Rows     int

	angle float64

	inner_radius float64
	seat_radius  float64
	row_distance float64

	canvasWidth int
	Parties     []Party
}

func NewParliament(inner_radius, seat_radius float64) *Parliament {
	return NewParliamentWithWidthAndRenderer(inner_radius, seat_radius, 800, NewPNGRenderer())
}

func NewParliamentWithRenderer(inner_radius, seat_radius float64, renderer Renderer) *Parliament {
	return NewParliamentWithWidthAndRenderer(inner_radius, seat_radius, 800, renderer)
}

func NewParliamentWithWidth(inner_radius, seat_radius float64, width int) *Parliament {
	return NewParliamentWithWidthAndRenderer(inner_radius, seat_radius, width, NewPNGRenderer())
}

func NewParliamentWithWidthAndRenderer(inner_radius, seat_radius float64, width int, renderer Renderer) *Parliament {
	p := &Parliament{
		inner_radius: inner_radius,
		seat_radius:  seat_radius,
		canvasWidth:  width,
		renderer:     renderer,
		// default angle is 180 degrees
		angle: math.Pi,
	}

	p.row_distance = seat_radius * 2.5

	return p
}

func (p *Parliament) AddParty(pt Party) {
	p.Parties = append(p.Parties, pt)
}

func (p *Parliament) allocatePartySeats(rowSeatCount int) []int {
	partySeatTotal := 0
	for _, party := range p.Parties {
		if party.Seats > 0 {
			partySeatTotal += party.Seats
		}
	}

	othersSeats := p.Seats - partySeatTotal
	partyCount := len(p.Parties)
	if othersSeats > 0 {
		partyCount++
	}

	partySeats := make([]int, partyCount)
	for range rowSeatCount {
		bestParty := -1
		bestQuotient := 0.0
		for i := range partySeats {
			weight := othersSeats
			if i < len(p.Parties) {
				weight = p.Parties[i].Seats
			}
			if weight <= 0 {
				continue
			}
			quotient := float64(weight) / (2*float64(partySeats[i]) + 1)
			if quotient > bestQuotient {
				bestParty = i
				bestQuotient = quotient
			}
		}
		if bestParty < 0 {
			break
		}
		partySeats[bestParty]++
	}
	return partySeats
}

func (p *Parliament) calculateSeats() []int {
	if p.Rows <= 0 {
		return nil
	}

	rows := make([]int, p.Rows)
	if p.Seats <= 0 {
		return rows
	}

	lengths := make([]float64, p.Rows)
	var totalLength float64
	for i := range lengths {
		radius := p.inner_radius + float64(i)*p.row_distance
		lengths[i] = radius * p.angle
		totalLength += lengths[i]
	}
	if totalLength <= 0 {
		return rows
	}

	type remainder struct {
		row      int
		fraction float64
	}
	remainders := make([]remainder, p.Rows)
	allocated := 0
	for i, length := range lengths {
		exact := float64(p.Seats) * length / totalLength
		rows[i] = int(math.Floor(exact))
		allocated += rows[i]
		remainders[i] = remainder{row: i, fraction: exact - float64(rows[i])}
	}

	sort.SliceStable(remainders, func(i, j int) bool {
		return remainders[i].fraction > remainders[j].fraction
	})
	for i := 0; i < p.Seats-allocated; i++ {
		rows[remainders[i].row]++
	}

	return rows
}

func (p *Parliament) SetAngle(degrees int) {
	p.angle = float64(degrees) * math.Pi / 180
}

func (p *Parliament) AsFile(filename string) error {
	if err := p.render(); err != nil {
		return err
	}

	return p.canvas.Save(filename)
}

func (p *Parliament) render() error {
	if p.renderer == nil {
		return fmt.Errorf("renderer must not be nil")
	}
	if p.canvasWidth <= 0 {
		return fmt.Errorf("canvas width must be positive")
	}

	rowseats := p.calculateSeats()
	firstangle := (math.Pi - p.angle) / 2

	type point struct {
		x     float64
		y     float64
		color color.Color
	}
	points := make([]point, 0)
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	for row, seatCount := range rowseats {
		if seatCount == 0 {
			continue
		}
		rowradius := p.inner_radius + float64(row)*p.row_distance
		angle_interval := 0.0
		if seatCount > 1 {
			angle_interval = p.angle / float64(seatCount-1)
		}

		partySeats := p.allocatePartySeats(seatCount)
		seatColors := make([]color.Color, 0, seatCount)
		for i, party := range p.Parties {
			partyColor := party.Color
			if partyColor == nil {
				partyColor = colornames.Grey
			}
			for range partySeats[i] {
				seatColors = append(seatColors, partyColor)
			}
		}
		if len(partySeats) > len(p.Parties) {
			for range partySeats[len(p.Parties)] {
				seatColors = append(seatColors, colornames.Grey)
			}
		}
		for len(seatColors) < seatCount {
			seatColors = append(seatColors, colornames.Grey)
		}

		for seat := range seatCount {
			seatAngle := firstangle + p.angle
			if seatCount == 1 {
				seatAngle = firstangle + p.angle/2
			} else {
				seatAngle -= float64(seat) * angle_interval
			}
			point := point{
				x:     math.Cos(seatAngle) * rowradius,
				y:     math.Sin(seatAngle) * rowradius,
				color: seatColors[seat],
			}
			points = append(points, point)
			x, y := point.x, point.y
			minX = math.Min(minX, x)
			minY = math.Min(minY, y)
			maxX = math.Max(maxX, x)
			maxY = math.Max(maxY, y)
		}
	}

	padding := 2 * p.seat_radius
	if len(points) == 0 {
		minX, minY, maxX, maxY = 0, 0, 0, 0
	}
	contentWidth := maxX - minX + 2*padding
	contentHeight := maxY - minY + 2*padding
	if contentWidth <= 0 {
		contentWidth = 1
	}
	scale := float64(p.canvasWidth) / contentWidth
	height := int(math.Ceil(contentHeight * scale))
	if height < 1 {
		height = 1
	}
	p.canvas = *newCanvas(p.canvasWidth, height, p.renderer)
	p.canvas.scale = scale
	p.canvas.originX = (padding - minX) * scale
	p.canvas.originY = (padding + maxY) * scale

	for _, point := range points {
		p.canvas.DrawCircle(point.x, point.y, p.seat_radius, point.color)
	}

	return nil
}

func (p *Parliament) AsReader() (io.Reader, error) {
	if err := p.render(); err != nil {
		return nil, err
	}
	return p.canvas.Reader()
}
