package entity

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

var le = binary.LittleEndian

type Frame struct {
	Data []byte
}

func FrameFromBytes(b []byte) *Frame {
	if len(b) == 0 {
		return nil
	}
	return &Frame{
		Data: b,
	}
}

func (f *Frame) Validate() error {
	// TODO - Waiting to implement this until frame encoding is stable
	return nil
}

func (f *Frame) ID32() uint32 {
	return f.Timestamp()*256 + uint32(f.TileID())
}

func (f *Frame) ID() []byte {
	var idBytes = make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, f.ID32())
	return idBytes
}

func (f *Frame) Timestamp() uint32 { return f.getUint24(0) }
func (f *Frame) TileID() uint8     { return f.getUint8(3) }
func (f *Frame) UserID() uint32    { return f.getUint24(4) }
func (f *Frame) Deleted() bool     { return f.getUint8(7)&4 > 0 }

func (f *Frame) SetTimestamp(timestamp uint32) { f.setUint24(0, timestamp) }
func (f *Frame) SetTileID(tileID uint8)        { f.setUint8(3, tileID) }
func (f *Frame) SetUserID(userID uint32)       { f.setUint24(4, userID) }
func (f *Frame) SetDeleted(v bool) {
	if v {
		f.Data[7] = f.Data[7] | 32
	} else {
		if f.Deleted() {
			f.Data[7] = f.Data[7] ^ 32
		}
	}
}

func (f *Frame) ToBytes() []byte { return f.Data }

func (f *Frame) IDHex() string {
	return fmt.Sprintf("%08x", f.ID())
}

func (f *Frame) TileIDHex() string {
	return fmt.Sprintf("%02x", f.TileID())
}

func (f *Frame) UserIDHex() string {
	return fmt.Sprintf("%06x", f.UserID())
}

func (f *Frame) DataHex() string {
	return fmt.Sprintf("%x", f.Data)
}

func (f *Frame) TimestampBytes() []byte {
	var b = make([]byte, 4)
	binary.BigEndian.PutUint32(b[0:4], f.Timestamp())
	return b[1:]
}

func (f *Frame) getUint8(o int) uint8   { return uint8(bits.Reverse16(le.Uint16([]byte{0, f.Data[o]}))) }
func (f *Frame) getUint16(o int) uint16 { return bits.Reverse16(le.Uint16(f.Data[o : o+2])) }
func (f *Frame) getUint24(o int) uint32 {
	return bits.Reverse32(le.Uint32(append([]byte{0}, f.Data[o:o+3]...)))
}
func (f *Frame) getUint32(o int) uint32 { return bits.Reverse32(le.Uint32(f.Data[o : o+4])) }

func (f *Frame) setUint8(o int, n uint8) {
	var b = make([]byte, 2)
	le.PutUint16(b, bits.Reverse16(uint16(n)))
	f.Data[o] = b[1]
}
func (f *Frame) setUint16(o int, n uint16) { le.PutUint16(f.Data[o:o+2], bits.Reverse16(n)) }
func (f *Frame) setUint24(o int, n uint32) {
	var b = make([]byte, 4)
	le.PutUint32(b, bits.Reverse32(n))
	copy(f.Data[o:o+3], b[1:])
}
func (f *Frame) setUint32(o int, n uint32) { le.PutUint32(f.Data[o:o+4], bits.Reverse32(n)) }
