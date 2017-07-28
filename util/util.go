package util

import (
	"bytes"
	"encoding/gob"
)

func Marshal(v interface{}) (b []byte, e error){
	var buf bytes.Buffer
	e = gob.NewEncoder(&buf).Encode(v)
	b = buf.Bytes()
	return
}


func Unmarshal(data []byte, v interface{}) error {
	return  gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}
