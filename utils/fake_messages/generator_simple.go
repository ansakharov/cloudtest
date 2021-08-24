package fake_messages

type BytesGenerator interface {
	Gen(n int) []byte
}

func (g *SimpleGenerator) Gen(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return b
}

type SimpleGenerator struct{}
