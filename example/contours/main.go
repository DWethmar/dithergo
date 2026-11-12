// A topographic map as a background, with a colour for every height.
package main

import (
	"image/color"
	"log"

	"github.com/dwethmar/dithergo"
	"github.com/hajimehoshi/ebiten/v2"
)

// The paper first, then the heights from the valleys up to the peaks.
var palette = []color.Color{
	color.RGBA{0x08, 0x0a, 0x12, 0xff},
	color.RGBA{0x4a, 0x5d, 0xd9, 0xff}, // blue
	color.RGBA{0x2f, 0xb5, 0xc4, 0xff}, // teal
	color.RGBA{0x4d, 0xe0, 0x8a, 0xff}, // green
	color.RGBA{0xf2, 0xc1, 0x4e, 0xff}, // amber
	color.RGBA{0xff, 0x6e, 0xb4, 0xff}, // pink
}

type game struct{ ticks int }

func (g *game) Update() error {
	g.ticks++
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	b := screen.Bounds()
	seconds := float64(g.ticks) / float64(ebiten.TPS())
	dithergo.Contours(screen, b.Dx(), b.Dy(), seconds, &dithergo.Options{Matrix: dithergo.Bayer8, Palette: palette})
}

func (g *game) Layout(w, h int) (int, int) { return w, h }

func main() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("dithergo: contours")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(&game{}); err != nil {
		log.Fatal(err)
	}
}
