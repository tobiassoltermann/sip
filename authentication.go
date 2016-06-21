package sip

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
)

type WWWAuthenticate interface {
	GetMechanism() string
}
type WWWAuthenticateImpl struct {
	Mechanism string
}

func (w WWWAuthenticateImpl) GetMechanism() string {
	return "UNKNOWN"
}

type DigestWWWAuthenticate struct {
	WWWAuthenticateImpl
	Realm     string
	Nonce     string
	Algorithm string
}

func (w DigestWWWAuthenticate) GetMechanism() string {
	return "DIGEST"
}

func ParseWWWAuthenticate(line HeaderLine) (WWWAuthenticate, error) {
	var auth WWWAuthenticate
	if strings.ToLower(line.Name) != "www-authenticate" {
		return WWWAuthenticateImpl{}, errors.New("Not a WWW-Authenticate line")
	}
	authenticateLine := strings.SplitN(line.Value, " ", 2)

	switch strings.ToUpper(authenticateLine[0]) {
	case "DIGEST":
		authDigest := DigestWWWAuthenticate{}
		digestParams := strings.Split(authenticateLine[1], ", ")
		for _, item := range digestParams {
			itemCombo := strings.Split(item, "=")
			itemName := itemCombo[0]
			itemValue := itemCombo[1]

			switch strings.ToLower(itemName) {
			case "realm":
				authDigest.Realm = itemValue
			case "nonce":
				authDigest.Nonce = itemValue
			case "algorithm":
				authDigest.Algorithm = itemValue
			}
		}
		auth = authDigest
	}

	return auth, nil
}

type AuthInformation struct {
	Wwwauth  DigestWWWAuthenticate
	Username string
	Password string
	URL      string
}

func (a *AuthInformation) FinalHash() string {
	ha1 := md5.Sum([]byte(a.Username + ":" + a.Wwwauth.Realm + ":" + a.Password))
	ha1s := fmt.Sprintf("%x", ha1)

	ha2 := md5.Sum([]byte("REGISTER:" + "sip:" + a.URL))
	ha2s := fmt.Sprintf("%x", ha2)

	finalHash := md5.Sum([]byte(ha1s + ":" + a.Wwwauth.Nonce + ":" + ha2s))
	finalHashs := fmt.Sprintf("%x", finalHash)
	return finalHashs

}
