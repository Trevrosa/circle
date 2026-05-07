package main

import (
	"bytes"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var faceSource *text.GoTextFaceSource

func Init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	faceSource = s
}

const (
	WIDTH  = 1300
	HEIGHT = 800
)

type Window struct {
	People         []Person
	draggingIndex  int // index of person being dragged, -1 if none
	dragOffsetX    float32
	dragOffsetY    float32
	connStartIndex int // index of person where connection is starting, -1 if not connecting
	connStrength   int // 0 = not chosen, 1-5 chosen
}

func (w *Window) Update() error {
	for i := range w.People {
		if w.People[i].Position == [2]float32{0, 0} {
			w.People[i].Position = [2]float32{
				float32(rand.Intn(WIDTH - 100)),
				float32(rand.Intn(HEIGHT - 60)),
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursorX, cursorY := ebiten.CursorPosition()
		for i := len(w.People) - 1; i >= 0; i-- {
			x, y, width, height := personRect(&w.People[i])
			if pointInRect(float32(cursorX), float32(cursorY), x, y, width, height) {
				w.draggingIndex = i
				w.dragOffsetX = float32(cursorX) - x
				w.dragOffsetY = float32(cursorY) - y
				break
			}
		}
	}

	if w.draggingIndex >= 0 {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			cursorX, cursorY := ebiten.CursorPosition()
			w.People[w.draggingIndex].Position = [2]float32{
				float32(cursorX) - w.dragOffsetX,
				float32(cursorY) - w.dragOffsetY,
			}
		} else {
			w.draggingIndex = -1
		}
	}

	// click a person to start -> shows swatches
	// click a swatch to choose strength (1-5)
	// click another person to create the connection
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursorX, cursorY := ebiten.CursorPosition()
		// if we're choosing a strength, check swatches first
		if w.connStartIndex >= 0 && w.connStrength == 0 {
			sx, sy := swatchesPos(&w.People[w.connStartIndex])
			for si := range strengthColors() {
				swX := sx + float32(si*22) - 11
				swY := sy
				if pointInRect(float32(cursorX), float32(cursorY), swX, swY, 20, 20) {
					w.connStrength = si + 1
					break
				}
			}
			// clicking outside cancels
			if w.connStrength == 0 {
				// check if clicked a person to cancel and possibly restart
				for i := len(w.People) - 1; i >= 0; i-- {
					x, y, width, height := personRect(&w.People[i])
					if pointInRect(float32(cursorX), float32(cursorY), x, y, width, height) {
						w.connStartIndex = i
						return nil
					}
				}
				w.connStartIndex = -1
			}
			return nil
		}

		// if strength is chosen, click a person to create connection
		if w.connStartIndex >= 0 && w.connStrength > 0 {
			for i := len(w.People) - 1; i >= 0; i-- {
				x, y, width, height := personRect(&w.People[i])
				if pointInRect(float32(cursorX), float32(cursorY), x, y, width, height) {
					if i != w.connStartIndex {
						// delete connection if already connected
						if w.People[w.connStartIndex].isConnectedTo(&w.People[i]) {
							w.People[w.connStartIndex].unconnect(&w.People[i])
						} else {
							w.People[w.connStartIndex].connect(&w.People[i], w.connStrength)
						} 
					}
					// reset state
					w.connStartIndex = -1
					w.connStrength = 0
					return nil
				}
			}
			// clicked outside, cancel
			w.connStartIndex = -1
			w.connStrength = 0
			return nil
		}

		// start connection: click a person
		for i := len(w.People) - 1; i >= 0; i-- {
			x, y, width, height := personRect(&w.People[i])
			if pointInRect(float32(cursorX), float32(cursorY), x, y, width, height) {
				w.connStartIndex = i
				w.connStrength = 0
				return nil
			}
		}
	}

	return nil
}

func personRect(person *Person) (x, y, width, height float32) {
	textWidth, textHeight := text.Measure(person.Name, textFace(16), 3)
	return person.Position[0], person.Position[1], float32(textWidth + 20), float32(textHeight + 10)
}

func (w *Window) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	
	// draw people rects
	for i := 0; i < len(w.People); i++ {
		person := &w.People[i]
		x, y, width, height := personRect(person)
		FillRoundedRect(screen, x, y, width, height, 5)
		DrawText(screen, person.Name, float64(x)+10, float64(y)+5, 16, color.Black)
	}

	// draw swatches if choosing a connection strength
	if w.connStartIndex >= 0 && w.connStrength == 0 {
		sx, sy := swatchesPos(&w.People[w.connStartIndex])
		cols := strengthColors()
		for i, col := range cols {
			cx := sx + float32(i*22) - 11
			cy := sy
			// draw small rect
			FillRoundedRect(screen, cx, cy, 20, 20, 4)
			// fill with color by drawing a smaller filled rect
			// use vector.FillPath directly for color overlay
			// Draw a filled square using FillRoundedRect with smaller size and color overlay by temporarily calling DrawText trick
			// We'll draw a colored rectangle by drawing a colored filled path
			p := &vector.Path{}
			p.MoveTo(cx, cy)
			p.LineTo(cx+20, cy)
			p.LineTo(cx+20, cy+20)
			p.LineTo(cx, cy+20)
			p.Close()
			opt := vector.DrawPathOptions{AntiAlias: true}
			opt.ColorScale.ScaleWithColor(col)
			vector.FillPath(screen, p, nil, &opt)
		}
	}

	// draw connections (arrows) on top so arrowheads are visible
	for i := 0; i < len(w.People); i++ {
		src := &w.People[i]
		sx, sy, sw, sh := personRect(src)
		srcCX := float64(sx + sw/2)
		srcCY := float64(sy + sh/2)
		for _, c := range src.Connections {
			ti := findPersonIndexByName(w.People, c.Person)
			if ti < 0 || ti >= len(w.People) {
				continue
			}
			tgt := &w.People[ti]
			tx, ty, tw, th := personRect(tgt)
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
}

func pointInRect(px, py, x, y, width, height float32) bool {
	return px >= x && px <= x+width && py >= y && py <= y+height
}

func findPersonIndexByName(people []Person, name string) int {
	for i := range people {
		if people[i].Name == name {
			return i
		}
	}
	return -1
}

func strengthColors() []color.Color {
	return []color.Color{
		color.RGBA{0xFF, 0x33, 0x33, 0xFF}, // red
		color.RGBA{0xFF, 0x99, 0x99, 0xFF}, // pink
		color.RGBA{0x66, 0x99, 0xFF, 0xFF}, // blue
		color.RGBA{0xFF, 0x99, 0x33, 0xFF}, // orange
		color.RGBA{0x80, 0x80, 0x80, 0xFF}, // grey
		color.Black, // black
	}
}

func strengthColor(str int) color.Color {
	cols := strengthColors()
	if str <= 0 || str > len(cols) {
		return color.Black
	}
	return cols[str-1]
}

func swatchesPos(person *Person) (float32, float32) {
	x, y, w, _ := personRect(person)
	// position swatches centered above the person rect
	totalW := float32(5*20 + 4*2)
	sx := x + w/2 - totalW/2
	sy := y - 26
	return sx, sy
}

func (w *Window) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WIDTH, HEIGHT
}
