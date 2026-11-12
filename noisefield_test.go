package dithergo

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// noiseInk draws a 200x200 noise field with the default options, white on black, and counts
// the white pixels inside r. Any pixel that is neither fails the test.
func noiseInk(t *testing.T, n *Noise, r image.Rectangle) int {
	t.Helper()
	dst := ebiten.NewImage(200, 200)
	NoiseField(dst, 200, 200, 10, n, &Options{Cell: 1})
	ink := 0
	for y := range 200 {
		for x := range 200 {
			switch dst.At(x, y) {
			case color.RGBA{255, 255, 255, 255}:
				if image.Pt(x, y).In(r) {
					ink++
				}
			case color.RGBA{0, 0, 0, 255}:
			default:
				t.Fatalf("pixel %d,%d is %v, neither ink nor paper", x, y, dst.At(x, y))
			}
		}
	}
	return ink
}

func TestNoiseField(t *testing.T) {
	all := image.Rect(0, 0, 200, 200)
	with := func(change func(*Noise)) int {
		t.Helper()
		n := DefaultNoise()
		change(&n)
		return noiseInk(t, &n, all)
	}

	if ink := noiseInk(t, nil, all); ink == 0 || ink == 200*200 {
		t.Errorf("no controls: %d of %d pixels inked, want some", ink, 200*200)
	}
	for v := NoiseSmoke; v <= NoiseMarble; v++ {
		if ink := with(func(n *Noise) { n.Variant = v }); ink == 0 || ink == 200*200 {
			t.Errorf("variant %d: %d of %d pixels inked, want some", v, ink, 200*200)
		}
	}
	if ink := with(func(n *Noise) { n.Intensity = 0 }); ink != 0 {
		t.Errorf("intensity 0: %d pixels inked, want none", ink)
	}
	sparse, dense := with(func(n *Noise) { n.Coverage = -0.4 }), with(func(n *Noise) { n.Coverage = 0.4 })
	if sparse >= dense {
		t.Errorf("coverage -0.4 inks %d pixels and 0.4 inks %d: want more with more coverage", sparse, dense)
	}
	plain, inverted := with(func(n *Noise) {}), with(func(n *Noise) { n.Invert = true })
	if inverted <= plain { // smoke is mostly paper, so inverted it is mostly ink
		t.Errorf("inverted smoke inks %d pixels, plain %d: want more inverted", inverted, plain)
	}

	// A mask in the middle clears the middle.
	middle := image.Rect(80, 80, 120, 120)
	marble := DefaultNoise()
	marble.Variant = NoiseMarble
	open := noiseInk(t, &marble, middle)
	marble.Mask, marble.MaskSize = MaskCenter, 1
	if masked := noiseInk(t, &marble, middle); masked >= open/2 {
		t.Errorf("the middle has %d inked pixels with a centre mask and %d without: want it cleared", masked, open)
	}
}
