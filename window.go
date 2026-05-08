package main

import (
	"bytes"
	"image/color"
	"log"
	"math/rand"
	"strconv"

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
	pageIndex      int
	draggingIndex  int // index of person being dragged, -1 if none
	dragOffsetX    float32
	dragOffsetY    float32
	connStartIndex int // index of person where connection is starting, -1 if not connecting
	connStrength   int
	connMap        map[*Person]float32 // map of the total connections (including from others) for each person
}

func NewWindow(people []Person) *Window {
	Init()
	return &Window{
		People:         people,
		draggingIndex:  -1,
		connStartIndex: -1,
		connMap:        initConnMap(people),
	}
}

func initConnMap(people []Person) map[*Person]float32 {
	connMap := make(map[*Person]float32, len(people))
	for pi := range people {
		p := &people[pi]
		for _, c := range p.Connections {
			connMap[p] += colorStrength(c.Strength)
			connMap[c.Person] += colorStrength(c.Strength)
		}
	}
	return connMap
}

func (w *Window) personPosition(i int) *[2]float32 {
	return &w.People[i].Positions[w.pageIndex]
}

func (w *Window) connect(p *Person, other *Person) {
	strength := w.connStrength
	p.connect(other, strength)
	w.connMap[p] += colorStrength(strength)
	w.connMap[other] += colorStrength(strength)
}

func (w *Window) unconnect(p *Person, other *Person) {
	strength := 100 // colorStrength will return 0
	for _, c := range p.Connections {
		if c.Person == other {
			strength = c.Strength
			break
		}
	}
	p.unconnect(other)
	w.connMap[p] -= colorStrength(strength)
	w.connMap[other] -= colorStrength(strength)
}

func (w *Window) Update() error {
	// move randomly if at 0,0
	for i := range w.People {
		if *w.personPosition(i) == [2]float32{0, 0} {
			*w.personPosition(i) = [2]float32{
				float32(rand.Intn(WIDTH - 100)),
				float32(rand.Intn(HEIGHT - 60)),
			}
		}
	}

	// dragging
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursorX, cursorY := ebiten.CursorPosition()
		for i := range w.People {
			x, y, width, height := w.personRect(&w.People[i])
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
			*w.personPosition(w.draggingIndex) = [2]float32{
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
			sx, sy := w.swatchesPos(&w.People[w.connStartIndex])
			for si := range strengthColors() {
				swX := swatchWidth(int(sx), si)
				swY := sy
				if pointInRect(float32(cursorX), float32(cursorY), swX, swY, 20, 20) {
					w.connStrength = si + 1
					break
				}
			}
			// clicking outside cancels
			if w.connStrength == 0 {
				// check if clicked a person to cancel and possibly restart
				for i := range w.People {
					x, y, width, height := w.personRect(&w.People[i])
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
			for i := range w.People {
				x, y, width, height := w.personRect(&w.People[i])
				if pointInRect(float32(cursorX), float32(cursorY), x, y, width, height) {
					if i != w.connStartIndex {
						person := &w.People[w.connStartIndex]
						other := &w.People[i]
						// delete connection if already connected
						if person.isConnectedTo(other) {
							w.unconnect(person, other)
						} else {
							w.connect(person, other)
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
		for i := range w.People {
			x, y, width, height := w.personRect(&w.People[i])
			if pointInRect(float32(cursorX), float32(cursorY), x, y, width, height) {
				w.connStartIndex = i
				w.connStrength = 0
				return nil
			}
		}
	}

	// switching pages
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		for i := range PAGES {
			x, y, wi, h := i*20+2, 2+1, 18, 18
			if pointInRect(float32(cx), float32(cy), float32(x), float32(y), float32(wi), float32(h)) {
				w.pageIndex = i

				for i := range w.People {
					person := &w.People[i]
					for range w.pageIndex - len(person.Positions) + 1 {
						person.Positions = append(person.Positions, [2]float32{0, 0})
					}
					if PAGES[w.pageIndex] != nil {
						person.Positions[w.pageIndex] = PAGES[w.pageIndex](person, i)
					}
				}
				break
			}
		}
	}

	return nil
}

// starts from the left of the rect, rect fits 3, center so only offset by half
func swatchWidth(w, si int) float32 {
	return float32(w) + float32(si*22) - ((float32(len(strengthColors())-3) / 2) * 11)
}

func (win *Window) swatchesPos(person *Person) (float32, float32) {
	x, y, w, _ := win.personRect(person)
	// position swatches centered above the person rect
	totalW := float32(5*20 + 4*2)
	sx := x + w/2 - totalW/2
	sy := y - 26
	return sx, sy
}

func (w *Window) personRect(person *Person) (x, y, width, height float32) {
	textWidth, textHeight := text.Measure(person.Name, textFace(16), 3)
	return person.Positions[w.pageIndex][0], person.Positions[w.pageIndex][1], float32(textWidth + 20), float32(textHeight + 10)
}

func (w *Window) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)

	// draw people rects
	for i := range w.People {
		w.drawPerson(screen, i)
	}

	// draw swatches if choosing a connection strength
	if w.connStartIndex >= 0 && w.connStrength == 0 {
		sx, sy := w.swatchesPos(&w.People[w.connStartIndex])
		cols := strengthColors()
		for i, col := range cols {
			cx := swatchWidth(int(sx), i)
			cy := sy
			// draw small rect
			FillRoundedRect(screen, cx, cy, 20, 20, 4, col)
		}
	}

	// draw connections (arrows) on top so arrowheads are visible
	for i := range w.People {
		w.drawConnections(screen, i)
	}

	// draw page selector
	for i := range PAGES {
		x := float32(i)*20 + 1
		vector.StrokeRect(screen, x, 2, 20, 20, 2, color.Black, true)
		if i == w.pageIndex {
			vector.FillRect(screen, x+1, 2+1, 18, 18, color.RGBA{0x0, 0xFF, 0xFF, 0x88}, true)
		}
		DrawText(screen, strconv.Itoa(i), float64(x)+5, 1, 16, color.Black)
	}
}

func pointInRect(px, py, x, y, width, height float32) bool {
	return px >= x && px <= x+width && py >= y && py <= y+height
}

func strengthColors() []color.Color {
	return []color.Color{
		color.RGBA{0xFF, 0x33, 0x33, 0xFF}, // red
		color.RGBA{0xFF, 0x99, 0x99, 0xFF}, // pink
		color.RGBA{0x66, 0x99, 0xFF, 0xFF}, // blue
		color.RGBA{0xFF, 0x99, 0x33, 0xFF}, // orange
		color.RGBA{0x80, 0x80, 0x80, 0xFF}, // grey
		color.Black,
		color.RGBA{0x80, 0x00, 0x80, 0xFF}, // purple
	}
}

func colorStrength(str int) float32 {
	if str == 1 {
		return 3
	} else if str <= 4 {
		return 5 - float32(str)
	} else if str == 5 {
		return 0.5
	} else if str == 6 {
		return -1
	} else {
		return 0
	}
}

func strengthColor(str int) color.Color {
	cols := strengthColors()
	if str <= 0 || str > len(cols) {
		return color.Black
	}
	return cols[str-1]
}

func (w *Window) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WIDTH, HEIGHT
}
