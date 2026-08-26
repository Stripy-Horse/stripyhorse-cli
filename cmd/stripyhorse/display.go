package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"os"
	"strings"
)

// isTTY reports whether f is a terminal (vs a pipe/file).
func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

type imageProtocol int

const (
	protoHalfBlock imageProtocol = iota
	protoKitty
	protoITerm
)

// detectProtocol picks the best inline-image protocol the terminal supports,
// falling back to unicode half-blocks (works in any truecolor terminal).
func detectProtocol() imageProtocol {
	term := os.Getenv("TERM")
	switch {
	case os.Getenv("KITTY_WINDOW_ID") != "",
		os.Getenv("GHOSTTY_RESOURCES_DIR") != "",
		strings.Contains(term, "kitty"),
		term == "xterm-ghostty":
		return protoKitty
	}
	switch os.Getenv("TERM_PROGRAM") {
	case "iTerm.app", "WezTerm", "vscode":
		return protoITerm
	}
	return protoHalfBlock
}

// displayImage renders PNG bytes inline in the terminal.
func displayImage(w io.Writer, png []byte, maxCols int) error {
	switch detectProtocol() {
	case protoKitty:
		return writeKitty(w, png)
	case protoITerm:
		return writeITerm(w, png)
	default:
		return writeHalfBlocks(w, png, maxCols)
	}
}

// writeITerm uses the iTerm2 inline-image protocol (also WezTerm, VS Code).
func writeITerm(w io.Writer, png []byte) error {
	b64 := base64.StdEncoding.EncodeToString(png)
	_, err := fmt.Fprintf(w, "\x1b]1337;File=inline=1;size=%d;preserveAspectRatio=1:%s\x07\n", len(png), b64)
	return err
}

// writeKitty uses the Kitty graphics protocol (Kitty, Ghostty), chunking the
// base64 payload into 4096-byte pieces as the protocol requires.
func writeKitty(w io.Writer, png []byte) error {
	b64 := base64.StdEncoding.EncodeToString(png)
	const chunk = 4096
	for i := 0; i < len(b64); i += chunk {
		end := i + chunk
		if end > len(b64) {
			end = len(b64)
		}
		more := 0
		if end < len(b64) {
			more = 1
		}
		var err error
		if i == 0 {
			// a=T transmit+display, f=100 PNG data.
			_, err = fmt.Fprintf(w, "\x1b_Ga=T,f=100,m=%d;%s\x1b\\", more, b64[i:end])
		} else {
			_, err = fmt.Fprintf(w, "\x1b_Gm=%d;%s\x1b\\", more, b64[i:end])
		}
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

// writeHalfBlocks renders the image with the ▀ half-block: each character cell
// shows two vertically-stacked pixels (foreground = top, background = bottom)
// via truecolor SGR. Works in any terminal with 24-bit color.
func writeHalfBlocks(w io.Writer, pngBytes []byte, maxCols int) error {
	img, _, err := image.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return fmt.Errorf("decoding rendered PNG: %w", err)
	}
	b := img.Bounds()
	iw, ih := b.Dx(), b.Dy()
	if iw == 0 || ih == 0 {
		return nil
	}
	if maxCols <= 0 {
		maxCols = 80
	}
	cols := iw
	if cols > maxCols {
		cols = maxCols
	}
	f := float64(iw) / float64(cols) // source px per output col (and per output px row)
	rowsPx := int(float64(ih) / f)
	rows := (rowsPx + 1) / 2

	at := func(x, y int) (uint8, uint8, uint8) {
		sx := b.Min.X + int(float64(x)*f)
		sy := b.Min.Y + int(float64(y)*f)
		if sx >= b.Max.X {
			sx = b.Max.X - 1
		}
		if sy >= b.Max.Y {
			sy = b.Max.Y - 1
		}
		r, g, bl, _ := img.At(sx, sy).RGBA()
		return uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8)
	}

	var sb strings.Builder
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			tr, tg, tb := at(col, row*2)
			if row*2+1 < rowsPx {
				br, bg, bb := at(col, row*2+1)
				fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀", tr, tg, tb, br, bg, bb)
			} else {
				fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm\x1b[49m▀", tr, tg, tb)
			}
		}
		sb.WriteString("\x1b[0m\n")
	}
	_, err = io.WriteString(w, sb.String())
	return err
}
