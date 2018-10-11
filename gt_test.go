package gt

import (
	"testing"
)

var gtClient = NewCt(*DefaultConfig)

func TestGt_Register(t *testing.T) {
	t.Log(gtClient.Register("", ""))
}
