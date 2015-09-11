package log

import (
	"os"
	"strings"
	"time"

	"github.com/docker/docker/pkg/term"
)

const (
	successChar = "✓"
	failureChar = "!"
	green       = "\x1b[32m"
	red         = "\x1b[31m"
	reset       = "\x1b[0m"
)

type CLISpinner struct {
	title       string
	stopChan    chan bool
	isActive    bool
	isTerminal  bool
	withColor   bool
	termWidth   int
	clearBuffer []byte
}

func TerminalSpinner() *CLISpinner {
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

	return &CLISpinner{
		stopChan:    make(chan bool, 1),
		isActive:    false,
		isTerminal:  isTerminal,
		withColor:   withColor,
		termWidth:   termWidth,
		clearBuffer: clearBuffer,
	}
}

func (sp *CLISpinner) Start(msg string) {
	if !sp.isTerminal {
		// Redirect message to debug logger if no terminal attached to stdout
		Debug(msg)
		return
	}

	if sp.isActive {
		sp.Stop()
	}
	sp.title = msg
	sp.isActive = true
	go sp.writer()
}

func (sp *CLISpinner) Stop() {
	if sp.isActive {
		sp.stopChan <- true
		sp.isActive = false
		prefix := successChar
		if sp.withColor {
			prefix = green + prefix + reset
		}
		os.Stdout.Write(sp.clearBuffer)
		os.Stdout.Write([]byte("\r" + prefix + " " + sp.title + "\n"))
		os.Stdout.Sync()
	}
}

func (sp *CLISpinner) StopWithError() {
	if sp.isActive {
		sp.stopChan <- true
		sp.isActive = false
		prefix := failureChar
		if sp.withColor {
			prefix = red + prefix + reset
		}
		os.Stdout.Write(sp.clearBuffer)
		os.Stdout.Write([]byte("\r" + prefix + " " + sp.title + "\n"))
		os.Stdout.Sync()
	}
}

func (sp *CLISpinner) Clear() {
	if sp.isActive {
		sp.stopChan <- true
		sp.isActive = false
		os.Stdout.Write(sp.clearBuffer)
		os.Stdout.Write([]byte("\r"))
		os.Stdout.Sync()
	}
}

func (sp *CLISpinner) writer() {
	for {
		select {
		case <-sp.stopChan:
			return
		default:
			var out string
			charset := []string{"|", "/", "-", "\\"}
			for i := 0; i < len(charset); i++ {
				out = charset[i]
				if sp.withColor {
					out = green + charset[i] + reset
				}
				os.Stdout.Write([]byte("\r" + out + " " + sp.title))
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
