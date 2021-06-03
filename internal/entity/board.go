package entity

import (
	"encoding/json"
)

type Board struct {
	ID         uint16 `json:"id"`
	Background string `json:"bg"`
	Width      uint8  `json:"w"`
	Height     uint8  `json:"h"`
	TileSize   uint8  `json:"tsz"`
	Created    uint32 `json:"created"`
	Active     uint32 `json:"active"`
	Finished   uint32 `json:"finished"`
}

type BoardDto struct {
	Board
}

type BoardList []*Board

func (bl BoardList) ToDto() (res []BoardDto) {
	res = make([]BoardDto, len(bl))
	for i, b := range bl {
		res[i] = BoardDto{
			Board: *b,
		}
	}
	return
}

func BoardFromJson(b []byte) *Board {
	var res Board
	err := json.Unmarshal(b, &res)
	if err != nil {
		return nil
	}
	return &res
}
