package dithergo

import (
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// Pixels can only be read while a game runs, so the tests run inside one, in a hidden window.
// That needs a graphics driver: on a machine without one, RunGame fails and so do the tests.
type runner struct {
	m    *testing.M
	code int
}

func (r *runner) Update() error {
	r.code = r.m.Run()
	return ebiten.Termination
}
func (*runner) Draw(*ebiten.Image)         {}
func (*runner) Layout(int, int) (int, int) { return 16, 16 }

func TestMain(m *testing.M) {
	ebiten.SetWindowVisible(false)
	r := &runner{m: m, code: 1}
	if err := ebiten.RunGame(r); err != nil {
		panic(err)
	}
	os.Exit(r.code)
}

// inked dithers a 16x16 image of one grey and counts the pixels that came out as ink.
// Every pixel must be exactly ink or paper.
func inked(t *testing.T, grey uint8, opts *Options) int {
	t.Helper()
	src := ebiten.NewImage(16, 16)
	src.Fill(color.Gray{grey})
	dst := ebiten.NewImage(16, 16)
	Draw(dst, src, opts)
	n := 0
	for y := range 16 {
		for x := range 16 {
			switch dst.At(x, y) {
			case color.RGBA{255, 0, 0, 255}:
				n++
			case color.RGBA{0, 0, 255, 255}:
			default:
				t.Fatalf("grey %d: pixel %d,%d is %v, neither ink nor paper", grey, x, y, dst.At(x, y))
			}
		}
	}
	return n
}

func TestDraw(t *testing.T) {
	red, blue := color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}
	for _, m := range []Matrix{Bayer2, Bayer4, Bayer8} {
		opts := &Options{Cell: 1, Matrix: m, Ink: red, Paper: blue}
		// The share of ink is the tone: none, a quarter, half, three quarters, all.
		for grey, want := range map[uint8]int{0: 0, 64: 64, 128: 128, 191: 192, 255: 256} {
			if got := inked(t, grey, opts); got != want {
				t.Errorf("matrix %d, grey %d: %d of 256 pixels inked, want %d", m, grey, got, want)
			}
		}
	}
}

// A cell is one flat square: with 4-pixel cells every 4x4 block is a single colour.
func TestCell(t *testing.T) {
	src := ebiten.NewImage(16, 16)
	src.Fill(color.Gray{128})
	dst := ebiten.NewImage(16, 16)
	Draw(dst, src, &Options{Cell: 4})
	white := 0
	for y := range 16 {
		for x := range 16 {
			c := dst.At(x, y)
			if c != dst.At(x/4*4, y/4*4) {
				t.Fatalf("pixel %d,%d differs from its cell's corner", x, y)
			}
			if c == (color.RGBA{255, 255, 255, 255}) {
				white++
			}
		}
	}
	if white != 128 { // default ink on default paper, half of it
		t.Errorf("%d of 256 pixels white, want 128", white)
	}
}

// With three colours the tone runs paper, middle, top: mid grey is all the middle colour, and
// a grey halfway up the second step is half middle, half top.
func TestPalette(t *testing.T) {
	pal := []color.Color{color.RGBA{0, 0, 255, 255}, color.RGBA{0, 255, 0, 255}, color.RGBA{255, 0, 0, 255}}
	for grey, wantRed := range map[uint8]int{128: 0, 191: 128, 255: 256} {
		src := ebiten.NewImage(16, 16)
		src.Fill(color.Gray{grey})
		dst := ebiten.NewImage(16, 16)
		Draw(dst, src, &Options{Cell: 1, Palette: pal})
		red := 0
		for y := range 16 {
			for x := range 16 {
				switch dst.At(x, y) {
				case pal[2]:
					red++
				case pal[1]:
				default:
					t.Fatalf("grey %d: pixel %d,%d is %v, want the middle or top colour", grey, x, y, dst.At(x, y))
				}
			}
		}
		if red != wantRed {
			t.Errorf("grey %d: %d of 256 pixels the top colour, want %d", grey, red, wantRed)
		}
	}
}

// The aurora fills its rectangle, and only that, with palette colours, some of them light.
func TestAurora(t *testing.T) {
	pal := AuroraPalette(color.Black, color.RGBA{128, 0, 255, 255}, color.RGBA{0, 255, 128, 255})
	// The GPU rounds a mixed colour to 8 bits its own way: allow one step either side.
	in := func(c color.Color) bool {
		r, g, b, _ := c.RGBA()
		for _, p := range pal {
			pr, pg, pb, _ := p.RGBA()
			if max(r, pr)-min(r, pr) < 0x200 && max(g, pg)-min(g, pg) < 0x200 && max(b, pb)-min(b, pb) < 0x200 {
				return true
			}
		}
		return false
	}
	dst := ebiten.NewImage(80, 60)
	opts := &Options{Palette: pal}
	opts.GeoM.Translate(8, 4)
	Aurora(dst, 64, 48, 10, opts)
	lit := 0
	for y := range 60 {
		for x := range 80 {
			c := dst.At(x, y)
			inside := x >= 8 && x < 72 && y >= 4 && y < 52
			if !inside && c != (color.RGBA{}) {
				t.Fatalf("pixel %d,%d outside the rectangle is %v", x, y, c)
			}
			if inside && !in(c) {
				t.Fatalf("pixel %d,%d is %v, not in the palette", x, y, c)
			}
			if inside && c != (color.RGBA{0, 0, 0, 255}) { // not the black sky
				lit++
			}
		}
	}
	if lit < 64*48/20 || lit > 64*48*9/10 {
		t.Errorf("%d of %d pixels lit: want a sky with lights in it", lit, 64*48)
	}
}

// A map is mostly paper, with lines and stipple in more than one of the palette's colours and
// in nothing else.
func TestContours(t *testing.T) {
	pal := []color.Color{color.RGBA{0, 0, 0, 255}, color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 255}, color.RGBA{0, 0, 255, 255}}
	dst := ebiten.NewImage(400, 400)
	Contours(dst, 400, 400, 10, &Options{Cell: 1, Palette: pal})
	count := map[color.Color]int{}
	for y := range 400 {
		for x := range 400 {
			count[dst.At(x, y)]++
		}
	}
	inks := 0
	for _, c := range pal[1:] {
		if count[c] > 0 {
			inks++
		}
		delete(count, c)
	}
	paper := count[pal[0]]
	delete(count, pal[0])
	if inks < 2 || paper < 400*400/2 || paper == 400*400 {
		t.Errorf("%d ink colours and %d of %d pixels paper: want a map in several colours, mostly paper", inks, paper, 400*400)
	}
	if len(count) > 0 {
		t.Errorf("colours that are not in the palette: %v", count)
	}
}
