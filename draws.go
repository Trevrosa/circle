package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func (w *Window) drawPerson(screen *ebiten.Image, i int) {
	person := &w.People[i]
	x, y, width, height := w.personRect(person)

	switch w.pageIndex {
	case 3:
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
		FillRoundedRect(screen, x, y, width, height, 5, color)
	case 4:
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
		FillRoundedRect(screen, x, y, width, height, 5, color)
	default:
		FillRoundedRect(screen, x, y, width, height, 5, color.White)
	}
	DrawText(screen, person.Name, float64(x)+10, float64(y)+5, textFace16, color.Black)
}

func (w *Window) drawConnections(screen *ebiten.Image, i int) {
	src := &w.People[i]
	sx, sy, sw, sh := w.personRect(src)
	srcCX := float64(sx + sw/2)
	srcCY := float64(sy + sh/2)
	for _, c := range src.Connections {
		tgt := c.Person
		tx, ty, tw, th := w.personRect(tgt)
		tgtCX := float64(tx + tw/2)
		tgtCY := float64(ty + th/2)
		angle := math.Atan2(tgtCY-srcCY, tgtCX-srcCX)
		// compute edge points
		startXf, startYf := EdgePointFromCenter(srcCX, srcCY, float64(sw)/2, float64(sh)/2, angle, true)
		endXf, endYf := EdgePointFromCenter(tgtCX, tgtCY, float64(tw)/2, float64(th)/2, angle, false)
		col := strengthColor(c.Strength)
		DrawArrow(screen, startXf, startYf, endXf, endYf, col)
	}
}
