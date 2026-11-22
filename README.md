# castplay

A small terminal player for [asciinema](https://asciinema.org) `.cast` recordings with a seekbar, current time, and total length. Single static binary, standard library only.

```
 >  01:23 [########--------------------------] 05:40
```

## Features

- Status bar on the bottom row: play/pause icon, elapsed time, progress bar, total length
- Seek forward and back
- Supports asciicast v2 and v3
- Optional double-click integration for `.cast` files on Linux desktops

## Controls

| Key | Action |
|-----|--------|
| Space | Pause / resume |
| Right / Left | Seek +5s / -5s |
| Up / Down | Seek +30s / -30s |
| 0 | Restart |
| q or Ctrl+C | Quit |

## Requirements

- Linux (uses Linux terminal ioctls; not portable to macOS or Windows as written)
- Go 1.18 or newer to build (`sudo apt install golang-go` on Debian/Ubuntu)

Go is only needed to build. The resulting binary has no runtime dependencies.

## Build

```bash
git clone https://github.com/YOUR_USERNAME/castplay.git
cd castplay
go build -o castplay .
./castplay path/to/recording.cast
```

## Install

```bash
./install.sh
```

This builds the binary into `~/.local/bin/castplay`, then registers the `.cast` MIME type and a desktop launcher so double-clicking a `.cast` file opens it in a terminal with castplay. Make sure `~/.local/bin` is on your `PATH`. You may need to restart your file manager (for example `nautilus -q`) or log out and back in.

To uninstall:

```bash
rm ~/.local/bin/castplay ~/.local/bin/castplay-launch ~/.local/share/applications/castplay.desktop
```

## Limitations

- Seeking replays the recording from the start up to the target time, so scrubbing in very long recordings can flicker.
- The whole recording is loaded into memory at startup; very large files take a moment to load and nothing is drawn until loading finishes.
- Terminal size is read once at startup and window resizes aren't handled. Use a terminal at least as large as the one the recording was made in, or output may wrap.
- Only output events (`"o"`) are played; input and marker events are ignored.

## License

MIT
