package darknettest

import (
	"image"
	"image/color"
	"math"
)

// HLine draws a horizontal line
func HLine(img *image.RGBA, x1, y, x2 int, c color.Color) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y, c)
	}
}

// VLine draws a veritcal line
func VLine(img *image.RGBA, x, y1, y2 int, c color.Color) {
	for ; y1 <= y2; y1++ {
		img.Set(x, y1, c)
	}
}

// Rect draws a rectangle utilizing HLine() and VLine()
func Rect(img *image.RGBA, r image.Rectangle, c color.Color) {
	for x := r.Min.X; x < r.Max.X; x++ {
		VLine(img, x, r.Min.Y, r.Max.Y, c)
	}
	HLine(img, r.Min.X, r.Min.Y, r.Max.X, c)
	HLine(img, r.Min.X, r.Max.Y, r.Max.X, c)
	VLine(img, r.Min.X, r.Min.Y, r.Max.Y, c)
	VLine(img, r.Max.X, r.Min.Y, r.Max.Y, c)
}

// Rect draws a filled rectangle
func RectFill(img *image.RGBA, r image.Rectangle, c color.Color) {
	for x := r.Min.X; x < r.Max.X; x++ {
		VLine(img, x, r.Min.Y, r.Max.Y, c)
	}
}

func CircleFill(img *image.RGBA, point image.Point, radius int, c color.Color) {
	if radius == 0 {
		return
	}
	for d := float64(0); d < 360; d++ {
		x := math.Cos(d) * float64(radius)
		y := math.Sin(d) * float64(radius)
		img.Set(int(x)+point.X, int(y)+point.Y, c)
		img.Set(int(x+1)+point.X, int(y)+point.Y, c)
		img.Set(int(x-1)+point.X, int(y)+point.Y, c)
		img.Set(int(x)+point.X, int(y+1)+point.Y, c)
		img.Set(int(x)+point.X, int(y-1)+point.Y, c)
		img.Set(int(x+1)+point.X, int(y+1)+point.Y, c)
		img.Set(int(x+1)+point.X, int(y-1)+point.Y, c)
		img.Set(int(x-1)+point.X, int(y+1)+point.Y, c)
		img.Set(int(x-1)+point.X, int(y-1)+point.Y, c)
	}
	CircleFill(img, point, radius-1, c)
}
