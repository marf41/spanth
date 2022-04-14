package main

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"
)

type Sample struct {
  Node
  ID int
  Name string
  Player *Player
  Value uint
  Started bool
  Min float64
  Max float64
  Mode uint
}

var SID = 1

func (s *Sample) Load(name, filename string) error {
  player, err := NewPlayer(filename)
  if err != nil { return err }
  s.Player = player
  s.Name = name
  if len(name) == 0 { s.Name = filename }
  s.ID = SID
  SID++
  // log.Println(s.Player.Play())
  return nil
}

func (s *Sample) SetValue(value uint) {
  if s.Player == nil { log.Println("Player is nil."); return }
  if s.Player.AF == nil { log.Println("File is nil.") }
  if s.Player == nil || s.Player.AF == nil { log.Println("No sample loaded."); return }
  s.Value = value
  // log.Printf("New value for sample %q: %d.\n", s.Name, s.Value)
  if value > 255 { return }
  if value == 0 { s.Stop(); s.Player.SetGain(0); return }
  if value <= 2 { s.Player.Pause(); return }
  if !s.Started || s.Player.Paused() { s.Player.Play(); s.Started = true }
  if value <= 4 { s.Player.SetGain(0); return }
  if value >=255 { s.Player.SetGain(1.0); return }
  // value is 3-253
  var v float32
  v = float32(value - 4) / 250.0
  s.Player.SetGain(v)
}

func (s *Sample) SetRange(min, max uint) {
  total := s.Player.TotalTime()
  s.Min = total * float64(min) / 255
  if max == 0 { max = 255 }
  s.Max = total * float64(max) / 255
  app.UI.Clear = true
}

func (s *Sample) SetMode(mode uint) {
  s.Mode = mode
  if mode == 0 { s.Player.SetLooping(false) }
  if mode == 255 { s.Player.SetLooping(true) }
}

func (s Sample) Render(ui *UI) string {
  bar := ui.BarRange(s.Min, s.Player.CurrentTime(), s.Player.TotalTime(), conf.BarWidth, true)
  loop := " "
  if s.Player.Looping() { loop = "L" }
  name := s.Name
  if len(name) > 16 { name = name[0:15] + "…" }
  adv := strings.Repeat(" ", conf.BarWidth + 20)
  hasadv := conf.Advanced > 0 && (s.Min != 0 || (s.Max != 0 && s.Max != s.Player.TotalTime()))
  // hasadv = true
  if hasadv {
    advbar := ui.BarRange(s.Min, s.Max, s.Player.TotalTime(), conf.BarWidth, false)
    volbar := ui.Bar(float64(s.Player.Gain()), 14)
    gain := "V---"
    g := s.Player.Gain()
    if g > 0 {
      gain = fmt.Sprintf("%.2f", g)
      if g == 1.0 { gain = "VMAX" }
    }
    adv = fmt.Sprintf(" %s  %s %s %s %s %s\n", " ", gain, volbar, " ", advbar, s.Range())
  }
  return fmt.Sprintf(" %s %2d. %-16s %s %s %s\n", s.Icon(), s.ID, name, loop, bar, s.Time()) + adv
}

func (s Sample) Percent() float64 {
  return float64(s.Player.CurrentTime()) / float64(s.Player.TotalTime())
}

func (s Sample) Time() string {
  if s.Player == nil { return "nil" }
  return fmt.Sprintf("%s / %s", timeParse(s.Player.CurrentTime()), timeParse(s.Player.TotalTime()))
  // return fmt.Sprintf("%s / %s", timeParse(90 * 60.1), timeParse(1.002))
}

func (s Sample) Range() string {
  return fmt.Sprintf("%s - %s", timeParse(s.Min), timeParse(s.Max))
}

func (s Sample) Seek(pos float64) {
  info := s.Player.AF.Info()
  seek := uint(math.Ceil(pos * float64(info.SampleRate)))
  s.Player.AF.Seek(seek)
}

func (s *Sample) Stop() {
  s.Player.Stop()
  s.Seek(0)
  s.Started = false
}

func (s Sample) Update(dt time.Duration) {
  // log.Printf("%f / %f\n", s.Player.CurrentTime(), s.Player.TotalTime())
  curr := s.Player.CurrentTime()
  if !s.Player.Playing() { return }
  if curr < s.Min { s.Seek(s.Min) }
  if curr > s.Max {
    if s.Player.Looping() { s.Seek(s.Min) } else { s.Stop() }
  }
}

func (s Sample) Destroy() {
  s.Player.Dispose()
}

func (s Sample) Icon() string {
  if s.Player.Playing() { return ICON_PLAY }
  if s.Player.Paused() { return ICON_PAUSE }
  return ICON_STOP
}

func timeParse(ts float64) string {
  t := time.Duration(ts * float64(time.Second))
  z := time.Unix(0, 0).UTC()
  return z.Add(t).Format("04:05.000")
}
