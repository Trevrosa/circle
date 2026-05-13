package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type personRectInfo struct {
	x, y          float32
	width, height float32
	centerX       float32
	centerY       float32
	halfWidth     float32
	halfHeight    float32
}

// colorWithAlpha replaces the alpha channel of a color while preserving RGB.
// RGBA() returns 16-bit values, so we shift right 8 bits to get 8-bit channels.
func colorWithAlpha(col color.Color, alpha uint8) color.Color {
	r, g, b, _ := col.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), alpha}
}

type connectionBatcher struct {
	vertices []ebiten.Vertex
	indices  []uint16
	solid    *ebiten.Image
	screen   *ebiten.Image
}

func (w *Window) drawConnectionEdge(cb *connectionBatcher, srcRect, tgtRect personRectInfo, strength int) {
	const maxVerticesPerBatch = 65000

	srcCX := float64(srcRect.centerX)
	srcCY := float64(srcRect.centerY)
	tgtCX := float64(tgtRect.centerX)
	tgtCY := float64(tgtRect.centerY)
	angle := math.Atan2(tgtCY-srcCY, tgtCX-srcCX)

	startXf, startYf := EdgePointFromCenter(srcCX, srcCY, float64(srcRect.halfWidth), float64(srcRect.halfHeight), angle, true)
	endXf, endYf := EdgePointFromCenter(tgtCX, tgtCY, float64(tgtRect.halfWidth), float64(tgtRect.halfHeight), angle, false)

	col := strengthColor(strength)

	if len(cb.vertices)+15 > maxVerticesPerBatch {
		cb.flush()
	}

	cb.appendFeatheredLine(startXf, startYf, endXf, endYf, col)
	drawArrowhead(cb, endXf, endYf, startXf, startYf, col)

	if len(cb.vertices) >= maxVerticesPerBatch-16 {
		cb.flush()
	}
}

// drawArrowhead creates a triangle pointing from start to end.
func drawArrowhead(cb *connectionBatcher, endX, endY, startX, startY float32, col color.Color) {
	dx := endX - startX
	dy := endY - startY
	length := float32(math.Hypot(float64(dx), float64(dy)))
	if length == 0 {
		return
	}

	// normalize direction vector along the line
	ux := dx / length
	uy := dy / length
	// perpendicular vector
	px := -uy
	py := ux

	// point at end, base back along the direction
	size := float32(10)
	baseX := endX - ux*size
	baseY := endY - uy*size
	cb.appendTriangle(
		endX, endY,
		baseX+px*size*0.5, baseY+py*size*0.5,
		baseX-px*size*0.5, baseY-py*size*0.5,
		col,
	)
}

func (cb *connectionBatcher) flush() {
	if len(cb.indices) == 0 {
		return
	}
	op := &ebiten.DrawTrianglesOptions{ColorScaleMode: ebiten.ColorScaleModeStraightAlpha}
	cb.screen.DrawTriangles(cb.vertices, cb.indices, cb.solid, op)
	cb.vertices = cb.vertices[:0]
	cb.indices = cb.indices[:0]
}

func (cb *connectionBatcher) appendVertex(x, y float32, col color.Color) uint16 {
	r, g, b, a := col.RGBA()
	cb.vertices = append(cb.vertices, ebiten.Vertex{
		DstX:   x,
		DstY:   y,
		SrcX:   0,
		SrcY:   0,
		ColorR: float32(r) / 0xffff,
		ColorG: float32(g) / 0xffff,
		ColorB: float32(b) / 0xffff,
		ColorA: float32(a) / 0xffff,
	})
	return uint16(len(cb.vertices) - 1)
}

func (cb *connectionBatcher) appendQuad(x0, y0 float32, c0 color.Color, x1, y1 float32, c1 color.Color, x2, y2 float32, c2 color.Color, x3, y3 float32, c3 color.Color) {
	base := uint16(len(cb.vertices))
	cb.appendVertex(x0, y0, c0)
	cb.appendVertex(x1, y1, c1)
	cb.appendVertex(x2, y2, c2)
	cb.appendVertex(x3, y3, c3)
	cb.indices = append(cb.indices, base, base+1, base+2, base, base+2, base+3)
}

func (cb *connectionBatcher) appendTriangle(x0, y0, x1, y1, x2, y2 float32, col color.Color) {
	base := uint16(len(cb.vertices))
	cb.appendVertex(x0, y0, col)
	cb.appendVertex(x1, y1, col)
	cb.appendVertex(x2, y2, col)
	cb.indices = append(cb.indices, base, base+1, base+2)
}

// appendFeatheredLine renders a line with soft anti-aliased edges using three bands:
// a transparent outer band, a solid core, and transparent outer band on the opposite side.
func (cb *connectionBatcher) appendFeatheredLine(startXf, startYf, endXf, endYf float32, col color.Color) {
	dx := endXf - startXf
	dy := endYf - startYf
	length := float32(math.Hypot(float64(dx), float64(dy)))
	if length == 0 {
		return
	}

	// normalized direction and perpendicular vectors
	ux := dx / length
	uy := dy / length
	px := -uy
	py := ux

	coreHalfWidth := float32(0.9)
	featherWidth := float32(0.6)
	outerHalfWidth := coreHalfWidth + featherWidth

	opaque := colorWithAlpha(col, 0xff)
	transparent := colorWithAlpha(col, 0x00)

	// left feather band
	cb.appendQuad(
		startXf-px*outerHalfWidth, startYf-py*outerHalfWidth, transparent,
		startXf-px*coreHalfWidth, startYf-py*coreHalfWidth, opaque,
		endXf-px*coreHalfWidth, endYf-py*coreHalfWidth, opaque,
		endXf-px*outerHalfWidth, endYf-py*outerHalfWidth, transparent,
	)

	// center
	cb.appendQuad(
		startXf-px*coreHalfWidth, startYf-py*coreHalfWidth, opaque,
		startXf+px*coreHalfWidth, startYf+py*coreHalfWidth, opaque,
		endXf+px*coreHalfWidth, endYf+py*coreHalfWidth, opaque,
		endXf-px*coreHalfWidth, endYf-py*coreHalfWidth, opaque,
	)

	// right feather band
	cb.appendQuad(
		startXf+px*coreHalfWidth, startYf+py*coreHalfWidth, opaque,
		startXf+px*outerHalfWidth, startYf+py*outerHalfWidth, transparent,
		endXf+px*outerHalfWidth, endYf+py*outerHalfWidth, transparent,
		endXf+px*coreHalfWidth, endYf+py*coreHalfWidth, opaque,
	)
}
