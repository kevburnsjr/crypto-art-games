package entity

import (
	// "fmt"
	"math"
	"testing"

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameHeaders(t *testing.T) {
	f := &Frame{
		Data: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	f.SetTimestamp(uint32(math.Pow(2, 24)) - 1)
	require.Equal(t, []byte{255, 255, 255}, f.Data[:3])
	require.Equal(t, []byte{255, 255, 255}, f.TimestampBytes())
	require.Equal(t, []byte{255, 255, 255, 0}, f.ID())

	f.Data[3] = byte(255)
	require.Equal(t, []byte{255, 255, 255, 255}, f.ID())

	f.SetUserID(uint32(math.Pow(2, 24)) - 1)
	require.Equal(t, []byte{255, 255, 255}, f.Data[4:7])

	f.SetTimestamp(uint32(420))
	require.Equal(t, uint32(420), f.Timestamp())

	f.SetUserID(uint32(4201))
	require.Equal(t, uint32(4201), f.UserID())

	f.SetTileID(uint8(54))
	require.Equal(t, uint8(54), f.TileID())

	f.SetDeleted(true)
	require.Equal(t, true, f.Deleted())
}
