package log

import (
	"os"
	"strings"
	"time"

	"github.com/docker/docker/pkg/term"
)

type Progressor struct {
	title       string
	stopChan    chan bool
	isActive    bool
	isTerminal  bool
	withColor   bool
	termWidth   int
	clearBuffer []byte
}

func NewProgress() *Progressor {
	var (
		isTerminal  = false
		withColor   = false
		termWidth   int
		clearBuffer []byte
	)
	fd := os.Stdout.Fd()
	isTerminal = term.IsTerminal(fd)
	if isTerminal {
		winsize, err := term.GetWinsize(fd)
		if err != nil {
			termWidth = 80
		} else {
			termWidth = int(winsize.Width)
		}
		if isColorTerminal() {
			withColor = true
		}
		clearBuffer = make([]byte, termWidth)
		for i := range clearBuffer {
			copy(clearBuffer[i:], " ")
		}
		copy(clearBuffer[0:], "\r")
	}
	return &Progressor{
		stopChan:    make(chan bool, 1),
		isActive:    false,
		isTerminal:  isTerminal,
		withColor:   withColor,
		termWidth:   termWidth,
		clearBuffer: clearBuffer,
	}
}

// invoke with: 'log.Progress.Start("some description")'
func (p *Progressor) Start(msg string) {
	if !p.isTerminal || isDebug() {
		Info(msg)
		return
	}
	// stop current progress
	if p.isActive {
		p.Stop()
	}
	p.title = msg
	p.isActive = true
	go p.writer()
}

// invoke with: 'log.Progress.Stop()'
func (p *Progressor) Stop() {
	if p.isActive {
		p.stopChan <- true
		p.isActive = false
		// variant a - remove progress
		os.Stdout.Write(p.clearBuffer)
		os.Stdout.Write([]byte("\r"))
		os.Stdout.Sync()
		// variant b - keep progress and prefix with "+"
		//os.Stdout.Write([]byte("\r+\n"))
		//os.Stdout.Sync()
	}
}

// write out progress in a loop
func (p *Progressor) writer() {
	for {
		select {
		case <-p.stopChan:
			return
		default:
			var out string
			charset := []string{"|", "/", "-", "\\"}
			for i := 0; i < len(charset); i++ {
				out = charset[i]
				if p.withColor {
					out = "\x1b[1;32m" + charset[i] + "\x1b[0m"
				}
				os.Stdout.Write([]byte("\r" + out + " " + p.title))
				os.Stdout.Sync()
				time.Sleep(time.Millisecond * 150)
			}
		}
	}
}

func isColorTerminal() bool {
	matches := []string{
		"Eterm",
		"ansi",
		"color",
		"console",
		"cygwin",
		"dtterm",
		"eterm-color",
		"gnome",
		"konsole",
		"kterm",
		"linux",
		"mach-color",
		"mlterm",
		"putty",
		"rxvt",
		"screen",
		"vt100",
		"xterm",
	}
	if os.Getenv("COLORTERM") != "" {
		return true
	}
	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}
	for _, name := range matches {
		if strings.Contains(term, name) {
			return true
		}
	}
	return false
}
