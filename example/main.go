// An animated dithered background: a tone drawn by field.kage, turned into dots by dithergo.
package main

import (
	_ "embed"
	"image/color"
	"log"

	"github.com/dwethmar/dithergo"
	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed field.kage
var fieldSrc []byte

type game struct {
	field *ebiten.Shader
	tone  *ebiten.Image // the field in greys, screen-sized, before dithering
	ticks int
}

func (g *game) Update() error {
	g.ticks++
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	b := screen.Bounds()
	if g.tone == nil || g.tone.Bounds() != b {
		g.tone = ebiten.NewImage(b.Dx(), b.Dy())
	}
	g.tone.DrawRectShader(b.Dx(), b.Dy(), g.field, &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{"Time": float32(g.ticks) / float32(ebiten.TPS())},
	})
	dithergo.Draw(screen, g.tone, &dithergo.Options{
		Ink:   color.RGBA{0x7c, 0x8c, 0xff, 0xff},
		Paper: color.RGBA{0x0a, 0x0a, 0x0a, 0xff},
	})
}

func (g *game) Layout(w, h int) (int, int) { return w, h }

func main() {
	field, err := ebiten.NewShader(fieldSrc)
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("dithergo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(&game{field: field}); err != nil {
		log.Fatal(err)
	}
}
