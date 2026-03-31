package main

import "time"

// Msg types
type tickMsg time.Time
type errorMsg error

// State enum
type state int

const (
	stateIdle state = iota
	statePicking
	stateEncoding
)

type bitChunk uint8 // 0-7 (3 bits)