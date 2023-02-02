package values

type MarshallerTo interface {
	MarshalToSizedBuffer([]byte) (int, error)
}
