package token

import (
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	secretToken := NewEncoding("secret")
	var userID int64 = 999999999999
	exp, _ := time.Parse(time.RFC3339, "2008-01-02T15:04:05+07:00")
	var clientID int64 = 7652
	secretStr := secretToken.Encode(userID, clientID, exp)
	sUserID, sClientID, sTime, err := secretToken.Decode(secretStr)
	if err != nil {
		t.Fatal(err)
		return
	}
	if sUserID != userID || sClientID != clientID || !sTime.Equal(exp) {
		t.Fatal("decode does not match", sTime)
		return
	}
	commonToken := NewEncoding("")
	commonStr := commonToken.Encode(userID, clientID, exp)
	cUserID, cClientID, cTime, err := commonToken.Decode(commonStr)
	if err != nil {
		t.Fatal(err)
		return
	}
	if cUserID != userID || cClientID != clientID || !cTime.Equal(exp) {
		t.Fatal("decode does not match")
		return
	}
}
