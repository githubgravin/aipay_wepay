package wepay

import (
	"crypto/rand"
	"fmt"
	"strings"
)

func parseTag(tag string) string {
	ts := strings.SplitN(tag, ",", 2)
	if len(ts) >= 1 {
		return ts[0]
	} else {
		return ""
	}
}

func getRandom() (string, error) {
	c := 16
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	str := fmt.Sprintf("%02x", b)
	return str, nil
}
