package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

var BARS = [...]rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
var SHADED_BLOCKS = [...]rune{' ', '░', '▒', '▓', '█'}

const ICON_PLAY = string(rune(0xf04b))
const ICON_PAUSE = string(rune(0xf04c))
const ICON_STOP = string(rune(0xf04d))
const ICON_SPEAKER = string(rune(0xf09f))

type UI struct {
	Data    []uint8
	Type    []string
	Nodes   []string
	Message string
	Clear   bool
}

func (ui *UI) Init() {
	clear()
}

func (ui *UI) Render() {
	if ui.Clear {
		clear()
		ui.Clear = false
	}
	home()
	fmt.Print("SPANTH")
	if len(ui.Data) == 0 {
		fmt.Print(" - NO SAMPLES")
	}
	fmt.Println()
	fmt.Print("   ")
	ui.line(func(i int) { fmt.Printf(" %2c", BARS[mapChannelToBar(ui.Data[i])]) })
	ui.line(func(i int) { fmt.Printf(" %2d", i+1) })
	ui.line(func(i int) { fmt.Printf(" %2s", mapChannelToValue(ui.Data[i])) })
	ui.line(func(i int) { fmt.Printf(" %2s", ui.Type[i]) })
	fmt.Print("\n\n")
	for _, s := range ui.Nodes {
		fmt.Println(s)
	}
	fmt.Print("\n\n")
	fmt.Println(ui.Message)
}

func (ui UI) Update(dt time.Duration) {
	ui.Render()
}

func (ui UI) line(fn func(i int)) {
	fmt.Print("\n ")
	split := conf.PreviewSplit
	for i := range ui.Data {
		fn(i)
		if split > 0 && ((i+1)%split) == 0 {
			fmt.Print("\n ")
		}
	}
}

func (ui UI) Bar(value float64, width int) string {
	b := int(math.Ceil(value * float64(width)))
	f := strings.Repeat(string(SHADED_BLOCKS[4]), b)
	e := strings.Repeat(string(SHADED_BLOCKS[1]), width-b)
	return f + e
}

func (ui UI) BarRange(from, to, total float64, width int, full bool) string {
	cb := '-'
	cmb := '['
	cmm := ' '
	cme := ']'
	if full {
		cb = SHADED_BLOCKS[1]
		cmb = SHADED_BLOCKS[4]
		cmm = SHADED_BLOCKS[4]
		cme = SHADED_BLOCKS[4]
	}
	ce := cb
	bvalue := from / total
	b := int(math.Ceil(bvalue * float64(width)))
	mvalue := to / total
	m := 0
	if to > 0 {
		m = int(math.Floor(mvalue*float64(width))) + 1 - b
	}
	e := width - (b + m)
	sb := strings.Repeat(string(cb), b)
	s := sb
	if m > 0 {
		if m >= 1 {
			s += string(cmb)
		}
		if m >= 3 {
			s += strings.Repeat(string(cmm), m-2)
		}
		if m >= 2 {
			s += string(cme)
		}
	}
	s += strings.Repeat(string(ce), e)
	return s
}

func clear() {
	print("\033[H\033[2J")
}

func home() {
	print("\033[H")
}

func mapChannelToBar(v uint8) int {
	if v == 0 {
		return 0
	}
	if v == 255 {
		return 8
	}
	var vf float64
	vf = (float64(v) - 0) / 254
	return int(math.Ceil(vf * 7))
}

func mapChannelToValue(v uint8) string {
	if v == 0 {
		return "--"
	}
	if v == 255 {
		return "FL"
	}
	return fmt.Sprintf("%02d", int(math.Ceil(float64(v)/2.57)))
}
