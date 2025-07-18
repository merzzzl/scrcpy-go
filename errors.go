package scrcpy

import "errors"

var (
	ErrTextTooLong      = errors.New("inject text > 300 bytes")
	ErrClipboardTooLong = errors.New("clipboard text too long")
	ErrUhidDataTooLong  = errors.New("uhid data exceeds 64 KiB")
	ErrAppNameTooLong   = errors.New("start app name exceeds 255 bytes")
	ErrUhidNameTooLong  = errors.New("uhid name exceeds 255 bytes")
)
