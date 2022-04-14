# SPANTH

## Sample Player (Art-Net) for THeatre

Play audio samples using channel levels.

In simple mode, each channel controls volume of one sample.

- if channel value falls to `0`, playback is stopped, and will restart from the beginning,
- for values `1-2`, playback is paused,
- if value is `4` or less, volume (gain) is `0`,
- if values is `255`, volume is `1.0`,
- between `5` and `254`, volume is calculated according to `vol = (val - 4) / 250`

In advanced mode, each sample is assigned 4 channels: `volume`, `start`, `end`, `mode`.

`start` and `end` controls playback range. `mode`, for now, switches sample looping on (at `255`) and off (at `0`).

## Status

Work In Progress. Proof-of-concept.

## TODO

[ ] MIDI control.
[ ] Webpage with status.
[ ] RPC.

## Internals

This app uses audio part of [g3n](https://github.com/g3n/engine) game engine, which uses OpenAL.

`player.go` is modified file from `g3n`, as original didn't expose any way to seek in loaded file.
