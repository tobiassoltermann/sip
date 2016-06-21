package sip

import (
	"bufio"
	"math/rand"
	"time"
)

const sipversion = "2.0"

type SipUri struct {
	uri string
}

func ParseSipUri(uri string) SipUri {
	return SipUri{uri}
}

func (s *SipUri) String() string {
	return s.uri
}

// Readln returns a single line (without the ending \n)
// from the input buffered reader.
// An error is returned iff there is an error with the
// buffered reader.
func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var randomized bool = false

func RandSeq(n int) string {
	if !randomized {
		rand.Seed(time.Now().UnixNano())
		randomized = true
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
