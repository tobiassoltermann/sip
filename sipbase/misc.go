package sipbase

import (
	"bufio"
	"math/rand"
	"net"
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

func GetLocalIP() net.IP {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return net.IPv4zero
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP
			}
		}
	}
	return net.IPv4zero
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
