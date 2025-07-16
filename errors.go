package scrcpy

import "errors"

var (
	ErrTextTooLong      = errors.New("text too long for scrcpy packet")
	ErrClipboardTooLong = errors.New("clipboard text too long")
)
