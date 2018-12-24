package sample

import (
	"bytes"
	empty "net/http"
)

type A struct {
	Data []bytes.Buffer
}

func init() {
	_ = empty.Header{}
}
