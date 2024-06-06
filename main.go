package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/marf41/artnet"
)

type Config struct {
	Port         int
	Physical     int
	Address      int
	Channels     int
	Dump         int
	Advanced     int
	Preview      int
	PreviewSplit int
	BarWidth     int
	Sample       []ConfigSample
}

type ConfigSample struct {
	Name string
	File string
	Loop bool
}

func parse(s artnet.ArtNet) {
	start := conf.Address - 1
	adv := conf.Advanced > 0
	chn := 1
	if adv {
		chn = 4
	}
	if conf.Physical < 0 || conf.Physical == int(s.Physical) {
		if conf.Port < 0 || conf.Port == int(s.Port) {
			to := start + conf.Channels*chn
			if to > int(s.Length) {
				to = int(s.Length)
			}
			// log.Printf("%d-%d: %s\n", start, to, s.Channels(start, to, " "))
			ls := len(samples)
			ld := len(s.Data)
			if conf.Advanced > 0 {
				if ld < ls*4 {
					ls = ld / 4
				}
				n := make([]int, ls)
				dn := 0
				for i := range n {
					samples[i].SetValue(uint(s.Data[dn]))
					samples[i].SetRange(uint(s.Data[dn+1]), uint(s.Data[dn+2]))
					samples[i].SetMode(uint(s.Data[dn+3]))
					dn += 4
				}
			} else {
				if ld < ls {
					ls = ld
				}
				n := make([]int, ls)
				for i := range n {
					samples[i].SetValue(uint(s.Data[i]))
				}
			}
		}
	}
}

func createExampleConfig() {
	buf := new(bytes.Buffer)
	exConf := Config{
		Address:  1,
		Channels: 1,
	}
	err := toml.NewEncoder(buf).Encode(exConf)
	if err != nil {
		log.Println("Error creating example config file.")
		return
	}
	os.WriteFile("spanth.toml", buf.Bytes(), 0666)
}

var conf Config
var app Application
var samples []*Sample

func main() {
	log.Println("Start.")
	_, err := toml.DecodeFile("spanth.toml", &conf)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No config file. Creating minimal example.")
			createExampleConfig()
		}
		log.Fatalln(err)
	}
	if conf.Address == 0 {
		log.Fatalln("No address defined.")
	}
	if conf.Channels == 0 {
		log.Fatalln("No channels defined.")
	}
	if conf.Preview == 0 {
		conf.Preview = 16
	}
	if conf.BarWidth == 0 {
		conf.BarWidth = 32
	}
	tn := conf.Preview
	if conf.Advanced > 0 {
		fmt.Println("Advanced mode active.\nChannel layout:\tch1volume, ch1from, ch1to, ch1mode, ch2volume, ch2from, ...\nMode: 0-127 - stop after, 127-255 - loop.")
	}

	app = NewApplication()
	app.UI.Type = make([]string, tn)
	app.openDefaultAudioDevice()

	samples = make([]*Sample, len(conf.Sample))
	ti := 0
	for i, cs := range conf.Sample {
		s := &Sample{}
		s.Load(cs.Name, cs.File)
		s.Player.SetLooping(cs.Loop)
		samples[i] = s
		app.AddNode(s)
		if ti < conf.Preview {
			if conf.Advanced > 0 {
				app.UI.Type[ti] = fmt.Sprintf("V%d", i+1)
				app.UI.Type[ti+1] = fmt.Sprintf("B%d", i+1)
				app.UI.Type[ti+2] = fmt.Sprintf("E%d", i+1)
				app.UI.Type[ti+3] = fmt.Sprintf("M%d", i+1)
				ti += 4
			} else {
				app.UI.Type[ti] = fmt.Sprintf("V%d", i+1)
				ti++
			}
		}
	}

	defer app.Exit()
	app.Run(func(dt time.Duration) {
		an, err := artnet.GetAndParse(conf.Dump > 0)
		if err != nil {
			log.Println(err)
		} else {
			parse(an)
		}
		// log.Printf("%f / %f\n", s.Player.CurrentTime(), s.Player.TotalTime())
		if an.Physical == uint8(conf.Physical) && an.Port == uint16(conf.Port) {
			app.UI.Data = an.Data[0:tn]
		}
	})
}
