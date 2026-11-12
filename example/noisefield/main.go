// The noise field with its controls on the keyboard, to find a look and read off its numbers.
//
//	1 2 3 4      smoke, ridges, contours, marble
//	M            the next mask
//	I            invert
//	up, down     pick a control
//	left, right  change it
package main

import (
	"fmt"
	"image/color"
	"log"
	"strings"

	"github.com/dwethmar/dithergo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	// Paper, ink, and a highlight on the brightest peaks.
	palette = []color.Color{
		color.RGBA{0x08, 0x0a, 0x12, 0xff},
		color.RGBA{0x7c, 0x8c, 0xff, 0xff},
		color.RGBA{0xf2, 0xf4, 0xff, 0xff},
	}
	variants = []string{"smoke", "ridges", "contours", "marble"}
	masks    = []string{"none", "center", "top", "bottom", "edges"}
)

// knob is one number of the controls, and how far a key press moves it.
type knob struct {
	name  string
	value *float64
	step  float64
}

type game struct {
	noise  dithergo.Noise
	detail float64 // noise.Detail, as a number the knobs can change
	knobs  []knob
	picked int
	ticks  int
}

func newGame() *game {
	g := &game{noise: dithergo.DefaultNoise()}
	g.detail = float64(g.noise.Detail)
	n := &g.noise
	g.knobs = []knob{
		{"Scale", &n.Scale, 0.1}, {"Warp", &n.Warp, 0.1}, {"Detail", &g.detail, 1},
		{"Contrast", &n.Contrast, 0.1}, {"Coverage", &n.Coverage, 0.05}, {"Intensity", &n.Intensity, 0.05},
		{"Bands", &n.Bands, 1}, {"Angle", &n.Angle, 15}, {"Flow", &n.Flow, 0.25}, {"MaskSize", &n.MaskSize, 0.05},
	}
	return g
}

// pressed is true when k goes down, and again and again while it is held.
func pressed(k ebiten.Key) bool {
	d := inpututil.KeyPressDuration(k)
	return d == 1 || d > 20 && d%3 == 0
}

func (g *game) Update() error {
	g.ticks++
	for i := range variants {
		if inpututil.IsKeyJustPressed(ebiten.KeyDigit1 + ebiten.Key(i)) {
			g.noise.Variant = dithergo.NoiseVariant(i)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.noise.Mask = (g.noise.Mask + 1) % dithergo.Mask(len(masks))
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.noise.Invert = !g.noise.Invert
	}
	if pressed(ebiten.KeyArrowDown) {
		g.picked = (g.picked + 1) % len(g.knobs)
	}
	if pressed(ebiten.KeyArrowUp) {
		g.picked = (g.picked + len(g.knobs) - 1) % len(g.knobs)
	}
	k := g.knobs[g.picked]
	if pressed(ebiten.KeyArrowRight) {
		*k.value += k.step
	}
	if pressed(ebiten.KeyArrowLeft) {
		*k.value -= k.step
	}
	g.noise.Detail = int(g.detail)
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	b := screen.Bounds()
	seconds := float64(g.ticks) / float64(ebiten.TPS())
	dithergo.NoiseField(screen, b.Dx(), b.Dy(), seconds, &g.noise, &dithergo.Options{Matrix: dithergo.Bayer8, Palette: palette})

	// The controls, as they are now. The library brings a number outside its range back
	// into it, so past that point a knob changes nothing.
	var s strings.Builder
	fmt.Fprintf(&s, "  Variant   %s  (1-4)\n  Mask      %s  (M)\n  Invert    %v  (I)\n", variants[g.noise.Variant], masks[g.noise.Mask], g.noise.Invert)
	for i, k := range g.knobs {
		mark := " "
		if i == g.picked {
			mark = ">"
		}
		fmt.Fprintf(&s, "%s %-9s %.2f\n", mark, k.name, *k.value)
	}
	vector.FillRect(screen, 8, 8, 190, 16*13+12, color.RGBA{0, 0, 0, 0xd0}, false)
	ebitenutil.DebugPrintAt(screen, s.String(), 14, 14)
}

func (g *game) Layout(w, h int) (int, int) { return w, h }

func main() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("dithergo: noise field")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
