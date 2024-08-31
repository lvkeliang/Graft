package protocol

import "encoding/json"

type Handshake struct {
	RPCListenPort string
}

func NewHandshake() *Handshake {
	return &Handshake{}
}

// Marshal Handshake 首字节标识为0
func (hs *Handshake) Marshal() ([]byte, error) {
	marshalHS, err := json.Marshal(hs)
	modifiedData := make([]byte, len(marshalHS)+3)

	modifiedData[0] = HandShakeMark

	// 长度字节
	copy(modifiedData[1:3], IntToBytes(len(marshalHS)))

	// Copy the rest of the data from 'marshalAE' to 'modifiedData'
	copy(modifiedData[3:], marshalHS)
	return modifiedData, err
}

// UNMarshal strae需要先去掉首字节(1字节)和长度字节(2字节)
func (hs *Handshake) UNMarshal(strhs []byte) error {
	return json.Unmarshal(strhs, hs)
}

type HandshakeResult struct {
	NodeAddresses []string
	Result        string
}

func NewHandshakeResult() *HandshakeResult {
	return &HandshakeResult{}
}

// Marshal HandshakeResult 首字节标识为0
func (hsr *HandshakeResult) Marshal() ([]byte, error) {
	marshalHS, err := json.Marshal(hsr)
	modifiedData := make([]byte, len(marshalHS)+3)

	// Set the first byte to indicate AppendEntries message type (00000001)
	modifiedData[0] = HandShakeResultMark

	// 长度字节
	copy(modifiedData[1:3], IntToBytes(len(marshalHS)))

	// Copy the rest of the data from 'marshalAE' to 'modifiedData'
	copy(modifiedData[3:], marshalHS)
	return modifiedData, err
}

// UNMarshal strae需要先去掉首字节(1字节)和长度字节(2字节)
func (hsr *HandshakeResult) UNMarshal(strhs []byte) error {
	return json.Unmarshal(strhs, hsr)
}
