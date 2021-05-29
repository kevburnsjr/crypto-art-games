package entity

import (
	"encoding/json"
)

type Palette struct {
	Name              string   `json:"name"`
	DisplayName       string   `json:"displayName"`
	AuthorName        string   `json:"author"`
	AuthorDisplayName string   `json:"authorDisplayName"`
	Colors            []string `json:"colors"`
}

type PaletteDto struct {
	Palette
}

type PaletteList []*Palette

func (bl PaletteList) ToDto() (res []PaletteDto) {
	res = make([]PaletteDto, len(bl))
	for i, b := range bl {
		res[i] = PaletteDto{
			Palette: *b,
		}
	}
	return
}

func PaletteFromJson(b []byte) *Palette {
	var res Palette
	err := json.Unmarshal(b, &res)
	if err != nil {
		return nil
	}
	return &res
}
