# CLAUDE.md

Guidance for Claude Code when working in this repository.

## What this is

dithergo: a small Go library that draws ordered (Bayer) dither dots with Ebitengine (`github.com/hajimehoshi/ebiten/v2`, v2.10+). Modelled on https://dithercn.dev. 
## Commands

```sh
go test ./...              # pixel tests; opens a hidden window, needs a graphics driver
go vet ./...
gofmt -l .                 # must print nothing
go run ./example           # Draw on a custom shader's picture
go run ./example/aurora    # Aurora as a background and in buttons
go run ./example/contours  # Contours
go run ./example/noisefield  # NoiseField, with its controls on the keyboard
```

## Files

- `dither.go` — most of the Go API: `Draw` (dither an image), `Aurora` and `Contours` (ready-made animated fields), `AuroraPalette`, `Options`, `Matrix`. `Options.shaderOptions` turns options into the uniforms every shader shares. It also embeds and compiles every shader.
- `noisefield.go` — `NoiseField` and its controls: the `Noise` struct, `DefaultNoise`, `NoiseVariant`, `Mask`. It clamps every control and passes each as a uniform of the same name.
- `dither.kage` — the start of every shader: the uniforms (`Cell`, `Order`, `Palette`, `Colors`), the Bayer `threshold`, and the two ways to colour a cell. `dither(cell, tone)` runs the tone along the palette. `ink(cell, index, amount)` dithers paper against one chosen palette colour, for when colour and density do not go together.
- `noise.kage` — simplex noise, `gnoise(p, turn)`, shared by `contours.kage` and `noisefield.kage`. A `turn` that grows with time makes the bumps morph in place (the lava lamp look of `Contours`); 0 keeps them still.
- `image.kage`, `aurora.kage`, `contours.kage`, `noisefield.kage` — one `Fragment` each. They are not Kage files on their own: `compile` in `dither.go` puts `dither.kage` in front of each, and `noise.kage` too for the ones that use it. So they have no `package` clause and no `//kage:unit` line. `aurora.kage` has its own simpler noise, which its look is tuned to.
- `dither_test.go`, `noisefield_test.go` — run inside a real game loop (`TestMain` in `dither_test.go`), because pixels can only be read while a game runs.
- `example/` — four small programs. `example/field.kage` shows how to write a tone picture of your own for `Draw`.
- `docs/*.png` — renders of the examples, used by the README.

## Adding a new effect

1. Write `name.kage` with a `var Time float` and a `Fragment` that works out the cell (`floor((srcPos - imageSrc0Origin()) / Cell)`) and returns `dither(...)` or `ink(...)`. Copy the shape of `contours.kage`.
2. Embed and compile it in `dither.go`, and add a function shaped like `Aurora`.
3. Add a pixel test and, if it is worth showing, an example.

## Rules for the shaders

- Read Ebitengine's own Kage guide before changing a shader: `skills/writing-kage-shaders/SKILL.md` inside the ebiten module in the Go module cache.
- Colours are premultiplied alpha everywhere.
- Index `Palette` only through `paletteAt` (a loop counter), never `Palette[int(x)]`: WebGL 1 rejects it.
- Keep `bayer2` exact: it works on 0 and 1 only, so the dot counts in the tests hold on any float precision.
- `Aurora` and `Contours` have no knobs on purpose: their numbers are constants in the `.kage` files. `NoiseField` is the one with controls. A new control there is a field of `Noise` with a default in `DefaultNoise`, a clamped uniform in `NoiseField`, a `var` in `noisefield.kage`, a row in the README table and a knob in `example/noisefield`.
- The caller controls speed by scaling the `seconds` they pass.
- `noisefield.kage` follows dithercn's noise field formula, whose numbers are tuned for simplex noise. Keep `gnoise` simplex: plain grid noise made the smoke blocky.

## Seeing what it draws

To check a look, render an example to a PNG without a window, using the driver that ships with Ebitengine (`skills/run-ebitengine-app-headless/` in the ebiten module). Copy `_driver/main.go` to `.vmdriver/` here, and on Windows change `filepath.Join(dir, "guest")` to `"guest.exe"` or the guest fails to start. Then:

```sh
go run ./.vmdriver -pkg ./example/contours -ticks 300 -w 800 -h 600 -out frame.png
```

Delete `.vmdriver/` afterwards. For motion, render a few frames some seconds (60 ticks each) apart and compare them.
