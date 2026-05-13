package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var GreyFill = color.RGBA{200, 200, 200, 255}

func DrawRoundedRect(dst *ebiten.Image, x, y, width, height, radius float32, fill color.Color) {
	path := &vector.Path{}

	// top left
	path.MoveTo(x+radius, y)
	// top edge
	path.LineTo(x+width-radius, y)
	// top right
	path.QuadTo(x+width, y, x+width, y+radius)
	// right edge
	path.LineTo(x+width, y+height-radius)
	// bottom right
	path.QuadTo(x+width, y+height, x+width-radius, y+height)
	// bottom edge
	path.LineTo(x+radius, y+height)
	// bottom left
	path.QuadTo(x, y+height, x, y+height-radius)
	// left edge
	path.LineTo(x, y+radius)
	// top left
	path.QuadTo(x, y, x+radius, y)

	fillOptions := vector.DrawPathOptions{AntiAlias: false}
	fillOptions.ColorScale.ScaleWithColor(fill)
	vector.FillPath(dst, path, nil, &fillOptions)

	strokeOptions := vector.StrokeOptions{Width: 2}
	strokeDrawOptions := vector.DrawPathOptions{AntiAlias: true}
	strokeDrawOptions.ColorScale.ScaleWithColor(color.RGBA{100, 100, 100, 255})
	vector.StrokePath(dst, path, &strokeOptions, &strokeDrawOptions)
}

func DrawArrow(dst *ebiten.Image, sx, sy, ex, ey float32, col color.Color) {
	vector.StrokeLine(dst, sx, sy, ex, ey, 3, col, true)

	// arrowhead
	ax := float64(ex)
	ay := float64(ey)
	angle := math.Atan2(float64(ey-sy), float64(ex-sx))
	size := 10.0
	// two base points
	bx := ax - math.Cos(angle)*size
	by := ay - math.Sin(angle)*size
	leftX := bx + math.Sin(angle)*size*0.5
	leftY := by - math.Cos(angle)*size*0.5
	rightX := bx - math.Sin(angle)*size*0.5
	rightY := by + math.Cos(angle)*size*0.5

	ph := &vector.Path{}
	ph.MoveTo(float32(ax), float32(ay))
	ph.LineTo(float32(leftX), float32(leftY))
	ph.LineTo(float32(rightX), float32(rightY))

	opt2 := vector.DrawPathOptions{AntiAlias: true}
	opt2.ColorScale.ScaleWithColor(col)
	vector.FillPath(dst, ph, nil, &opt2)
}

// equilateral triangle pointing down with center at (x, y)
func DrawTriangle(dst *ebiten.Image, x, y, length float32, col color.Color) {
	path := &vector.Path{}
	height := length * float32(math.Sqrt(3)) / 2
	topY := y + (2*height)/3
	baseY := y - height/3
	leftX := x - length/2
	rightX := x + length/2

	path.MoveTo(x, topY)
	path.LineTo(leftX, baseY)
	path.LineTo(rightX, baseY)
	path.Close()

	fillOptions := vector.DrawPathOptions{AntiAlias: false}
	fillOptions.ColorScale.ScaleWithColor(col)
	vector.FillPath(dst, path, nil, &fillOptions)

	// strokeOptions := vector.StrokeOptions{Width: 2}
	// strokeDrawOptions := vector.DrawPathOptions{AntiAlias: true}
	// strokeDrawOptions.ColorScale.ScaleWithColor(col)
	// vector.StrokePath(dst, path, &strokeOptions, &strokeDrawOptions)
}

func EdgePointFromCenter(cx, cy, hw, hh, angle float64, outward bool) (float32, float32) {
	ca := math.Cos(angle)
	sa := math.Sin(angle)
	// avoid division by zero
	absCa := math.Abs(ca)
	absSa := math.Abs(sa)
	var radius float64
	if absCa < 1e-6 {
		radius = hh / absSa
	} else if absSa < 1e-6 {
		radius = hw / absCa
	} else {
		rx := hw / absCa
		ry := hh / absSa
		if rx < ry {
			radius = rx
		} else {
			radius = ry
		}
	}
	if outward {
		return float32(cx + ca*radius), float32(cy + sa*radius)
	}
	return float32(cx - ca*radius), float32(cy - sa*radius)
}

func PointInRect(px, py int, x, y, width, height float32) bool {
	pxf, pyf := float32(px), float32(py)
	return pxf >= x && pxf <= x+width && pyf >= y && pyf <= y+height
}

func DrawText(dst *ebiten.Image, str string, x, y float64, textFace text.Face, color color.Color) {
	op := &text.DrawOptions{}
	op.ColorScale.ScaleWithColor(color)
	op.GeoM.Translate(x, y)
	text.Draw(dst, str, textFace, op)
}
