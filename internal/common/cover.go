package common

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strings"
)

func DecodeCover(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

func bilinearSample(img image.Image, fx, fy float64) (r, g, b uint8) {
	bounds := img.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())
	x := fx * (w - 1)
	y := fy * (h - 1)
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= w-1 {
		x = w - 2
	}
	if y >= h-1 {
		y = h - 2
	}
	ix, iy := int(math.Floor(x)), int(math.Floor(y))
	fx2, fy2 := x-float64(ix), y-float64(iy)

	c00 := colorAt(img, ix, iy)
	c10 := colorAt(img, ix+1, iy)
	c01 := colorAt(img, ix, iy+1)
	c11 := colorAt(img, ix+1, iy+1)

	r = uint8((1-fx2)*(1-fy2)*float64(c00.R) + fx2*(1-fy2)*float64(c10.R) + (1-fx2)*fy2*float64(c01.R) + fx2*fy2*float64(c11.R))
	g = uint8((1-fx2)*(1-fy2)*float64(c00.G) + fx2*(1-fy2)*float64(c10.G) + (1-fx2)*fy2*float64(c01.G) + fx2*fy2*float64(c11.G))
	b = uint8((1-fx2)*(1-fy2)*float64(c00.B) + fx2*(1-fy2)*float64(c10.B) + (1-fx2)*fy2*float64(c01.B) + fx2*fy2*float64(c11.B))
	return
}

type rgb struct{ R, G, B uint8 }

func colorAt(img image.Image, x, y int) rgb {
	r, g, b, _ := img.At(x, y).RGBA()
	return rgb{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
}

func scaleBilinear(img image.Image, w, h int) [][]rgb {
	pixels := make([][]rgb, h)
	for y := 0; y < h; y++ {
		pixels[y] = make([]rgb, w)
		for x := 0; x < w; x++ {
			fx := float64(x) / float64(w)
			fy := float64(y) / float64(h)
			r, g, b := bilinearSample(img, fx, fy)
			pixels[y][x] = rgb{r, g, b}
		}
	}
	return pixels
}

func RenderCover(data []byte, charW, charH int) string {
	img, err := DecodeCover(data)
	if err != nil {
		return ""
	}

	pixW := charW
	pixH := charH * 2
	pixels := scaleBilinear(img, pixW, pixH)

	var lines []string
	for y := 0; y < charH; y++ {
		var sb strings.Builder
		for x := 0; x < charW; x++ {
			top := pixels[y*2][x]
			bottom := pixels[y*2+1][x]
			sb.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm\u2584\033[0m",
				bottom.R, bottom.G, bottom.B,
				top.R, top.G, top.B,
			))
		}
		lines = append(lines, sb.String())
	}
	return strings.Join(lines, "\n")
}
