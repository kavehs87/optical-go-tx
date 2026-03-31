package main

import (
	"opticaltransfercli/filepicker" // Internal package reference
	"time"
)

type model struct {
	state      state
	filePicker filepicker.Model

	// Data to encode
	rawData   []byte
	bitChunks []bitChunk

	// Pagination
	currentIndex int // current frame/page index
	frameSize    int // how many chunks fit in one frame

	// Settings
	delay time.Duration
	loop  bool

	// UI
	width  int
	height int
	styles styles
}

func initialModel() model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{} // All allowed

	return model{
		state:      stateIdle,
		filePicker: fp,
		delay:      1000 * time.Millisecond,
		loop:       true, // Loop by default
		styles:     newStyles(true),
	}
}
