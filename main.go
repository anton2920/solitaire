package main

import (
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"runtime"

	"github.com/anton2920/gofa/debug"
	"github.com/anton2920/gofa/gui"
	"github.com/anton2920/gofa/gui/color"
	"github.com/anton2920/gofa/gui/gr"
	"github.com/anton2920/gofa/log"
	_ "github.com/anton2920/gofa/time_"
	"github.com/anton2920/gofa/trace"
	"github.com/anton2920/gofa/trace_"
)

type GameType int

const (
	GameNone GameType = iota
	GameSolitaire
	GameFreeCell
)

const Title = "Classic solitaire collection"

var (
	CurrentGame  GameType
	FreeCellGame FreeCell
)

func DrawRectWithShadow(renderer gui.Renderer, x0, y0, x1, y1 int, pclr, sclr color.Color) {
	renderer.RenderLine(x0, y0, x1-1, y0, pclr)
	renderer.RenderLine(x0, y0, x0, y1-1, pclr)
	renderer.RenderLine(x0+1, y1, x1, y1, sclr)
	renderer.RenderLine(x1, y0+1, x1, y1, sclr)
}

func DrawBackButton(window *gui.Window, ui *gui.UI) {
	t := trace.Begin("")

	ui.Layout.CurrentY = window.Height - 50
	if ui.Button(gui.ID(&CurrentGame), "Back") {
		window.SetTitle(Title)
		CurrentGame = GameNone
	}

	trace.End(t)
}

func DrawSolitaire(window *gui.Window, renderer gui.Renderer, ui *gui.UI) {
	const text = "Playing Solitaire..."
	textWidth := ui.Font.TextWidth(text)
	textHeight := ui.Font.TextHeight(text)
	renderer.RenderText(text, ui.Font, window.Width/2-textWidth/2, window.Height/2-textHeight/2, color.White)

	DrawBackButton(window, ui)
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
	trace.BeginProfile()

	if debug.Debug {
		log.SetLevel(log.LevelDebug)
	}
	log.Infof("Starting Solitaire... (%s)", runtime.Version())

	f, err := os.Open("assets/assets.png")
	if err != nil {
		log.Fatalf("Failed to load assets file: %v", err)
	}
	assetsImage, err := png.Decode(f)
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

	/* NOTE(anton2920); first time calling 'math.Rand()' allocates memory. */
	_ = rand.Int()

	runtime.AllocationsAreDisabled = true

	events := make([]gui.Event, 64)
	var quit bool
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

		switch CurrentGame {
		case GameNone:
			renderer.Clear(color.Black)
			if ui.Button(gui.ID2(gui.ID(&CurrentGame)), "Play Solitaire") {
				window.SetTitle(Title + ": Solitaire")
				CurrentGame = GameSolitaire
			}
			if ui.Button(gui.ID3(gui.ID(&CurrentGame)), "Play FreeCell") {
				FreeCellGame.Init(window, renderer, ui, &assets)
				FreeCellGame.NewRandomGame()
				CurrentGame = GameFreeCell
			}
		case GameSolitaire:
			DrawSolitaire(window, renderer, ui)
		case GameFreeCell:
			FreeCellGame.UpdateAndRender()
			DrawBackButton(window, ui)
		}

		ui.End()

		renderer.Present()
		window.SyncFPS(60)
	}

	runtime.AllocationsAreDisabled = false
	trace_.EndAndPrintProfile()
}
