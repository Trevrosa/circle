package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (w *Window) checkConnectionInputs() {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	cursorX, cursorY := ebiten.CursorPosition()

	// from a previous click, person and strength were both chosen
	// so user must be trying to make a connection
	if w.connStartIndex >= 0 && w.connStrength > 0 {
		// check if clicked a person to connect to
		for i := range w.People {
			x, y, width, height := w.personRect(&w.People[i])
			if PointInRect(cursorX, cursorY, x, y, width, height) {
				// cannot connect to self
				if i != w.connStartIndex {
					person := &w.People[w.connStartIndex]
					other := &w.People[i]
					// dirty is set by the connect/unconnect functions
					// delete connection if already connected
					if person.isConnectedTo(other) {
						w.unconnect(person, other)
					} else {
						w.connect(person, other)
					}
				}
				break
			}
		}

		// either a connection was made or cancelled, for both cases we want to reset
		w.connStartIndex = -1
		w.connStrength = 0
	}

	// either no person chosen or no strength chosen now

	// check if clicked a person
	for i := range w.People {
		x, y, width, height := w.personRect(&w.People[i])
		if PointInRect(cursorX, cursorY, x, y, width, height) {
			w.connStartIndex = i
			break
		}
	}

	// a person was clicked, so user must be selecting strength now
	if w.connStartIndex >= 0 {
		// must be dirty because all clicks will change something
		w.dirty = true

		// check if clicked a strength swatch to set the strength
		sx, sy := w.swatchesPos(&w.People[w.connStartIndex])
		for si := range strengthColors() {
			swX := swatchWidth(int(sx), si)
			swY := sy
			if PointInRect(cursorX, cursorY, swX, swY, 20, 20) {
				w.connStrength = si + 1
				return
			}
		}

		// check if clicked a person to immediately switch to that person
		for i := range w.People {
			x, y, width, height := w.personRect(&w.People[i])
			if PointInRect(cursorX, cursorY, x, y, width, height) {
				w.connStartIndex = i
				return
			}
		}

		// clicked somewhere else, so reset
		w.connStartIndex = -1
	}
}
