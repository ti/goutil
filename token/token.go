// Package token implements a base64 token from ObjectID
package token

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"sync/atomic"
	"time"
)

// NewEncoding returns a new token Encoding defined by the given secret,
// secret should be less 256 bytes
func NewEncoding(secret string) *Encoding {
	if len(secret) == 0 {
		return &Encoding{}
	}
	sec := []byte(secret)
	decoder, err := rc4.NewCipher(sec)
	if err != nil {
		panic(err)
	}
	encoder, err := rc4.NewCipher(sec)
	if err != nil {
		panic(err)
	}
	return &Encoding{
		withSecret: true,
		dconder:    decoder,
		encoder:    encoder,
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
	return base64.RawURLEncoding.EncodeToString(e.EncodeBytes(userID, meta, expires))
}

// UnionID encode a interface to bytes
func (e *Encoding) UnionID(src []uint64) string {
	dst := intArrayToBytes(src)
	if !e.withSecret {
		dst = append([]byte{0}, dst...)
		return base64.RawURLEncoding.EncodeToString(dst)
	}
	enc := make([]byte, len(dst))
	e.encoder.XORKeyStream(enc, dst)
	enc = append([]byte{0}, enc...)
	return base64.RawURLEncoding.EncodeToString(enc)
}

//FromUnionID convert interface to union id
func (e *Encoding) FromUnionID(src string) ([]uint64, error) {
	dst, err := base64.RawURLEncoding.DecodeString(src)
	if err != nil {
		return nil, err
	}
	dst = dst[1:]
	if e.withSecret {
		src := make([]byte, len(dst))
		e.dconder.XORKeyStream(src, dst)
		dst = src
	}
	return bytesToIntArray(dst), nil
}

// Decode decode a string to token info
func (e *Encoding) Decode(s string) (userID int64, meta interface{}, expires time.Time, err error) {
	var dst []byte
	dst, err = base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return
	}
	return e.DecodeBytes(dst)
}

//encodeBytes encode info to byetes
/*
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |Ver| TIME  | MID |PID| CNT | User ID |t|    meta (if any) ...  |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

   t is the type of meta
*/
func (e *Encoding) EncodeBytes(userID int64, meta interface{}, expires time.Time) []byte {
	src := make([]byte, 17)
	user := big.NewInt(userID).Bytes()
	if len(user) > 5 {
		panic("user number is too big for big.NewInt(userID).Bytes()")
	}
	objectID := NewObjectIDWithTime(expires)
	for i, v := range objectID {
		src[i] = v
	}
	start := 12 + (5 - len(user))
	for i, v := range user {
		src[start+i] = v
	}
	if meta != nil {
		b, err := interfaceToBytes(meta)
		if err != nil {
			panic(err)
		}
		src = append(src, b...)
	}
	if !e.withSecret {
		//add version
		src = append([]byte{0}, src...)
		return src
	}
	dst := make([]byte, len(src))
	e.encoder.XORKeyStream(dst, src)
	dst = append([]byte{0}, dst...)
	return dst
}

func (e *Encoding) DecodeBytes(dst []byte) (userID int64, meta interface{}, expires time.Time, err error) {
	if dst[0] != 0 {
		err = errors.New("invalidate token version")
	}
	if len(dst) < 18 {
		err = errors.New("invalidate token size")
	}
	dst = dst[1:]
	if e.withSecret {
		src := make([]byte, len(dst))
		e.dconder.XORKeyStream(src, dst)
		dst = src
	}
	objectID := ObjectID(dst[0:12])
	expires = objectID.Time()
	number := dst[12:17]
	var bigNumber = big.Int{}
	bigNumber.SetBytes(number)
	userID = bigNumber.Int64()
	meta = bytesToInterface(dst[17:])
	return
}

const (
	idTypeInt64        byte = 0
	idTypeInt          byte = 1
	idTypeString       byte = 2
	idTypeByte         byte = 3
	idTypeMap          byte = 4
	idTypeIntMap       byte = 5
	idTypeIntArray     byte = 6
	idTypeBase64String byte = 7
)

func interfaceToBytes(src interface{}) (dst []byte, err error) {
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
	case map[uint64]uint64:
		dst = append(dst, idTypeIntMap)
		dst = append(dst, intMapToBytes(s)...)
	case []uint64:
		dst = append(dst, idTypeIntArray)
		dst = append(dst, intArrayToBytes(s)...)
	case map[string]string:
		dst = append(dst, idTypeMap)
		dst = append(dst, strMapToBytes(s)...)
	default:
		return nil, fmt.Errorf("meta type %s unspport", s)
	}
	return
}

func bytesToInterface(dst []byte) (src interface{}) {
	if len(dst) < 1 {
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
	case idTypeIntMap:
		src = bytesToIntMap(dst[1:])
	case idTypeIntArray:
		src = bytesToIntArray(dst[1:])
	case idTypeMap:
		src = bytesToStrMap(dst[1:])
	}
	return
}

func intArrayToBytes(src []uint64) (dst []byte) {
	for _, v := range src {
		if len(dst) > 1 {
			dst = append(dst, 255)
		}
		dst = append(dst, intToBytes(v, 254)...)
	}
	return
}

func bytesToIntArray(b []byte) []uint64 {
	var numbers []uint64
	for len(b) > 0 {
		m := bytes.IndexByte(b, 255)
		if m < 0 {
			numbers = append(numbers, bytesToInt(b, 254))
			break
		}
		n := b[0:m]
		numbers = append(numbers, bytesToInt(n, 254))
		b = b[m+1:]
	}
	return numbers
}
func intMapToBytes(m map[uint64]uint64) []byte {
	if len(m) == 0 {
		return nil
	}
	var ret []byte
	var sep byte = 255
	for k, v := range m {
		if len(ret) != 0 {
			ret = append(ret, sep)
		}
		if k == 0 {
			ret = append(ret, 0)
		} else {
			ret = append(ret, intToBytes(k, 254)...)
		}
		ret = append(ret, sep)
		if v == 0 {
			ret = append(ret, 0)
		} else {
			ret = append(ret, intToBytes(v, 254)...)
		}
	}
	return ret
}

func bytesToIntMap(src []byte) map[uint64]uint64 {
	m := make(map[uint64]uint64)
	query := src
	var sep byte = 255
	for len(query) != 0 {
		var k, v []byte
		if i := bytes.IndexByte(query, sep); i >= 0 {
			if i == 0 {
				query = query[i+2:]
			} else {
				k, query = query[:i], query[i+1:]
			}
		} else {
			query = []byte{}
		}
		if len(query) == 0 {
			continue
		}
		if i := bytes.IndexByte(query, sep); i >= 0 {
			if i == 0 {
				if len(query) < 2 {
					query = []byte{}
				} else {
					query = query[i+2:]
				}
			} else {
				v, query = query[:i], query[i+1:]
			}
		} else {
			v = query
			query = []byte{}
		}
		var keyInt, valueInt uint64
		if len(k) > 0 {
			keyInt = bytesToInt(k, 254)
		}
		if len(v) > 0 {
			valueInt = bytesToInt(v, 254)
		}
		m[keyInt] = valueInt
	}
	return m
}

// decimal to any bytes less than 255
func intToBytes(num uint64, base uint8) []byte {
	var ret []byte
	var remainder uint64
	var remainderByte byte
	var baseNum = uint64(base) + 1
	for num != 0 {
		remainder = num % baseNum
		remainderByte = byte(remainder)
		ret = append([]byte{remainderByte}, ret...)
		num = num / baseNum
	}
	return ret
}

// []byte to nay  decimal
func bytesToInt(num []byte, base uint8) uint64 {
	if len(num) == 0 {
		return 0
	}
	nNum := float64(len(num) - 1)
	baseNum := float64(base) + 1
	var ret float64
	for _, value := range num {
		ret = ret + float64(value)*math.Pow(baseNum, nNum)
		nNum = nNum - 1
	}
	return uint64(ret)
}

func strMapToBytes(m map[string]string) []byte {
	if len(m) == 0 {
		return nil
	}
	var ret []byte
	for k, v := range m {
		if len(ret) != 0 {
			ret = append(ret, 0)
		}
		ret = append(ret, []byte(k)...)
		ret = append(ret, 0)
		ret = append(ret, []byte(v)...)
	}
	return ret
}

func bytesToStrMap(src []byte) map[string]string {
	m := make(map[string]string)
	query := src
	var sep byte
	for len(query) != 0 {
		var k, v []byte
		if i := bytes.IndexByte(query, sep); i >= 0 {
			k, query = query[:i], query[i+1:]
		} else {
			query = []byte{}
		}
		if len(query) == 0 {
			continue
		}
		if i := bytes.IndexByte(query, sep); i >= 0 {
			v, query = query[:i], query[i+1:]
		} else {
			v = query
			query = []byte{}
		}
		m[string(k)] = string(v)
	}
	return m
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

// Bytes ObjectID to bytes
func (id ObjectID) Bytes() []byte {
	return []byte(id)
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
