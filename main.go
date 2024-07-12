package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"runtime/pprof"
	"unsafe"

	"github.com/anton2920/gofa/gui"
	"github.com/anton2920/gofa/gui/color"
	"github.com/anton2920/gofa/gui/gr"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/slices"

	"freecell"
)

type State int

const (
	MainMenu State = iota
	GameSolitaire
	GameFreeCell
)

const Title = "Classic solitaire collection"

//go:embed assets/assets.png
var AssetsData []byte

var BuildMode string
var Debug bool

func DrawMainMenu(window *gui.Window, renderer gui.Renderer, ui *gui.UI, state *State) {
	renderer.Clear(color.Black)

	if ui.Button(gui.ID(uintptr(1)), "Play Solitaire") {
		window.SetTitle(Title + ": Solitaire")
		*state = GameSolitaire
	}

	if ui.Button(gui.ID(uintptr(2)), "Play FreeCell") {
		// N := rand.Int() % 1000001
		// const N = 17330
		const N = 473846

		var n int
		buffer := make([]byte, 128)
		n += copy(buffer[n:], Title)
		n += copy(buffer[n:], ": FreeCell Game #")
		n += slices.PutInt(buffer[n:], N)
		title := unsafe.String(&buffer[0], n)
		window.SetTitle(title)

		freecell.NewGame(N)
		*state = GameFreeCell
	}
}

func DrawBackButton(window *gui.Window, ui *gui.UI, state *State) {
	ui.Layout.CurrentY = window.Height - 50
	if ui.Button(gui.ID(uintptr(1)), "Back") {
		window.SetTitle(Title)
		*state = MainMenu
	}
}

func DrawSolitaire(window *gui.Window, renderer gui.Renderer, ui *gui.UI, state *State) {
	const text = "Playing Solitaire..."
	textWidth := ui.Font.TextWidth(text)
	textHeight := ui.Font.TextHeight(text)
	renderer.RenderText(text, ui.Font, window.Width/2-textWidth/2, window.Height/2-textHeight/2, color.White)

	DrawBackButton(window, ui, state)
}

func Image2RGBA(src image.Image) *image.RGBA {
	if dst, ok := src.(*image.RGBA); ok {
		return dst
	}

	dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	return dst
}

func main() {
	switch BuildMode {
	default:
		log.Fatalf("Build mode %q is not recognized", BuildMode)
	case "Release":
	case "Debug":
		Debug = true
		log.SetLevel(log.LevelDebug)
	case "Profiling":
		f, err := os.Create(fmt.Sprintf("solitaire-%d-cpu.pprof", os.Getpid()))
		if err != nil {
			log.Fatalf("Failed to create a profiling file: %v", err)
		}
		defer f.Close()

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	assetsImage, err := png.Decode(bytes.NewReader(AssetsData))
	if err != nil {
		log.Fatalf("Failed to decode assets data: %v", err)
	}
	assets := gr.NewPixmapFromImage(Image2RGBA(assetsImage), gr.AlphaOpaque)

	window, err := gui.NewWindow("Classic solitaire collection", 632, 452, gui.WindowResizable)
	if err != nil {
		log.Fatalf("Failed to open new window: %v", err)
	}
	defer window.Close()

	renderer := gui.NewSoftwareRenderer(window)
	ui := gui.NewUI(renderer)

	events := make([]gui.Event, 64)
	var state State
	quit := false

	for !quit {
		for window.HasEvents() {
			n, err := window.GetEvents(events)
			if err != nil {
				log.Errorf("Failed to get events from window: %v", err)
				continue
			}

			for i := 0; i < n; i++ {
				event := &events[i]
				// log.Debugf("Event %#v", event)

				switch event.Type {
				case gui.DestroyEvent:
					quit = true
				case gui.ResizeEvent:
					renderer.Resize(event.Width, event.Height)
				case gui.MousePressEvent:
					ui.MousePress(event.X, event.Y, event.Button)
				case gui.MouseReleaseEvent:
					ui.MouseRelease(event.X, event.Y, event.Button)
				case gui.MouseMoveEvent:
					ui.MouseMove(event.X, event.Y)
				}
			}
		}

		ui.Begin()

		switch state {
		case MainMenu:
			DrawMainMenu(window, renderer, ui, &state)
		case GameSolitaire:
			DrawSolitaire(window, renderer, ui, &state)
		case GameFreeCell:
			freecell.Main(window, renderer, ui, &assets)
			DrawBackButton(window, ui, &state)
		}

		ui.End()

		renderer.Present()
		window.SyncFPS(60)
	}
}
