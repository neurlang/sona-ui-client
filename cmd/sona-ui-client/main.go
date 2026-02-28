// Copyright 2021 Neurlang project

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

// Go Wayland Smoke demo
package main

import "time"
import "math/rand"
import "os"
import (
	"flag"
	"image"
	"image/color"
	"sync"
	"sync/atomic"
)
import cairo "github.com/neurlang/wayland/cairoshim"
import "github.com/neurlang/wayland/wl"
import "github.com/neurlang/wayland/window"
import xkb "github.com/neurlang/wayland/xkbcommon"
import "fmt"
import gg "github.com/fogleman/gg"

type smoke struct {
	display     *window.Display
	window      *window.Window
	widget      *window.Widget
	width       int32
	height      int32
	smallwidth  int32
	smallheight int32
	rs          *RecordedSamples
	start, stop chan struct{}
	entered     atomic.Bool
	transcribin atomic.Bool
	fontSize    byte
	input       *window.Input
	inputMut    sync.Mutex
}

const maxx = 512
const maxy = 256

func (smoke *smoke) Resize(_ *window.Widget, _ int32, _ int32, width int32, height int32) {

	if smoke.smallwidth == width && smoke.smallheight == height {
		return
	}

	if width > maxx {
		smoke.smallwidth = width
		smoke.width = maxx
	} else {
		smoke.smallwidth = width
		smoke.width = width
	}
	if height > maxy {
		smoke.smallheight = height
		smoke.height = maxy
	} else {
		smoke.smallheight = height
		smoke.height = height
	}
}

func makeContextFromCairo(s cairo.Surface) *gg.Context {
	if s == nil {
		println("no cairo")
		return gg.NewContext(0, 0)
	}

	context := gg.NewContext(s.ImageSurfaceGetWidth(), s.ImageSurfaceGetHeight())
	(context.Image()).(*image.RGBA).Pix = s.ImageSurfaceGetData()
	return context
}
func render(smoke *smoke, surface cairo.Surface) {
	var dst8 = surface.ImageSurfaceGetData()
	var width = surface.ImageSurfaceGetWidth()
	var height = surface.ImageSurfaceGetHeight()
	var stride = surface.ImageSurfaceGetStride()

	if smoke.rs == nil {
		return
	}

	const stretch = 32

	data := smoke.rs.GetLastSamples(width * stretch)

	var a = byte(0x33)

	if !smoke.entered.Load() {
		a = 0x55
	}
	for y := 0; y < height; y++ {
		var yy = (int(32768) * y / int(smoke.smallheight)) - 16384
		for x := 0; x < width; x++ {
			var c byte

			for i := 0; i < stretch; i++ {

				if data[stretch*x+i] > 0 {
					if yy > 0 && yy < data[stretch*x+i] {
						c = byte(0xff)
					}
				} else {
					if yy < 0 && yy > data[stretch*x+i] {
						c = byte(0xff)
					}
				}
			}

			if dst8 != nil {
				dst8[4*x+y*stride+0] = byte(c)
				dst8[4*x+y*stride+1] = byte(c)
				dst8[4*x+y*stride+2] = byte(c)
				dst8[4*x+y*stride+3] = byte(a)
			}
		}
	}

	ctx := makeContextFromCairo(surface)

	// Load system font with IPA support
	// On macOS, try fonts with good Unicode/IPA coverage
	fontPath := "/System/Library/Fonts/Supplemental/Arial Unicode.ttf"
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		fontPath = "/System/Library/Fonts/SFNS.ttf" // SF Pro has excellent Unicode support
		if _, err := os.Stat(fontPath); os.IsNotExist(err) {
			fontPath = "/System/Library/Fonts/Menlo.ttc" // Menlo also has good IPA support
			if _, err := os.Stat(fontPath); os.IsNotExist(err) {
				// Fallback to Linux font
				fontPath = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
			}
		}
	}
	err := ctx.LoadFontFace(fontPath, float64(smoke.fontSize))
	if err != nil {
		// Silently continue without font
	}

	ctx.SetColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	ctx.DrawStringWrapped(smoke.rs.GetText(),
		float64(smoke.smallwidth/2), float64(smoke.smallheight/2), 0.5, 0.5,
		float64(smoke.smallwidth), 2, gg.AlignCenter)

}
func (smoke *smoke) Redraw(widget *window.Widget) {
	//var lastTime = smoke.widget.WidgetGetLastTime()
	var surface = smoke.window.WindowGetSurface()

	if surface != nil {

		render(smoke, surface)
		surface.Destroy()
	}

	smoke.widget.ScheduleRedraw()
}
func (smoke *smoke) Key(
	_ *window.Window,
	input *window.Input,
	time uint32,
	key uint32,
	notUnicode uint32,
	state wl.KeyboardKeyState,
	_ window.WidgetHandler,
) {
	println(notUnicode)

	if input != nil {
		smoke.inputMut.Lock()
		smoke.input = input
		smoke.inputMut.Unlock()
	}

	if notUnicode == xkb.KeyQ || notUnicode == xkb.KEYq {
		smoke.display.Exit()
	} else {
		if state == wl.KeyboardKeyStateReleased {
			smoke.Leave(nil, input)
		}
	}
}
func (smoke *smoke) Focus(_ *window.Window, input *window.Input) {
	if input != nil {
		smoke.inputMut.Lock()
		smoke.input = input
		smoke.inputMut.Unlock()
	}
}
func (smoke *smoke) Enter(_ *window.Widget, input *window.Input, x float32, y float32) {

	if input != nil {
		smoke.inputMut.Lock()
		smoke.input = input
		smoke.inputMut.Unlock()
	}

	if !smoke.entered.Load() {
		smoke.entered.Store(true)
		smoke.start <- struct{}{}
	}

}
func (smoke *smoke) Leave(_ *window.Widget, input *window.Input) {

	if input != nil {
		smoke.inputMut.Lock()
		smoke.input = input
		smoke.inputMut.Unlock()
	}

	if smoke.entered.Load() {
		smoke.entered.Store(false)
		smoke.stop <- struct{}{}
	}

}

func (smoke *smoke) Motion(
	_ *window.Widget,
	input *window.Input,
	time uint32,
	x float32,
	y float32,
) int {
	if input != nil {
		smoke.inputMut.Lock()
		smoke.input = input
		smoke.inputMut.Unlock()
	}
	return window.CursorHand1
}

func (smoke *smoke) Button(
	widget *window.Widget,
	input *window.Input,
	time uint32,
	button uint32,
	state wl.PointerButtonState,
	handler window.WidgetHandler,
) {

	if !smoke.transcribin.Load() {
		if state == wl.PointerButtonStatePressed {
			smoke.Leave(widget, input)
		} else {
			smoke.Enter(widget, input, 0, 0)
		}
	}
}

func (*smoke) TouchUp(
	_ *window.Widget,
	_ *window.Input,
	serial uint32,
	time uint32,
	id int32,
) {
}

func (*smoke) TouchDown(
	_ *window.Widget,
	_ *window.Input,
	serial uint32,
	time uint32,
	id int32,
	x float32,
	y float32,
) {
}

func (smoke *smoke) TouchMotion(
	_ *window.Widget,
	_ *window.Input,
	time uint32,
	id int32,
	x float32,
	y float32,
) {

}
func (*smoke) TouchFrame(_ *window.Widget, _ *window.Input) {
}
func (*smoke) TouchCancel(_ *window.Widget, width int32, height int32) {
}

func (smoke *smoke) Axis(
	widget *window.Widget,
	input *window.Input,
	time uint32,
	axis uint32,
	value float32,
) {
	if input != nil {
		smoke.inputMut.Lock()
		smoke.input = input
		smoke.inputMut.Unlock()
	}
	if value < 0 {
		smoke.fontSize++
	} else {
		smoke.fontSize--
	}

}
func (*smoke) AxisSource(_ *window.Widget, _ *window.Input, source uint32) {
}
func (*smoke) AxisStop(_ *window.Widget, _ *window.Input, time uint32, axis uint32) {
}

func (*smoke) AxisDiscrete(
	_ *window.Widget,
	_ *window.Input,
	axis uint32,
	discrete int32,
) {
}
func (*smoke) PointerFrame(_ *window.Widget, _ *window.Input) {
}

func (smoke *smoke) free() {
	// tear down the rendering pipe
	smoke.rs = nil
}

func main() {

	// Define command line flags
	host := flag.String("host", "127.0.0.1", "Host address for the API server")
	port := flag.String("port", "", "Port number for the API server")
	setup := flag.Bool("setup", false, "Setup Hotkey")
	hotkey := flag.String("hotkey", "<Ctrl><Alt>R", "Hotkey to setup")
	filePath := flag.String("file", "", "Path to the WAV file")
	once := flag.Bool("once", false, "Run once")
	flag.Parse()

	if setup != nil && *setup {
		err := SetupHotkey(*hotkey)
		if err != nil {
			*hotkey = err.Error()
		} else {
			*hotkey = "Hotkey " + *hotkey + " setup sucessfully. "
		}
	}

	if port != nil && *port == "" {
		sonaPort, err := findPort("sona")
		if err != nil {
			fmt.Println("Error finding sona port:", err)
		} else {
			*port = fmt.Sprint(sonaPort)
		}
	}

	var smoke smoke

	d, err := window.DisplayCreate([]string{})
	if err != nil {
		fmt.Println(err)
		return
	}

	smoke.width = 200
	smoke.height = 200
	smoke.display = d
	smoke.window = window.Create(d)

	smoke.widget = smoke.window.AddWidget(&smoke)

	smoke.window.SetTitle("sona")
	smoke.window.SetBufferType(window.BufferTypeShm)
	smoke.window.SetKeyboardHandler(&smoke)
	rand.Seed(int64(time.Now().Nanosecond()))

	smoke.rs = new(RecordedSamples)
	smoke.start, smoke.stop = make(chan struct{}, 0), make(chan struct{}, 0)
	smoke.fontSize = 16

	var str = "Click to record, press any key to transcribe. IPA is automatically copied."

	if setup != nil && *setup {
		str = *hotkey + str
	}

	go smoke.rs.Run(str,
		*host, *port, *filePath, !*once, smoke.start, smoke.stop, func() {
			smoke.entered.Store(false)
			smoke.transcribin.Store(true)
		}, func(text string) {
			smoke.copyToClipboard(text, func() {
				if once != nil && *once {
					go func() {
						d.Exit()
						os.Exit(0)
					}()
				}
			})
			smoke.transcribin.Store(false)
		})

	smoke.widget.SetUserDataWidgetHandler(&smoke)

	smoke.widget.ScheduleResize(smoke.width, smoke.height)

	if once != nil && *once {
		go func() {
			time.Sleep(time.Second)
			smoke.entered.Store(true)
			smoke.start <- struct{}{}
		}()
	}

	window.DisplayRun(d)

	smoke.widget.Destroy()
	smoke.window.Destroy()
	d.Destroy()

}
