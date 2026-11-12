// A two-colour aurora behind two buttons that are filled with an aurora of their own.
package main

import (
	"image"
	"image/color"
	"log"

	"github.com/dwethmar/dithergo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	night  = color.RGBA{0x06, 0x08, 0x14, 0xff}
	green  = color.RGBA{0x4d, 0xe0, 0x8a, 0xff}
	violet = color.RGBA{0x8a, 0x5c, 0xe6, 0xff}
	blue   = color.RGBA{0x7c, 0x8c, 0xff, 0xff}
	pink   = color.RGBA{0xff, 0x6e, 0xb4, 0xff}

	sky     = dithergo.AuroraPalette(night, violet, green) // paper, the rays' tips, the bright edges
	resting = dithergo.AuroraPalette(color.Black, violet, blue)
	hovered = dithergo.AuroraPalette(color.Black, blue, pink)
)

type game struct {
	ticks   int
	buttons []image.Rectangle
}

func (g *game) Update() error {
	g.ticks++
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	seconds := float64(g.ticks) / float64(ebiten.TPS())
	b := screen.Bounds()

	// The background: the whole screen is the rectangle.
	dithergo.Aurora(screen, b.Dx(), b.Dy(), seconds, &dithergo.Options{Matrix: dithergo.Bayer8, Palette: sky})

	// A button: the same call with the button's size, moved to its place with GeoM. Draw
	// (a dithered image) is placed the same way.
	for i, r := range g.buttons {
		opts := &dithergo.Options{Cell: 2, Matrix: dithergo.Bayer8, Palette: resting}
		if image.Pt(ebiten.CursorPosition()).In(r) {
			opts.Palette = hovered
		}
		opts.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
		// Each button a different stretch of time, or they would all show the same lights.
		dithergo.Aurora(screen, r.Dx(), r.Dy(), seconds+float64(i)*40, opts)

		vector.StrokeRect(screen, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), 1, color.White, false)
		label := []string{"New game", "Options"}[i]
		ebitenutil.DebugPrintAt(screen, label, r.Min.X+r.Dx()/2-len(label)*3, r.Max.Y-22)
	}
}

func (g *game) Layout(w, h int) (int, int) {
	// Two 180x64 buttons side by side near the bottom.
	g.buttons = []image.Rectangle{
		image.Rect(w/2-196, h-120, w/2-16, h-56),
		image.Rect(w/2+16, h-120, w/2+196, h-56),
	}
	return w, h
}

func main() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("dithergo: aurora")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(&game{}); err != nil {
		log.Fatal(err)
	}
}
