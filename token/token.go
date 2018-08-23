// Package Token implements a base64 token from ObjectID
package token

import (
	"io"
	"crypto/rand"
	"fmt"
	"encoding/binary"
	"time"
	"sync/atomic"
	"crypto/rc4"
	"crypto/md5"
	"os"
	"encoding/base64"
	"math/big"
	"errors"
	"reflect"
	"bytes"
	"strings"
)

// NewEncoding returns a new token Encoding defined by the given secret,
// secret should be less 256 bytes
func NewEncoding(secret string) *Encoding {
	if len(secret) == 0 {
		return &Encoding{}
	}
	cipher, err := rc4.NewCipher([]byte(secret))
	if err != nil {
		panic(err)
	}
	ciphere, err := rc4.NewCipher([]byte(secret))
	if err != nil {
		panic(err)
	}
	return &Encoding{
		withSecret: true,
		dconder:    cipher,
		encoder:    ciphere,
	}
}

// An Encoding is token encoding/decoding scheme, defined by a rc4 Secret
type Encoding struct {
	withSecret bool
	encoder    *rc4.Cipher
	dconder    *rc4.Cipher
}

// Encode encode a token to a string
// meta type is better fo int64, int, string and []byte, you can use map[string][]string
func (e *Encoding) Encode(userID int64, meta interface{}, expires time.Time) string {
	return base64.RawURLEncoding.EncodeToString(e.encodeBytes(userID, meta, expires))
}

// Decode decode a string to token info
func (e *Encoding) Decode(s string) (userID int64, meta interface{}, expires time.Time, err error) {
	var dst []byte
	dst, err = base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return
	}
	return e.decodeBytes(dst)
}

//encodeBytes encode info to byetes
/*
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |Ver| TIME  | MID |PID| CNT | User ID |t|    meta (if any) ...  |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

   t is the type of meta
*/
func (e *Encoding) encodeBytes(userID int64, meta interface{}, expires time.Time) []byte {
	src := make([]byte, 18)
	user := big.NewInt(userID).Bytes()
	if len(user) > 5 {
		panic("user number is too big for big.NewInt(userID).Bytes()")
	}
	objectID := NewObjectIDWithTime(expires)
	for i, v := range objectID {
		src[i+1] = v
	}
	start := 13 + (5 - len(user))
	for i, v := range user {
		src[start+i] = v
	}
	if meta != nil {
		b := interfaceToBytes(meta)
		if len(b) < 2 {
			panic(fmt.Sprintf("meta type %s unspport", reflect.TypeOf(meta)))
		}
		src = append(src, b...)
	}
	if !e.withSecret {
		return src
	}
	dst := make([]byte, len(src))
	e.encoder.XORKeyStream(dst, src)
	return src
}

func (e *Encoding) decodeBytes(dst []byte) (userID int64, meta interface{}, expires time.Time, err error) {
	if dst[0] != 0 {
		err = errors.New("invalidate token version")
	}
	if len(dst) < 18 {
		err = errors.New("invalidate token size")
	}
	objectID := ObjectID(dst[1:13])
	expires = objectID.Time()
	if e.withSecret {
		dst := make([]byte, len(dst))
		e.dconder.XORKeyStream(dst, dst)
	}
	number := dst[13:18]
	var bigNubmer = big.Int{}
	bigNubmer.SetBytes(number)
	userID = bigNubmer.Int64()
	meta = bytesToInterface(dst[18:])
	return
}

const (
	idTypeInt64        byte = 0
	idTypeInt          byte = 1
	idTypeString       byte = 2
	idTypeByte         byte = 3
	idTypeMap          byte = 4
	idTypeBase64String byte = 5
)

func interfaceToBytes(src interface{}) (dst []byte) {
	switch s := src.(type) {
	case int64:
		dst = append(dst, idTypeInt64)
		dst = append(dst, big.NewInt(s).Bytes()...)
	case int:
		dst = append(dst, idTypeInt)
		dst = append(dst, big.NewInt(int64(s)).Bytes()...)
	case string:
		if len(s) > 12 {
			b, err := base64.RawURLEncoding.DecodeString(s)
			if err == nil {
				dst = append(dst, idTypeBase64String)
				dst = append(dst, b...)
				break
			}
		}
		dst = append(dst, idTypeString)
		dst = append(dst, []byte(s)...)
	case []byte:
		dst = append(dst, idTypeByte)
		dst = append(dst, s...)
	case map[string]string:
		var buf bytes.Buffer
		for k, v := range s {
			if strings.ContainsAny(k, "&=") {
				panic(fmt.Sprintf("map key %s can not contains & or = ", k))
			}
			if strings.ContainsAny(k, "&=") {
				panic(fmt.Sprintf("map value %s can not contains & or = ", k))
			}
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(k + "=" + v)
		}
		dst = append(dst, idTypeMap)
		dst = append(dst, buf.Bytes()...)
	default:

	}
	return
}

func bytesToInterface(dst []byte) (src interface{}) {
	if len(dst) < 2 {
		return nil
	}
	t := dst[0]
	switch t {
	case idTypeInt64:
		var bigNumber = big.Int{}
		bigNumber.SetBytes(dst[1:])
		src = bigNumber.Int64()
	case idTypeInt:
		var bigNumber = big.Int{}
		bigNumber.SetBytes(dst[1:])
		src = int(bigNumber.Int64())
	case idTypeString:
		src = string(dst[1:])
	case idTypeByte:
		src = dst[1:]
	case idTypeBase64String:
		src = base64.RawURLEncoding.EncodeToString(dst[1:])
	case idTypeMap:
		m := make(map[string]string)
		query := string(dst[1:])
		for query != "" {
			key := query
			if i := strings.IndexRune(key, '&'); i >= 0 {
				key, query = key[:i], key[i+1:]
			} else {
				query = ""
			}
			if key == "" {
				continue
			}
			value := ""
			if i := strings.IndexRune(key, '='); i >= 0 {
				key, value = key[:i], key[i+1:]
			}
			m[key] = value
		}
		src = m
	}
	return
}

// machineId stores machine id generated once and used in subsequent calls
// to NewObjectID function.
var machineID = readMachineID()
var processID = os.Getpid()
var objectIDCounter = readRandomUint32()

// readRandomUint32 returns a random objectIDCounter.
func readRandomUint32() uint32 {
	var b [4]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot read random object id: %v", err))
	}
	return uint32((uint32(b[0]) << 0) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24))
}

// readMachineId generates and returns a machine id.
// If this function fails to get the hostname it will cause a runtime error.
func readMachineID() []byte {
	var sum [3]byte
	id := sum[:]
	hostname, err1 := os.Hostname()
	if err1 != nil {
		_, err2 := io.ReadFull(rand.Reader, id)
		if err2 != nil {
			panic(fmt.Errorf("cannot get hostname: %v; %v", err1, err2))
		}
		return id
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(id, hw.Sum(nil))
	return id
}

//ObjectID the ObjectID bytes
type ObjectID []byte

// NewObjectID returns a new unique ObjectID.
func NewObjectID() ObjectID {
	b := make([]byte, 12)
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	b[4] = machineID[0]
	b[5] = machineID[1]
	b[6] = machineID[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	b[7] = byte(processID >> 8)
	b[8] = byte(processID)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&objectIDCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return b
}

// NewObjectIDWithTime returns a new unique ObjectID that contains time info
func NewObjectIDWithTime(t time.Time) ObjectID {
	b := make([]byte, 12)
	binary.BigEndian.PutUint32(b[:4], uint32(t.Unix()))
	// Machine, first 3 bytes of md5(hostname)
	b[4] = machineID[0]
	b[5] = machineID[1]
	b[6] = machineID[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	b[7] = byte(processID >> 8)
	b[8] = byte(processID)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&objectIDCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return b
}

// Time get the time of ObjectID
func (id ObjectID) Time() time.Time {
	// First 4 bytes of ObjectID is 32-bit big-endian seconds from epoch.
	secs := int64(binary.BigEndian.Uint32(id.byteSlice(0, 4)))
	return time.Unix(secs, 0)
}

// byteSlice returns byte slice of id from start to end.
// Calling this function with an invalid id will cause a runtime panic.
func (id ObjectID) byteSlice(start, end int) []byte {
	if len(id) != 12 {
		panic(fmt.Sprintf("invalid ObjectID: %q", id))
	}
	return id[start:end]
}

// String ObjectID to base64 string
func (id ObjectID) String() string {
	return base64.RawURLEncoding.EncodeToString(id)
}

// ObjectIDFromString convert a string to objectID
func ObjectIDFromString(src string) (id ObjectID, err error) {
	var dst []byte
	dst, err = base64.RawURLEncoding.DecodeString(src)
	if err != nil {
		return
	}
	if len(dst) != 12 {
		err = errors.New("src size must be 12")
		return
	}
	return dst, nil
}
