package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func (w *Window) drawPerson(screen *ebiten.Image, i int) {
	person := &w.People[i]
	x, y, width, height := w.personRect(person)

	switch w.colorMode {
	case ByNumConnections:
		connections := len(person.Connections)

		t := float64(connections) / (30)
		if t > 1 {
			t = 1
		} else if t < 0 {
			t = 0
		}

		// more connections more red, least connections white
		greenBlue := uint8(math.Round(255 * (1 - t)))
		color := color.RGBA{255, greenBlue, greenBlue, 255}
		DrawRoundedRect(screen, x, y, width, height, 5, color)
	case ByConnectionStrength:
		// based on connection strength
		connectionStrength := float64(w.connMap[person])

		t := connectionStrength / (20) // max strength is 5, max connections is 30
		if t > 1 {
			t = 1
		} else if t < 0 {
			t = 0
		}
		redGreen := uint8(math.Round(255 * (1 - t)))
		color := color.RGBA{redGreen, redGreen, 255, 255}
		DrawRoundedRect(screen, x, y, width, height, 5, color)
	default:
		DrawRoundedRect(screen, x, y, width, height, 5, color.White)
	}
	DrawText(screen, person.Name, float64(x)+10, float64(y)+5, textFace16, color.Black)
}

func (w *Window) drawConnections(screen *ebiten.Image) {
	if w.solid == nil {
		return
	}

	cb := &connectionBatcher{
		vertices: make([]ebiten.Vertex, 0, 1024),
		indices:  make([]uint16, 0, 1536),
		solid:    w.solid,
		screen:   screen,
	}

	for i := range w.People {
		src := &w.People[i]
		srcRect := w.personRects[src]
		for _, c := range src.Connections {
			tgtRect, ok := w.personRects[c.Person]
			if !ok {
				continue
			}
			w.drawConnectionEdge(cb, srcRect, tgtRect, c.Strength)
		}
	}

	cb.flush()
}
