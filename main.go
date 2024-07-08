package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"unsafe"

	"github.com/anton2920/gofa/gui"
	"github.com/anton2920/gofa/gui/gr"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/slices"
)

var BuildMode string

var Debug bool

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

		for i := 1; i <= 10; i++ {
			var n int

			buffer := make([]byte, 32)
			n += copy(buffer, "Press me!!! ")
			n += slices.PutInt(buffer[n:], i)
			text := unsafe.String(&buffer[0], n)

			if ui.Button(gui.ID(uintptr(i)), text) {
				log.Infof("Button %d pressed!!!", i)
			}
		}

		ui.End()

		renderer.Present()

		window.SyncFPS(60)
	}
}
