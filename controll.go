package scrcpy

import (
	"bytes"
	"encoding/binary"
	"unicode/utf8"
)

func (c *Client) InjectKeycode(keycode uint32, action byte, repeat, meta uint32) error {
	buf := make([]byte, lenInjectKeycode)
	buf[0] = byte(CtrlInjectKeycode)
	buf[1] = action
	binary.BigEndian.PutUint32(buf[2:], keycode)
	binary.BigEndian.PutUint32(buf[6:], repeat)
	binary.BigEndian.PutUint32(buf[10:], meta)
	_, err := c.controlConn.Write(buf)

	return err
}

func (c *Client) InjectText(text string) error {
	if !utf8.ValidString(text) || len(text) > maxTextLength {
		return ErrTextTooLong
	}

	var buf bytes.Buffer

	_ = buf.WriteByte(byte(CtrlInjectText))
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(text)))
	_, _ = buf.WriteString(text)
	_, err := c.controlConn.Write(buf.Bytes())

	return err
}

func (c *Client) InjectTouch(action byte, pointerID uint64, x, y uint32, pressure uint16, actionButton, buttons uint32) error {
	buf := make([]byte, lenInjectTouch)
	buf[0] = byte(CtrlInjectTouchEvent)
	buf[1] = action
	binary.BigEndian.PutUint64(buf[2:], pointerID)
	binary.BigEndian.PutUint32(buf[10:], x)
	binary.BigEndian.PutUint32(buf[14:], y)
	binary.BigEndian.PutUint16(buf[18:], uint16(c.handshake.Width))
	binary.BigEndian.PutUint16(buf[20:], uint16(c.handshake.Height))
	binary.BigEndian.PutUint16(buf[22:], pressure)
	binary.BigEndian.PutUint32(buf[24:], actionButton)
	binary.BigEndian.PutUint32(buf[28:], buttons)
	_, err := c.controlConn.Write(buf)

	return err
}

func (c *Client) InjectScroll(x, y int32, hscroll, vscroll int16, buttons uint32) error {
	buf := make([]byte, lenInjectScroll)
	buf[0] = byte(CtrlInjectScrollEvent)
	binary.BigEndian.PutUint32(buf[1:], uint32(x))
	binary.BigEndian.PutUint32(buf[5:], uint32(y))
	binary.BigEndian.PutUint16(buf[9:], uint16(c.handshake.Width))
	binary.BigEndian.PutUint16(buf[11:], uint16(c.handshake.Height))
	binary.BigEndian.PutUint16(buf[13:], uint16(hscroll))
	binary.BigEndian.PutUint16(buf[15:], uint16(vscroll))
	binary.BigEndian.PutUint32(buf[17:], buttons)
	_, err := c.controlConn.Write(buf)

	return err
}

func (c *Client) BackOrScreenOn(action byte) error {
	_, err := c.controlConn.Write([]byte{byte(CtrlBackOrScreenOn), action})

	return err
}

func (c *Client) ExpandNotificationPanel() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlExpandNotificationPanel)})

	return err
}

func (c *Client) ExpandSettingsPanel() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlExpandSettingsPanel)})

	return err
}

func (c *Client) CollapsePanels() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlCollapsePanels)})

	return err
}

func (c *Client) GetClipboard(copyKey byte) error {
	_, err := c.controlConn.Write([]byte{byte(CtrlGetClipboard), copyKey})

	return err
}

func (c *Client) SetClipboard(sequence uint64, text string, paste bool) error {
	if !utf8.ValidString(text) || len(text) > maxClipLength {
		return ErrClipboardTooLong
	}

	var buf bytes.Buffer

	_ = buf.WriteByte(byte(CtrlSetClipboard))
	_ = binary.Write(&buf, binary.BigEndian, sequence)
	_ = buf.WriteByte(boolByte(paste))
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(text)))
	_, _ = buf.WriteString(text)
	_, err := c.controlConn.Write(buf.Bytes())

	return err
}

func (c *Client) SetDisplayPower(on bool) error {
	_, err := c.controlConn.Write([]byte{byte(CtrlSetDisplayPower), boolByte(on)})

	return err
}

func (c *Client) RotateDevice() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlRotateDevice)})

	return err
}

func (c *Client) UhidCreate(id, vendorID, productID uint16, name string, data []byte) error {
	if len(name) > 255 {
		return ErrUhidNameTooLong
	}

	if len(data) > 0xffff {
		return ErrUhidDataTooLong
	}

	var buf bytes.Buffer

	_ = buf.WriteByte(byte(CtrlUhidCreate))
	_ = binary.Write(&buf, binary.BigEndian, id)
	_ = binary.Write(&buf, binary.BigEndian, vendorID)
	_ = binary.Write(&buf, binary.BigEndian, productID)
	_ = buf.WriteByte(byte(len(name)))
	_, _ = buf.WriteString(name)
	_ = binary.Write(&buf, binary.BigEndian, uint16(len(data)))
	_, _ = buf.Write(data)
	_, err := c.controlConn.Write(buf.Bytes())

	return err
}

func (c *Client) UhidInput(id uint16, data []byte) error {
	if len(data) > 0xffff {
		return ErrUhidDataTooLong
	}

	var buf bytes.Buffer

	_ = buf.WriteByte(byte(CtrlUhidInput))
	_ = binary.Write(&buf, binary.BigEndian, id)
	_ = binary.Write(&buf, binary.BigEndian, uint16(len(data)))
	_, _ = buf.Write(data)
	_, err := c.controlConn.Write(buf.Bytes())

	return err
}

func (c *Client) UhidDestroy(id uint16) error {
	var buf [3]byte

	buf[0] = byte(CtrlUhidDestroy)
	binary.BigEndian.PutUint16(buf[1:], id)
	_, err := c.controlConn.Write(buf[:])

	return err
}

func (c *Client) OpenHardKeyboardSettings() error {
	_, err := c.controlConn.Write([]byte{byte(CtrlOpenHardKeyboardSettings)})

	return err
}

func (c *Client) StartApp(name string) error {
	if len(name) > 255 {
		return ErrAppNameTooLong
	}

	var buf bytes.Buffer

	_ = buf.WriteByte(byte(CtrlStartApp))
	_ = buf.WriteByte(byte(len(name)))
	_, _ = buf.WriteString(name)
	_, err := c.controlConn.Write(buf.Bytes())

	return err
}

func boolByte(b bool) byte {
	if b {
		return 1
	}

	return 0
}
