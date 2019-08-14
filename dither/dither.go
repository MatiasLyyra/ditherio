package dither

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
)

// RGBAVec is container for single pixel values.
// Values are stored as int64 to prevent underflow and overflow.
type RGBAVec struct {
	// TODO: int32 or int would be probably enough
	R int64
	G int64
	B int64
	A int64
}

// ColorFinder specifies who to transform given color to suit the palette
type ColorFinder func(c color.Color) color.Color

// Algo is callback function type for Dither function that applies the actual dithering.
type Algo func(int, int, *image.RGBA, RGBAVec)

// Dither applies given dithering algorithm to the image.
// Quantinization can be controlled by fjnd callback
// that should return closet color available.
func Dither(img image.Image, algo Algo, find ColorFinder) image.Image {
	rect := img.Bounds()
	dithered := image.NewRGBA(rect)
	draw.Draw(dithered, rect, img, rect.Min, draw.Src)

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			oldPixel := dithered.At(x, y)
			newPixel := find(oldPixel)
			dithered.Set(x, y, newPixel)
			diff := colorDiff(oldPixel, newPixel)
			algo(x, y, dithered, diff)
		}
	}
	return dithered
}

// FloydSteinberg applies Floyd-Steinber dithering to image.
// Should not be called directly and should be rather passed Dither function.
func FloydSteinberg(x, y int, img *image.RGBA, err RGBAVec) {
	img.Set(x+1, y, applyWeightsFloydSteinber(img.At(x+1, y), err, 7))
	img.Set(x+1, y+1, applyWeightsFloydSteinber(img.At(x+1, y+1), err, 1))
	img.Set(x, y+1, applyWeightsFloydSteinber(img.At(x, y+1), err, 5))
	img.Set(x-1, y+1, applyWeightsFloydSteinber(img.At(x-1, y+1), err, 3))
}

// Burkes applies Burkes dithering to image.
// Should not be called directly and should be rather passed Dither function.
func Burkes(x, y int, img *image.RGBA, err RGBAVec) {
	img.Set(x+1, y, applyWeightsBurkes(img.At(x+1, y), err, 3))
	img.Set(x+2, y, applyWeightsBurkes(img.At(x+2, y), err, 2))
	img.Set(x-2, y+1, applyWeightsBurkes(img.At(x-2, y+1), err, 1))
	img.Set(x-1, y+1, applyWeightsBurkes(img.At(x-1, y+1), err, 2))
	img.Set(x, y+1, applyWeightsBurkes(img.At(x, y+1), err, 3))
	img.Set(x+1, y+1, applyWeightsBurkes(img.At(x+1, y+1), err, 2))
	img.Set(x+2, y+1, applyWeightsBurkes(img.At(x+2, y+1), err, 1))
}

// BlackAndWhitePalette converts colors to either black or white.
// RGB values are converted into grayscale with:
// gray = (r + g + b) / 3
// If gray is under 0x0FFF it is converted to black and otherwise to white.
func BlackAndWhitePalette(c color.Color) color.Color {
	r, g, b, _ := c.RGBA()
	gray := (uint64(r) + uint64(g) + uint64(b)) / 3
	if gray < 0xFFFF>>1 {
		gray = 0
	} else {
		gray = 0xFFFF
	}
	return color.Gray16{uint16(gray)}
}

// WebSafePalette converts colors to the ones found in palette.WebSafe
func WebSafePalette(c color.Color) color.Color {
	return color.Palette(palette.WebSafe).Convert(c)
}

// Plan9Palette converts colors to the ones found in palette.Plan9
func Plan9Palette(c color.Color) color.Color {
	return color.Palette(palette.Plan9).Convert(c)
}

func clamp(val, min, max int64) int64 {
	if val > max {
		return max
	}
	if val < min {
		return min
	}
	return val
}

func makeColor(r, g, b, a int64) color.Color {
	// Make sure that colors are in the range [0, 0xFFFF]
	r = clamp(r, 0, 0xFFFF)
	g = clamp(g, 0, 0xFFFF)
	b = clamp(b, 0, 0xFFFF)
	a = clamp(a, 0, 0xFFFF)
	col := color.NRGBA64{}
	col.R = uint16(r)
	col.G = uint16(g)
	col.B = uint16(b)
	col.A = uint16(a)
	return col
}
func applyWeightsFloydSteinber(orig color.Color, err RGBAVec, mul int) color.Color {
	oR, oG, oB, oA := orig.RGBA()
	var (
		nR = int64(oR) + (err.R*int64(mul))>>4
		nG = int64(oG) + (err.G*int64(mul))>>4
		nB = int64(oB) + (err.B*int64(mul))>>4
		nA = int64(oA) + (err.A*int64(mul))>>4
	)
	return makeColor(nR, nG, nB, nA)
}

func applyWeightsBurkes(orig color.Color, err RGBAVec, mul uint) color.Color {
	oR, oG, oB, oA := orig.RGBA()
	var (
		nR = int64(oR) + (err.R<<uint64(mul))>>5
		nG = int64(oG) + (err.G<<uint64(mul))>>5
		nB = int64(oB) + (err.B<<uint64(mul))>>5
		nA = int64(oA) + (err.A<<uint64(mul))>>5
	)
	return makeColor(nR, nG, nB, nA)
}

func colorDiff(a, b color.Color) (diff RGBAVec) {
	aR, aG, aB, aA := a.RGBA()
	bR, bG, bB, bA := b.RGBA()
	diff.R = int64(aR) - int64(bR)
	diff.G = int64(aG) - int64(bG)
	diff.B = int64(aB) - int64(bB)
	diff.A = int64(aA) - int64(bA)
	return
}
