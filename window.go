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
var textFace16 *text.GoTextFace

func Init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	faceSource = s

	textFace16 = &text.GoTextFace{
		Source: faceSource,
		Size:   16,
	}
}

const (
	WIDTH  = 1300
	HEIGHT = 800
)

type ColorMode int

const (
	Default ColorMode = iota
	ByNumConnections
	ByConnectionStrength
	NumColorModes
)

func (cm ColorMode) String() string {
	switch cm {
	case ByNumConnections:
		return "ByNumConnections"
	case ByConnectionStrength:
		return "ByConnectionStrength"
	case NumColorModes:
		panic("invalid color mode")
	default:
		return "Default"
	}
}

type Window struct {
	People             []Person
	pageIndex          int
	personDragged      int // index of person being dragged, -1 if none
	dragOffsetX        float32
	dragOffsetY        float32
	connStartIndex     int                 // index of person where connection is starting, -1 if not connecting
	connStrength       int                 // stored starting from 1, 0 means not chosen
	connMap            map[*Person]float32 // map of the total connections (including from others) for each person
	dirty              bool                // whether we need to render
	colorMode          ColorMode
	switchingColorMode bool
}

func NewWindow(people []Person) *Window {
	pagePeopleCount = len(people)
	Init()
	return &Window{
		People:         people,
		personDragged:  -1,
		connStartIndex: -1,
		connMap:        initConnMap(people),
		dirty:          true,
		colorMode:      Default,
	}
}

func initConnMap(people []Person) map[*Person]float32 {
	connMap := make(map[*Person]float32, len(people))
	for pi := range people {
		p := &people[pi]
		for _, c := range p.Connections {
			connMap[p] += strengthValue(c.Strength)
			connMap[c.Person] += strengthValue(c.Strength)
		}
	}
	return connMap
}

func (w *Window) personPosition(i int) *[2]float32 {
	return &w.People[i].Positions[w.pageIndex]
}

func (w *Window) connect(p *Person, other *Person) {
	w.dirty = true

	strength := w.connStrength
	p.connect(other, strength)
	w.connMap[p] += strengthValue(strength)
	w.connMap[other] += strengthValue(strength)
}

func (w *Window) unconnect(p *Person, other *Person) {
	w.dirty = true

	strength := 100 // strengthValue will return 0 for out of range
	for _, c := range p.Connections {
		if c.Person == other {
			strength = c.Strength
			break
		}
	}
	p.unconnect(other)
	w.connMap[p] -= strengthValue(strength)
	w.connMap[other] -= strengthValue(strength)
}

func (w *Window) Update() error {
	// move randomly if at 0,0
	for i := range w.People {
		if *w.personPosition(i) == [2]float32{0, 0} {
			w.dirty = true
			*w.personPosition(i) = [2]float32{
				float32(rand.Intn(WIDTH - 100)),
				float32(rand.Intn(HEIGHT - 60)),
			}
		}
	}

	// dragging - detect first click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursorX, cursorY := ebiten.CursorPosition()
		for i := range w.People {
			x, y, width, height := w.personRect(&w.People[i])
			if PointInRect(cursorX, cursorY, x, y, width, height) {
				w.personDragged = i
				w.dragOffsetX = float32(cursorX) - x
				w.dragOffsetY = float32(cursorY) - y
				break
			}
		}
	}
	// detect mousedown to drag
	if w.personDragged >= 0 {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			w.dirty = true
			cursorX, cursorY := ebiten.CursorPosition()
			*w.personPosition(w.personDragged) = [2]float32{
				float32(cursorX) - w.dragOffsetX,
				float32(cursorY) - w.dragOffsetY,
			}
		} else {
			w.personDragged = -1
		}
	}

	// making connections
	w.checkConnectionInputs()

	// switching pages
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		for i := range PAGES {
			x, y, wi, h := i*20+2, 2+1, 18, 18
			if PointInRect(cx, cy, float32(x), float32(y), float32(wi), float32(h)) {
				w.dirty = true
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

	// switching color modes
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		_modeTextW, _modeTextH := text.Measure(w.colorMode.String(), textFace16, 3)
		modeTextW, modeTextH := float32(_modeTextW), float32(_modeTextH)
		// +10 padding, +10 trangle
		if PointInRect(cx, cy, 0, 25, modeTextW+10+10, modeTextH) {
			w.dirty = true
			w.switchingColorMode = !w.switchingColorMode
		} else {
			pos := float32(0)
			for i := range NumColorModes {
				if i == w.colorMode {
					continue
				}
				pos++
				if PointInRect(cx, cy, 0, 25+pos*modeTextH, modeTextW+10+10, modeTextH) {
					w.dirty = true
					w.colorMode = i
					w.switchingColorMode = false
					break
				}
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
	textWidth, textHeight := text.Measure(person.Name, textFace16, 3)
	return person.Positions[w.pageIndex][0], person.Positions[w.pageIndex][1], float32(textWidth + 20), float32(textHeight + 10)
}

func (w *Window) Draw(screen *ebiten.Image) {
	if !w.dirty {
		return
	}
	w.dirty = false

	screen.Fill(color.White)

	// draw people rects
	for i := range w.People {
		w.drawPerson(screen, i)
	}

	// draw connection arrows on top of rects
	for i := range w.People {
		w.drawConnections(screen, i)
	}

	// draw swatches if choosing a connection strength
	if w.connStartIndex >= 0 && w.connStrength == 0 {
		sx, sy := w.swatchesPos(&w.People[w.connStartIndex])
		cols := strengthColors()
		for i, col := range cols {
			cx := swatchWidth(int(sx), i)
			cy := sy
			// draw small rect
			DrawRoundedRect(screen, cx, cy, 20, 20, 4, col)
		}
	}

	// draw page selector
	for i := range PAGES {
		x := float32(i)*20 + 1
		vector.StrokeRect(screen, x, 2, 20, 20, 2, color.Black, true)
		if i == w.pageIndex {
			vector.FillRect(screen, x+1, 2+1, 18, 18, color.RGBA{0x0, 0xFF, 0xFF, 0x88}, true)
		}
		DrawText(screen, strconv.Itoa(i), float64(x)+5, 1, textFace16, color.Black)
	}

	// draw color mode selector
	modeTextW, modeTextH := text.Measure(w.colorMode.String(), textFace16, 3)
	vector.FillRect(screen, 0, 25, float32(modeTextW)+10+10, float32(modeTextH), color.RGBA{60, 200, 10, 180}, false)
	DrawTriangle(screen, float32(modeTextW)+10, 25+float32(modeTextH)/2, 10, color.Black)
	DrawText(screen, w.colorMode.String(), 0, 25, textFace16, color.Black)
	if w.switchingColorMode {
		pos := 0.0
		for i := range NumColorModes {
			if i == w.colorMode {
				continue
			}
			pos++
			DrawText(screen, ColorMode(i).String(), 0, 25+pos*modeTextH, textFace16, color.Black)
		}
	}
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

func strengthValue(str int) float32 {
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
