package dithergo

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// NoiseVariant is what a noise field draws with its noise.
type NoiseVariant int

const (
	NoiseSmoke    NoiseVariant = iota // soft folds that dither into smooth gradients
	NoiseRidges                       // sharp creases, like veins or lightning
	NoiseContours                     // topographic lines, Bands of them
	NoiseMarble                       // warped stripes, Bands of them
)

// Mask is where a noise field fades to paper, to keep text on top of it readable.
type Mask int

const (
	MaskNone   Mask = iota
	MaskCenter      // clears the middle
	MaskTop         // clears the top
	MaskBottom      // clears the bottom
	MaskEdges       // clears everything but the middle
)

// Noise is the controls of NoiseField. The zero value is not a useful field: start from
// DefaultNoise and change what you need. Values outside a range are brought back into it.
type Noise struct {
	Variant NoiseVariant

	// The shape of the noise.
	Scale  float64 // size of the features: higher is smaller, busier noise. From 0.1
	Warp   float64 // how far the noise folds back on itself, 0 (plain noise) to about 4
	Detail int     // layers of noise, 1 (smooth blobs) to 6 (fine grain on top)

	// How much of it turns to ink.
	Contrast  float64 // 1 is soft; 2 and up give hard-edged islands. Not for NoiseContours
	Coverage  float64 // how much of the rectangle is inked, -0.5 (sparse) to 0.5 (dense)
	Intensity float64 // the brightest tone it reaches, 0 to 1
	Invert    bool    // swaps paper and ink

	Bands float64 // lines for NoiseContours, stripes for NoiseMarble, 2 to 24

	// Where it goes.
	Angle float64 // the direction it drifts in, in degrees: 0 is right, 90 is up
	Flow  float64 // how fast it drifts that way, 0 (it only folds in place) to about 4

	Mask     Mask
	MaskSize float64 // how much of the rectangle the mask clears, 0 to 1
}

// DefaultNoise is slow smoke that drifts upward.
func DefaultNoise() Noise {
	return Noise{
		Scale: 1.3, Warp: 1.7, Detail: 4,
		Contrast: 1.2, Intensity: 0.9,
		Bands: 8,
		Angle: 90, Flow: 1,
		MaskSize: 0.5,
	}
}

// NoiseField draws a w by h rectangle of dithered, slowly folding noise on dst, as it is at
// seconds: pass a clock that keeps running, and multiply it to change the speed of everything.
// noise may be nil for DefaultNoise, and opts may be nil. The tone runs along the palette,
// so Options.Palette with three colours (paper, ink, highlight) puts the third on the
// brightest peaks.
func NoiseField(dst *ebiten.Image, w, h int, seconds float64, noise *Noise, opts *Options) {
	if noise == nil {
		d := DefaultNoise()
		noise = &d
	}
	clamp := func(x, lo, hi float64) float32 { return float32(min(max(x, lo), hi)) }
	angle := noise.Angle * math.Pi / 180
	flow := max(noise.Flow, 0)
	invert := float32(0)
	if noise.Invert {
		invert = 1
	}
	op := opts.shaderOptions()
	for name, v := range map[string]any{
		"Time":      float32(seconds),
		"Variant":   float32(noise.Variant),
		"Scale":     clamp(noise.Scale, 0.1, math.Inf(1)),
		"Warp":      clamp(noise.Warp, 0, math.Inf(1)),
		"Detail":    clamp(float64(noise.Detail), 1, 6),
		"Contrast":  clamp(noise.Contrast, 0, math.Inf(1)),
		"Coverage":  clamp(noise.Coverage, -1, 1),
		"Intensity": clamp(noise.Intensity, 0, 1),
		"Bands":     clamp(noise.Bands, 1, 48),
		"Drift":     []float32{float32(math.Cos(angle) * flow), float32(math.Sin(angle) * flow)},
		"Invert":    invert,
		"Mask":      float32(noise.Mask),
		"MaskSize":  clamp(noise.MaskSize, 0, 1),
	} {
		op.Uniforms[name] = v
	}
	dst.DrawRectShader(w, h, noiseFieldShader(), op)
}
