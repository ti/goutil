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

func MustMarshal(v interface{}) []byte{
	if b, err := Marshal(v); err != nil {
		panic(err)
	} else {
		return b
	}
}

func Get(value )  {

}

func Unmarshal(data []byte, v interface{}) error {
	return  gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}
