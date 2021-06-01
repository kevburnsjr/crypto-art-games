package socket

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Msg wsmessage

func (m *Msg) Raw(channel_id string) wsmessage {
	return wsmessage{m.msgType, channel_id, m.data}
}

type wsmessage struct {
	msgType    int
	channel_id string
	data       []byte
}

func Message(channel_id string, text string) wsmessage {
	return wsmessage{websocket.TextMessage, channel_id, []byte(channel_id + "-" + text)}
}

func JsonMessage(channel_id string, data map[string]interface{}) wsmessage {
	data["channel_id"] = channel_id
	json_bytes, _ := json.Marshal(data)
	return wsmessage{websocket.TextMessage, channel_id, json_bytes}
}

func JsonMessagePure(channel_id string, data interface{}) wsmessage {
	json_bytes, _ := json.Marshal(data)
	return wsmessage{websocket.TextMessage, channel_id, json_bytes}
}

func TextMsgFromBytes(channel_id string, b []byte) wsmessage {
	return wsmessage{websocket.TextMessage, channel_id, b}
}

func BinaryMsgFromBytes(channel_id string, b []byte) wsmessage {
	return wsmessage{websocket.BinaryMessage, channel_id, b}
}

func NewJsonRes(data interface{}) *Msg {
	var json_bytes []byte
	if d, ok := data.([]byte); ok {
		json_bytes = d
	} else {
		json_bytes, _ = json.Marshal(data)
	}
	msg := Msg(wsmessage{websocket.TextMessage, "", json_bytes})
	return &msg
}
