// Package dithergo draws ordered (Bayer) dither dots with Ebitengine.
//
// Draw dithers an image: draw anything bright-on-dark (a frame of the game, a sprite, a
// gradient from a shader of your own) and its brightness becomes dots. Black stays paper,
// white is solid ink, and the tones between are dots at a matching density.
//
// Aurora, Contours and NoiseField draw ready-made animated backgrounds the same way, into any
// rectangle. NoiseField is the one with controls: see Noise.
package dithergo

import (
	_ "embed"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Every shader is dither.kage, then noise.kage if it needs noise, then the file with its Fragment.
var (
	//go:embed dither.kage
	ditherSrc string
	//go:embed noise.kage
	noiseSrc string
	//go:embed image.kage
	imageSrc string
	//go:embed aurora.kage
	auroraSrc string
	//go:embed contours.kage
	contoursSrc string
	//go:embed noisefield.kage
	noiseFieldSrc string

	imageShader      = compile(imageSrc)
	auroraShader     = compile(auroraSrc)
	contoursShader   = compile(noiseSrc + contoursSrc)
	noiseFieldShader = compile(noiseSrc + noiseFieldSrc)
)

func compile(fragment string) func() *ebiten.Shader {
	return sync.OnceValue(func() *ebiten.Shader {
		s, err := ebiten.NewShader([]byte(ditherSrc + fragment))
		if err != nil {
			panic("dithergo: " + err.Error()) // an embedded shader is broken: a bug in this package
		}
		return s
	})
}

// Matrix is the Bayer tile the dots are laid out by.
type Matrix int

const (
	Bayer4 Matrix = iota // 17 tones, the default
	Bayer2               // 5 tones, a coarse crosshatch
	Bayer8               // 65 tones, the smoothest gradients
)

const maxColors = 8 // the length of Palette in dither.kage

// Options says how to dither. The zero value is white 3-pixel dots on black.
type Options struct {
	Cell   int         // a dot's size in pixels; 0 means 3
	Matrix Matrix      // the dot pattern
	Ink    color.Color // the dots; nil means white
	Paper  color.Color // behind the dots; nil means black. color.Transparent leaves dst showing
	// Palette is for more than two colours, and then replaces Ink and Paper: 2 to 8 colours,
	// paper first and the brightest last. The tone runs along the list, and each cell is
	// dithered between the two colours it falls between.
	Palette []color.Color
	GeoM    ebiten.GeoM // where on dst the picture goes; the identity is dst's top left
}

// Draw draws src on dst as dither dots. opts may be nil.
func Draw(dst, src *ebiten.Image, opts *Options) {
	op := opts.shaderOptions()
	op.Images[0] = src
	b := src.Bounds()
	dst.DrawRectShader(b.Dx(), b.Dy(), imageShader(), op)
}

// Aurora draws a w by h rectangle of dithered northern lights on dst, as they are seconds
// into their slow motion: pass a clock that keeps running. opts may be nil; AuroraPalette
// makes the palette for a two-colour one.
func Aurora(dst *ebiten.Image, w, h int, seconds float64, opts *Options) {
	op := opts.shaderOptions()
	op.Uniforms["Time"] = float32(seconds)
	dst.DrawRectShader(w, h, auroraShader(), op)
}

// Contours draws a w by h rectangle of a dithered topographic map on dst: contour lines over
// hills that swell, sink and merge like a lava lamp, as they are at seconds, with a stipple on
// the slopes facing away from the light. Multiply seconds to make it faster or slower. Every colour of the palette after the paper is a height, the lowest ground
// first, so Options.Palette with several colours gives a map with a colour per height.
// The hills keep their size in dots, so a small rectangle shows a small piece of the map.
// opts may be nil.
func Contours(dst *ebiten.Image, w, h int, seconds float64, opts *Options) {
	op := opts.shaderOptions()
	op.Uniforms["Time"] = float32(seconds)
	dst.DrawRectShader(w, h, contoursShader(), op)
}

// AuroraPalette is the palette of a two-colour aurora on a sky of colour paper: the curtains'
// bright lower edges are edge, and their rays fade through tip into the sky.
func AuroraPalette(paper, tip, edge color.Color) []color.Color {
	return []color.Color{paper, mix(tip, paper, 0.3), mix(tip, paper, 0.7), edge, mix(color.White, edge, 0.5)}
}

// mix is base with share (0 to 1) of c mixed in.
func mix(c, base color.Color, share float64) color.Color {
	r, g, b, a := c.RGBA()
	br, bg, bb, ba := base.RGBA()
	lerp := func(x, y uint32) uint16 { return uint16(float64(y) + (float64(x)-float64(y))*share) }
	return color.RGBA64{lerp(r, br), lerp(g, bg), lerp(b, bb), lerp(a, ba)}
}

// shaderOptions is what dither.kage needs to draw as o says. o may be nil.
func (o *Options) shaderOptions() *ebiten.DrawRectShaderOptions {
	if o == nil {
		o = &Options{}
	}
	cell := o.Cell
	if cell <= 0 {
		cell = 3
	}
	colors := o.Palette
	if len(colors) < 2 {
		colors = []color.Color{o.Paper, o.Ink}
		if colors[0] == nil {
			colors[0] = color.Black
		}
		if colors[1] == nil {
			colors[1] = color.White
		}
	}
	if len(colors) > maxColors {
		panic("dithergo: a palette holds at most 8 colours")
	}
	palette := make([]float32, maxColors*4) // premultiplied, like everything in Ebitengine
	for i, c := range colors {
		r, g, b, a := c.RGBA()
		palette[i*4], palette[i*4+1], palette[i*4+2], palette[i*4+3] =
			float32(r)/0xffff, float32(g)/0xffff, float32(b)/0xffff, float32(a)/0xffff
	}
	return &ebiten.DrawRectShaderOptions{GeoM: o.GeoM, Uniforms: map[string]any{
		"Cell":    float32(cell),
		"Order":   [...]float32{Bayer4: 2, Bayer2: 1, Bayer8: 3}[o.Matrix],
		"Palette": palette,
		"Colors":  float32(len(colors)),
	}}
}
