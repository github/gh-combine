package cmd

import (
	"time"

	"github.com/briandowns/spinner"
)

type Spinner struct {
	spinner *spinner.Spinner
}

func NewSpinner(message string) *Spinner {
	dotStyle := spinner.CharSets[11]
	color := spinner.WithColor("fgCyan")
	duration := 60 * time.Millisecond
	spinner := spinner.New(dotStyle, duration, color)
	if message != "" {
		spinner.Suffix = " " + message
	}
	spinner.Start()

	return &Spinner{
		spinner: spinner,
	}
}

func (s *Spinner) Stop() {
	defer func() {
		if r := recover(); r != nil {
			s.spinner.Stop()
			panic(r) // Re-raise the panic after stopping the spinner
		}
	}()
	s.spinner.Stop()
}

func (s *Spinner) Suffix(message string) {
	s.spinner.Suffix = " " + message
}
