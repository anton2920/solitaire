package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/anton2920/gofa/gui"
	"github.com/anton2920/gofa/gui/gr"
	"github.com/anton2920/gofa/log"
)

var BuildMode string

var Debug bool

type State int

const (
	MainMenu State = iota
	GameSolitaire
	GameFreeCell
)

func DrawMainMenu(ui *gui.UI, state *State) {
	if ui.Button(gui.ID(uintptr(1)), "Play Solitaire") {
		*state = GameSolitaire
	}

	if ui.Button(gui.ID(uintptr(2)), "Play FreeCell") {
		*state = GameFreeCell
	}
}

func DrawSolitaire(window *gui.Window, renderer *gui.Renderer, ui *gui.UI, state *State) {
	const text = "Playing Solitaire..."
	textWidth := ui.Font.TextWidth(text)
	textHeight := ui.Font.TextHeight(text)
	renderer.GraphText(text, ui.Font, window.Width/2-textWidth/2, window.Height/2-textHeight/2, gr.ColorWhite)

	if ui.Button(gui.ID(uintptr(1)), "Back") {
		*state = MainMenu
	}
}

func DrawFreeCell(window *gui.Window, renderer *gui.Renderer, ui *gui.UI, state *State) {
	const text = "Playing FreeCell..."
	textWidth := ui.Font.TextWidth(text)
	textHeight := ui.Font.TextHeight(text)
	renderer.GraphText(text, ui.Font, window.Width/2-textWidth/2, window.Height/2-textHeight/2, gr.ColorWhite)

	if ui.Button(gui.ID(uintptr(1)), "Back") {
		*state = MainMenu
	}
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

	window, err := gui.NewWindow("Classic solitaire collection", 0, 0, 800, 600, gui.WindowResizable)
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

		renderer.Clear(gr.ColorBlack)
		ui.Begin()

		switch state {
		case MainMenu:
			DrawMainMenu(ui, &state)
		case GameSolitaire:
			DrawSolitaire(window, renderer, ui, &state)
		case GameFreeCell:
			DrawFreeCell(window, renderer, ui, &state)
		}

		ui.End()
		renderer.Present()
		window.SyncFPS(60)
	}
}
