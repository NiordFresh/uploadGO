package main

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	barFull  = '█'
	barEmpty = '░'
	barHalf  = '▒'
)

type Progress struct {
	total     int64
	current   int64
	startTime time.Time
	done      bool
	rendered  bool
}

func NewProgress(total int64) *Progress {
	return &Progress{
		total:     total,
		current:   0,
		startTime: time.Now(),
		done:      false,
		rendered:  false,
	}
}

func (p *Progress) Set(current int64) {
	p.current = current
	p.render()
}

func (p *Progress) render() {
	if p.total <= 0 {
		return
	}

	elapsed := time.Since(p.startTime)
	percent := float64(p.current) / float64(p.total) * 100
	barWidth := 30
	filled := int(float64(barWidth) * float64(p.current) / float64(p.total))
	if filled > barWidth {
		filled = barWidth
	}

	var bar strings.Builder
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar.WriteRune(barFull)
		} else if i == filled && !p.done {
			bar.WriteRune(barHalf)
		} else {
			bar.WriteRune(barEmpty)
		}
	}

	speedMB := float64(p.current) / elapsed.Seconds() / 1024 / 1024
	speedStr := fmt.Sprintf("%.1f MB/s", speedMB)

	var eta string
	if speedMB > 0 && !p.done {
		remaining := float64(p.total-p.current) / (speedMB * 1024 * 1024)
		if remaining < 60 {
			eta = fmt.Sprintf("%.0fs", remaining)
		} else {
			eta = fmt.Sprintf("%.0fm", remaining/60)
		}
	} else {
		eta = "---"
	}

	barStr := ColorCyan + bar.String() + ColorReset
	pct := ColorYellow + fmt.Sprintf("%5.1f%%", percent) + ColorReset
	spd := ColorGreen + speedStr + ColorReset
	etaStr := ColorMagenta + eta + ColorReset

	fmt.Printf("\r  %s %s %s %s %s %s  ETA %s   ", barStr, pct, EmojiPlay, spd, EmojiPlay, spd, etaStr)
}

func (p *Progress) Done() {
	if p.done {
		return
	}
	p.current = p.total
	p.done = true

	barWidth := 30
	fullBar := ColorCyan + strings.Repeat(string(barFull), barWidth) + ColorReset
	fmt.Printf("\r  %s %s %s %s %s %s  ETA %s   \n", fullBar,
		ColorYellow+"100.0%%"+ColorReset,
		EmojiPlay,
		ColorGreen+"DONE"+ColorReset,
		EmojiPlay,
		ColorGreen+"---"+ColorReset,
		ColorMagenta+"0s"+ColorReset)
}

type ProgressReader struct {
	reader   io.Reader
	total    int64
	current  int64
	callback func(current, total int64)
}

func NewProgressReader(reader io.Reader, total int64, callback func(current, total int64)) *ProgressReader {
	return &ProgressReader{
		reader:   reader,
		total:    total,
		current:  0,
		callback: callback,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.current += int64(n)
		if pr.callback != nil {
			pr.callback(pr.current, pr.total)
		}
	}
	return n, err
}
