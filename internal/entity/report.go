package entity

import (
	"encoding/json"
)

type Report struct {
	TargetID  uint32 `json:"targetID"`
	BoardID   uint16 `json:"boardID"`
	Timecode  uint32 `json:"timecode"`
	UserID    uint32 `json:"userID"`
	Date      uint32 `json:"date"`
	FrameDate uint32 `json:"frameDate"`
	Reason    string `json:"string"`
}

type ReportDto struct {
	Report
	Type string `json:"type"`
}

func (u *Report) ToJson() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *Report) ToResDto() []byte {
	b, _ := json.Marshal(ReportDto{*u, "report-success"})
	return b
}

func (u *Report) ToDto() []byte {
	b, _ := json.Marshal(ReportDto{*u, "report"})
	return b
}
