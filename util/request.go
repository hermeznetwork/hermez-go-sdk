package util

import (
	"bytes"
	"encoding/json"
	"io"
)

// MarshallBody cast object body to io.ReadWriter buffer
func MarshallBody(body interface{}) (buf io.ReadWriter, err error) {
	if body != nil {
		buf = new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(body)
	}
	return
}
