package main

import "math"

var PAGES = [](func(*Person, int) [2]float32){
	nil,
	page1,
	page2,
	nil,
	nil,
}

var pagePeopleCount int = 0

func page1(person *Person, i int) [2]float32 {
	n := pagePeopleCount
	centerX := float64(WIDTH) / 2
	centerY := float64(HEIGHT) / 2
	radius := math.Min(float64(WIDTH), float64(HEIGHT)) * 0.399
	angle := (2 * math.Pi * float64(i) / float64(n)) - (math.Pi / 2)

	x := centerX + radius*math.Cos(angle) - 50
	y := centerY + radius*math.Sin(angle) - 16
	return [2]float32{float32(x), float32(y)}
}

func page2(person *Person, i int) [2]float32 {
	n := pagePeopleCount
	centerX := float64(WIDTH) / 2
	centerY := float64(HEIGHT) / 2

	maxRadius := math.Min(float64(WIDTH), float64(HEIGHT)) * 0.5
	radius := 0.0
	if n > 1 {
		radius = maxRadius * math.Sqrt(float64(i)/float64(n-1))
	}

	goldenAngle := math.Pi * (3 - math.Sqrt(5))
	angle := float64(i) * goldenAngle

	x := centerX + radius*math.Cos(angle) - 50
	y := centerY + radius*math.Sin(angle) - 16
	return [2]float32{float32(x), float32(y)}
}
