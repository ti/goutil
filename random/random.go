package random

import (
	"io"
	"strings"
	"encoding/base64"
	"crypto/rand"
	mathrand "math/rand"
	"crypto/sha512"
	"time"
	"encoding/hex"
)

//NewRandomString default 16 bit string like uuid
func NewRandomString() string {
	return hex.EncodeToString(NewRandomByte(16))
}

//NewRandomStringLen make any len RandomString, len must > 16
func NewRandomByte(len int) []byte {
	uuid := make([]byte, len)
	if _, err := io.ReadFull(rand.Reader, uuid); err != nil {
		panic(err.Error()) // rand should never fail
	}
	if len > 8 {
		uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
		uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	}
	return uuid
}

//NewRandomStringLen make any len RandomString, len must > 16
func NewRandomStringLen(len int) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(NewRandomByte(len)), "=")
}

func NewRandomInt(min, max int) int {
	mathrand.Seed(time.Now().Unix())
	return mathrand.Intn(max - min) + min
}

func EncryptedBySalt(src, salt string)  string {
	sha512Hash := sha512.New()
	sha512Hash.Write([]byte(salt))
	sha512Hash.Write([]byte(src))
	return base64.URLEncoding.EncodeToString(sha512Hash.Sum(nil))
}