package message

type Msg interface {
	Raw() []byte
	GetCode() int
}
type H5Message struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

func (m *H5Message) GetCode() int {
	return m.Code
}

func (m *H5Message) Raw() []byte {
	return []byte(m.Data)
}
