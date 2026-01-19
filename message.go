package koduck

import (
	"encoding/json"
)

// Message 简单协议：| length(uint32) | cmd(uint16) | data |
type Message struct {
	Conn   *Conn           `json:"-"`
	Method uint32          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// func Encode(msg *Message) ([]byte, error) {
// 	buf := new(bytes.Buffer)
// 	totalLen := uint32(2 + len(msg.Data))
// 	if err := binary.Write(buf, binary.BigEndian, totalLen); err != nil {
// 		return nil, err
// 	}
// 	if err := binary.Write(buf, binary.BigEndian, msg.Cmd); err != nil {
// 		return nil, err
// 	}
// 	buf.Write(msg.Data)
// 	return buf.Bytes(), nil
// }

func DecodeMessage(raw []byte) (*Message, error) {
	msg := &Message{}
	if err := json.Unmarshal(raw, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func EncodeMessage(method uint32, v any) (*Message, error) {
	params, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	msg := &Message{
		Method: method,
		Params: params,
	}
	return msg, nil
}

func (msg *Message) Decode(raw []byte, v any) error {
	if err := json.Unmarshal(raw, v); err != nil {
		return err
	}
	return nil
}

func (msg *Message) Encode() ([]byte, error) {
	return json.Marshal(msg)
}
