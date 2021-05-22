package entity

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

var le = binary.LittleEndian

func FrameFromBytes(b []byte) *Frame {
	return &Frame{
		Data: b,
	}
}

type Frame struct {
	Data []byte
}

func (f *Frame) Validate() error {
	// TODO - Waiting to implement this until frame encoding is stable
	return nil
}

func (f *Frame) Timecode() uint16  { return f.getUint16(0) }
func (f *Frame) UserID() uint16    { return f.getUint16(2) }
func (f *Frame) TileID() uint16    { return f.getUint8(4) }
func (f *Frame) Timestamp() uint16 { return f.getUint16(6) }
func (f *Frame) Timecheck() uint32 { return f.getUint32(8) }

func (f *Frame) SetTimecode(timecode uint16)   { f.setUint16(0, timecode) }
func (f *Frame) SetUserID(userID uint16)       { f.setUint16(2, userID) }
func (f *Frame) SetTimestamp(timestamp uint16) { f.setUint16(6, timestamp) }
func (f *Frame) SetTimecheck(timecheck uint32) {
	f.Data = append(f.Data[0:12], f.Data[8:]...)
	f.setUint32(8, timecheck)
}

func (f *Frame) ToBytes() []byte { return f.Data }

func (f *Frame) TimecodeHex() string {
	return fmt.Sprintf("%04x", f.Timecode())
}

func (f *Frame) TileIDHex() string {
	return fmt.Sprintf("%04x", f.TileID())
}

func (f *Frame) DataHex() string {
	return fmt.Sprintf("%04x", f.Data)
}

func (f *Frame) getUint8(o int) uint16  { return bits.Reverse16(le.Uint16([]byte{0, f.Data[o]})) }
func (f *Frame) getUint16(o int) uint16 { return bits.Reverse16(le.Uint16(f.Data[o : o+2])) }
func (f *Frame) getUint32(o int) uint32 { return bits.Reverse32(le.Uint32(f.Data[o : o+4])) }

func (f *Frame) setUint16(o int, n uint16) { le.PutUint16(f.Data[o:o+2], bits.Reverse16(n)) }
func (f *Frame) setUint32(o int, n uint32) { le.PutUint32(f.Data[o:o+4], bits.Reverse32(n)) }
