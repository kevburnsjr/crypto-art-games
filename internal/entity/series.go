package entity

import (
	"encoding/json"
	"fmt"
)

type Series struct {
	ID       uint16  `json:"id"`
	Name     string  `json:"name"`
	Author   string  `json:"author"`
	Palette  Palette `json:"palette"`
	Boards   []Board `json:"boards"`
	Created  uint32  `json:"created"`
	Active   uint32  `json:"active"`
	Finished uint32  `json:"finished"`
}

type SeriesList []*Series

func (c *Series) ToJson() []byte {
	b, _ := json.Marshal(c)
	return b
}

func (s *Series) IDHex() string {
	return fmt.Sprintf("%04x", s.ID)
}

func SeriesFromJson(b []byte) *Series {
	var res Series
	err := json.Unmarshal(b, &res)
	if err != nil {
		println(err.Error())
		println(string(b))
		return nil
	}
	return &res
}
