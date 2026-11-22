package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type event struct {
	t float64
	d string
}

func load(path string) ([]event, float64) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<28)
	sc.Scan()
	var h struct{ Version int }
	json.Unmarshal(sc.Bytes(), &h)
	var ev []event
	t := 0.0
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e []interface{}
		if json.Unmarshal(line, &e) != nil || len(e) < 3 {
			continue
		}
		dt, _ := e[0].(float64)
		if h.Version == 3 {
			t += dt
		} else {
			t = dt
		}
		if k, _ := e[1].(string); k == "o" {
			s, _ := e[2].(string)
			ev = append(ev, event{t, s})
		}
	}
	return ev, t
}

func ioctl(req uintptr, p unsafe.Pointer) {
	syscall.Syscall(syscall.SYS_IOCTL, 0, req, uintptr(p))
}

func fmtT(s float64) string {
	n := int(s)
	return fmt.Sprintf("%02d:%02d", n/60, n%60)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: castplay file.cast")
		os.Exit(1)
	}
	ev, total := load(os.Args[1])

	var ws struct{ Row, Col, X, Y uint16 }
	ioctl(syscall.TIOCGWINSZ, unsafe.Pointer(&ws))
	rows, cols := int(ws.Row), int(ws.Col)

	var old syscall.Termios
	ioctl(syscall.TCGETS, unsafe.Pointer(&old))
	raw := old
	raw.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.ISIG
	raw.Cc[syscall.VMIN] = 0
	raw.Cc[syscall.VTIME] = 0
	ioctl(syscall.TCSETS, unsafe.Pointer(&raw))

	out := bufio.NewWriterSize(os.Stdout, 1<<16)
	defer func() {
		ioctl(syscall.TCSETS, unsafe.Pointer(&old))
		fmt.Fprintf(out, "\x1b[r\x1b[0m\x1b[%d;1H\x1b[2K\x1b[%d;1H\x1b[2K\n", rows-1, rows)
		out.Flush()
	}()

	i := 0
	pos := 0.0
	paused := false

	seek := func(t float64) {
		if t < 0 {
			t = 0
		}
		if t > total {
			t = total
		}
		pos = t
		fmt.Fprintf(out, "\x1bc\x1b[1;%dr\x1b[H", rows-2)
		i = 0
		for i < len(ev) && ev[i].t <= pos {
			out.WriteString(ev[i].d)
			i++
		}
	}

	draw := func() {
		w := cols - 22
		if w < 10 {
			w = 10
		}
		filled := w
		if total > 0 {
			filled = int(float64(w) * pos / total)
		}
		icon := "> "
		if paused {
			icon = "||"
		}
		bar := make([]byte, w)
		for k := range bar {
			if k < filled {
				bar[k] = '#'
			} else {
				bar[k] = '-'
			}
		}
		help := " Space pause   Left/Right +-5s   Up/Down +-30s   0 restart   q quit"
		if len(help) > cols {
			help = help[:cols]
		}
		fmt.Fprintf(out, "\x1b7\x1b[%d;1H\x1b[2K\x1b[2m%s\x1b[0m\x1b[%d;1H\x1b[2K\x1b[7m %s %s [%s] %s \x1b[0m\x1b8",
			rows-1, help, rows, icon, fmtT(pos), bar, fmtT(total))
		out.Flush()
	}

	seek(0)
	last := time.Now()
	buf := make([]byte, 8)

	for {
		now := time.Now()
		if !paused {
			pos += now.Sub(last).Seconds()
			if pos > total {
				pos = total
			}
		}
		last = now
		for i < len(ev) && ev[i].t <= pos {
			out.WriteString(ev[i].d)
			i++
		}
		draw()

		time.Sleep(30 * time.Millisecond)
		n, _ := syscall.Read(0, buf)
		if n > 0 {
			switch string(buf[:n]) {
			case "q", "\x03":
				return
			case " ":
				paused = !paused
			case "\x1b[C":
				seek(pos + 5)
			case "\x1b[D":
				seek(pos - 5)
			case "\x1b[A":
				seek(pos + 30)
			case "\x1b[B":
				seek(pos - 30)
			case "0":
				seek(0)
			}
			last = time.Now()
		}
	}
}
