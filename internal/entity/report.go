package entity

import (
	"encoding/json"
	"time"
)

type Report struct {
	Timecode uint16    `json:"timecode"`
	UserID   uint16    `json:"userID"`
	Date     time.Time `json:"date"`
	Reason   string    `json:"string"`
}

type ReportDto struct {
	Report
	Type string `json:"type"`
}

func (u *Report) ToJson() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *Report) ToDto() []byte {
	b, _ := json.Marshal(ReportDto{*u, "report"})
	return b
}
