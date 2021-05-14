package entity

import (
	"encoding/binary"
	"encoding/hex"

	"fmt"
)

func FrameFromBytes(timecode uint16, b []byte) *Frame {
	return &Frame{
		Timecode: timecode,
		Data:     b,
	}
}

type Frame struct {
	Timecode uint16
	Data     []byte
}

func (f *Frame) Validate() error {
	// Check
	return nil
}

func (f *Frame) SetUserID(userID uint16) {
	binary.LittleEndian.PutUint16(f.Data[0:2], userID)
}

func (f *Frame) GetUserID() uint16 {
	return binary.LittleEndian.Uint16(f.Data[0:2])
}

func (f *Frame) TileID() uint16 {
	if len(f.Data) < 3 {
		return 0
	}
	return binary.LittleEndian.Uint16([]byte{f.Data[2], 0})
}

func (f *Frame) GetUserIDHex() string {
	var b []byte
	hex.Encode(b, f.Data[0:2])
	return fmt.Sprintf("%04x", f.GetUserID())
}

func (f *Frame) TimecodeHex() string {
	return fmt.Sprintf("%04x", f.Timecode)
}

func (f *Frame) TileIDHex() string {
	return fmt.Sprintf("%04x", f.TileID())
}
