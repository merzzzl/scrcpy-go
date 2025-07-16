package scrcpy

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func (c *Client) InjectKeycode(keycode uint32, action byte, repeat uint32) error {
	buf := make([]byte, lenInjectKeycode)

	buf[0] = byte(CtrlInjectKeycode)
	buf[1] = action
	binary.BigEndian.PutUint32(buf[2:], keycode)
	binary.BigEndian.PutUint32(buf[6:], repeat)
	binary.BigEndian.PutUint32(buf[10:], 0)

	_, err := c.controlConn.Write(buf)
	if err != nil {
		return fmt.Errorf("inject keycode: %w", err)
	}

	return nil
}

func (c *Client) InjectText(text string) error {
	if len(text) > maxTextLength {
		return ErrTextTooLong
	}

	var buf bytes.Buffer
	if err := buf.WriteByte(byte(CtrlInjectText)); err != nil {
		return fmt.Errorf("inject text type: %w", err)
	}

	if err := binary.Write(&buf, binary.BigEndian, uint16(len(text))); err != nil {
		return fmt.Errorf("inject text length: %w", err)
	}

	if _, err := buf.WriteString(text); err != nil {
		return fmt.Errorf("inject text content: %w", err)
	}

	if _, err := c.controlConn.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("inject text send: %w", err)
	}

	return nil
}

func (c *Client) InjectTouch(action byte, pointerID uint64, x, y uint32, screenWidth, screenHeight, pressure uint16, buttons uint32) error {
	buf := make([]byte, lenInjectTouch)
	buf[0] = byte(CtrlInjectTouchEvent)
	buf[1] = action
	binary.BigEndian.PutUint64(buf[2:], pointerID)
	binary.BigEndian.PutUint32(buf[10:], x)
	binary.BigEndian.PutUint32(buf[14:], y)
	binary.BigEndian.PutUint16(buf[18:], screenWidth)
	binary.BigEndian.PutUint16(buf[20:], screenHeight)
	binary.BigEndian.PutUint16(buf[22:], pressure)
	binary.BigEndian.PutUint32(buf[24:], buttons)

	_, err := c.controlConn.Write(buf)
	if err != nil {
		return fmt.Errorf("inject touch: %w", err)
	}

	return nil
}

func (c *Client) InjectScroll(x, y int32, screenWidth, screenHeight uint16, hscroll, vscroll int32) error {
	buf := make([]byte, lenInjectScroll)
	buf[0] = byte(CtrlInjectScrollEvent)
	binary.BigEndian.PutUint32(buf[1:], uint32(x))
	binary.BigEndian.PutUint32(buf[5:], uint32(y))
	binary.BigEndian.PutUint16(buf[9:], screenWidth)
	binary.BigEndian.PutUint16(buf[11:], screenHeight)
	binary.BigEndian.PutUint32(buf[13:], uint32(hscroll))
	binary.BigEndian.PutUint32(buf[17:], uint32(vscroll))

	_, err := c.controlConn.Write(buf)
	if err != nil {
		return fmt.Errorf("inject scroll: %w", err)
	}

	return nil
}

func (c *Client) BackOrScreenOn() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlBackOrScreenOn)})
	if err != nil {
		return fmt.Errorf("back or screen on: %w", err)
	}

	return nil
}

func (c *Client) ExpandNotificationPanel() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlExpandNotificationPanel)})
	if err != nil {
		return fmt.Errorf("expand notification: %w", err)
	}

	return nil
}

func (c *Client) ExpandSettingsPanel() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlExpandSettingsPanel)})
	if err != nil {
		return fmt.Errorf("expand settings: %w", err)
	}

	return nil
}

func (c *Client) CollapsePanels() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlCollapsePanels)})
	if err != nil {
		return fmt.Errorf("collapse panels: %w", err)
	}

	return nil
}

func (c *Client) GetClipboard() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlGetClipboard), 0})
	if err != nil {
		return fmt.Errorf("get clipboard: %w", err)
	}

	return nil
}

func (c *Client) SetClipboard(text string, paste bool) error {
	if len(text) > maxTextLength {
		return ErrClipboardTooLong
	}

	var buf bytes.Buffer
	if err := buf.WriteByte(byte(CtrlSetClipboard)); err != nil {
		return fmt.Errorf("clipboard type: %w", err)
	}

	if err := binary.Write(&buf, binary.BigEndian, uint16(len(text))); err != nil {
		return fmt.Errorf("clipboard length: %w", err)
	}

	pasteByte := byte(0)
	if paste {
		pasteByte = 1
	}

	if err := buf.WriteByte(pasteByte); err != nil {
		return fmt.Errorf("clipboard paste flag: %w", err)
	}

	if _, err := buf.WriteString(text); err != nil {
		return fmt.Errorf("clipboard content: %w", err)
	}

	if _, err := c.controlConn.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("clipboard send: %w", err)
	}

	return nil
}

func (c *Client) SetScreenPowerMode(mode byte) error {
	_, err := c.controlConn.Write([]byte{byte(CtrlSetScreenPowerMode), mode})
	if err != nil {
		return fmt.Errorf("set screen power mode: %w", err)
	}

	return nil
}

func (c *Client) RotateDevice() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlRotateDevice)})
	if err != nil {
		return fmt.Errorf("rotate device: %w", err)
	}

	return nil
}
