package main

import (
	"fmt"
	"log"
	"time"

	"github.com/g3n/engine/audio/al"
	"github.com/g3n/engine/audio/vorbis"
)

type Application struct {
  audioDev *al.Device
  start time.Time
  frame time.Time
  delta time.Duration
  Delay time.Duration
  Close bool
  Nodes []Node
  UI    *UI
}

type Node interface {
  Update(dt time.Duration)
  Destroy()
  Render(*UI) string
}

func NewApplication() Application {
  a := Application{}
  if a.audioDev == nil { a.openDefaultAudioDevice() }
  if a.UI == nil { a.UI = &UI{}; a.UI.Init() }
  return a
}

func (a *Application) Run(update func(dt time.Duration)) {
  a.start = time.Now()
  a.frame = time.Now()
  for {
    now := time.Now()
    a.delta = now.Sub(a.frame)
    a.start = now
    update(a.delta)
    a.UI.Nodes = []string{}
    for _, n := range(a.Nodes) {
      n.Update(a.delta)
      a.UI.Nodes = append(a.UI.Nodes, n.Render(a.UI))
    }
    if a.Close { break }
    a.UI.Render()
    time.Sleep(a.Delay)
  }
  if a.audioDev != nil { al.CloseDevice(a.audioDev) }
  a.Destroy()
}

func (a *Application) Exit() {
  a.Close = true
}

func (a *Application) Destroy() {
    for _, n := range(a.Nodes) { n.Destroy() }
}

func (a *Application) AddNode(node Node) {
  a.Nodes = append(a.Nodes, node)
}

// Source: https://github.com/g3n/engine/blob/master/app/app-desktop.go

// Copyright (c) 2016 The G3N Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

func (a *Application) openDefaultAudioDevice() error {

	// Opens default audio device
	var err error
	a.audioDev, err = al.OpenDevice("")
	if err != nil {
		return fmt.Errorf("opening OpenAL default device: %s", err)
	}
	// Check for OpenAL effects extension support
	var attribs []int
	if al.IsExtensionPresent("ALC_EXT_EFX") {
		attribs = []int{al.MAX_AUXILIARY_SENDS, 4}
	}
	// Create audio context
	acx, err := al.CreateContext(a.audioDev, attribs)
	if err != nil {
		return fmt.Errorf("creating OpenAL context: %s", err)
	}
	// Makes the context the current one
	err = al.MakeContextCurrent(acx)
	if err != nil {
		return fmt.Errorf("setting OpenAL context current: %s", err)
	}
	// Logs audio library versions
	log.Printf("%s version: %s", al.GetString(al.Vendor), al.GetString(al.Version))
	log.Printf("%s", vorbis.VersionString())
	return nil
}
