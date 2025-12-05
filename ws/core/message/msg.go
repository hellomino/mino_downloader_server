package message

type Msg interface {
	Raw() []byte
	GetCode() int
}
type H5Message struct {
	Code int
	Data string
}

func (m *H5Message) GetCode() int {
	return m.Code
}

func (m *H5Message) Raw() []byte {
	return []byte(m.Data)
}
