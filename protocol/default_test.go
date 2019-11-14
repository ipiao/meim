package protocol

import "testing"

func TestDefaultBody(t *testing.T) {
	b := make(DefaultBody, 0)
	b.FromData([]byte{1, 2, 3, 4})
	t.Log(b)

	d := b.ToData()
	t.Log(d)
}
