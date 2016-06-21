package sip

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// ---------------
type HeaderLine struct {
	Name  string
	Value string
}

// ---------------

type Headers struct {
	Lines []HeaderLine
}

func (h *Headers) AddHeader(name string, value string) {
	h.Lines = append(h.Lines,
		HeaderLine{
			name,
			value,
		})
}

func (h *Headers) ReplaceAddHeader(name string, value string) {
	found := false
	for _, crt := range h.Lines {
		if crt.Name == name {
			crt.Value = value
			found = true
		}
	}
	if !found {
		h.AddHeader(name, value)
	}
}

func (h *Headers) FindHeaderByName(name string) (header HeaderLine, err error) {
	for _, crt := range h.Lines {
		if crt.Name == name {
			return crt, nil
		}
	}
	return HeaderLine{}, errors.New("Not found")
}

// ---------------

type Headline interface {
	ToString() string
}

// ---------------

type MessageType int

const (
	REQUEST MessageType = iota
	RESPONSE
)

type Message struct {
	Headline    Headline
	Headers     Headers
	Body        []byte
	RequestID   string
	MessageType MessageType
}

func (m *Message) SetRequestId(requestID string) {
	m.RequestID = requestID
}

func (m *Message) GetFrom() string {
	header, err := m.Headers.FindHeaderByName("From")
	if err != nil {
		log.Println("Error finding header From", err)
	}
	return header.Value
}

func (m *Message) SetFrom(proto string, user string, host string, tag string) *Message {
	var value string
	if tag == "" {
		value = fmt.Sprintf("<%s:%s@%s>", proto, user, host)
	} else {
		value = fmt.Sprintf("<%s:%s@%s>;tag=%s", proto, user, host, tag)
	}
	m.Headers.ReplaceAddHeader("From", value)
	return m
}
func (m *Message) SetFromValue(value string) *Message {
	m.Headers.ReplaceAddHeader("From", value)
	return m
}

func (m *Message) SetContact(proto string, user string, host string, port int) *Message {
	var value string
	value = fmt.Sprintf("<%s:%s@%s:%d>;transport=tcp", proto, user, host, port)
	m.Headers.ReplaceAddHeader("Contact", value)
	return m
}
func (m *Message) SetUserAgent(value string) *Message {
	m.Headers.ReplaceAddHeader("User-Agent", value)
	return m
}
func (m *Message) GetTo() string {
	header, err := m.Headers.FindHeaderByName("To")
	if err != nil {
		log.Println("Error finding header To", err)
	}
	return header.Value
}
func (m *Message) SetTo(proto string, user string, host string, tag string) *Message {
	var value string
	if tag == "" {
		value = fmt.Sprintf("<%s:%s@%s>", proto, user, host)
	} else {
		value = fmt.Sprintf("<%s:%s@%s>;tag=%s", proto, user, host, tag)
	}
	m.Headers.ReplaceAddHeader("To", value)
	return m
}
func (m *Message) SetToValue(value string) *Message {
	m.Headers.ReplaceAddHeader("To", value)
	return m
}

func (m *Message) GetVia() string {
	header, err := m.Headers.FindHeaderByName("Via")
	if err != nil {
		log.Println("Error finding header Via", err)
	}
	return header.Value
}

func (m *Message) SetVia(transport string, host string, port int, branch string) *Message {
	value := fmt.Sprintf("SIP/2.0/%s %s:%d;rport;branch=z9hG4bK%s", transport, host, port, branch)
	m.Headers.ReplaceAddHeader("Via", value)
	return m
}
func (m *Message) SetViaValue(value string) *Message {
	m.Headers.ReplaceAddHeader("Via", value)
	return m
}

func (m *Message) GetCallId() string {
	header, err := m.Headers.FindHeaderByName("Call-ID")
	if err != nil {
		log.Println("Error finding header Call-ID", err)
	}
	return header.Value
}

func (m *Message) SetCallId(value string) *Message {
	m.Headers.ReplaceAddHeader("Call-ID", value)
	return m
}
func (m *Message) SetContentLength(length int) *Message {
	m.Headers.AddHeader("Content-Length", strconv.Itoa(length))
	return m
}

func (m *Message) GetCSeq() (number uint32, verb string) {
	header, err := m.Headers.FindHeaderByName("CSeq")
	if err != nil {
		log.Println("Error finding header CSeq", err)
	}
	headerLine := strings.Split(header.Value, " ")
	numberString := headerLine[0]
	number64, parseErr := strconv.ParseUint(numberString, 10, 0)
	number = uint32(number64)
	verb = headerLine[1]
	if parseErr != nil {
		log.Println("Cannot parse numberString to int" + numberString)
	}

	return number, verb
}
func (m *Message) SetCSeq(number uint32, verb string) *Message {
	m.Headers.AddHeader("CSeq", strconv.FormatUint(uint64(number), 10)+" "+verb)
	return m
}

func (m *Message) GetExpires() int {
	header, err := m.Headers.FindHeaderByName("Expires")
	if err != nil {
		log.Println("Field 'Expires' could not be found. Assume 120")
		return 120
	}
	expirationValue, err := strconv.Atoi(header.Value)
	if err != nil {
		log.Println("Field 'Expires' could not be parsed. Assume 120")
		return 120
	}
	return expirationValue
}

func (m *Message) SetExpires(value int) *Message {
	m.Headers.AddHeader("Expires", strconv.Itoa(300))
	return m
}

func (m *Message) SetDigestAuthorizationHeader(authInfo AuthInformation) *Message {
	value := fmt.Sprintf(`Digest username="%s" realm="%s" nonce="%s" response="%s"`, authInfo.Username, authInfo.Wwwauth.Realm, authInfo.Wwwauth.Nonce, authInfo.FinalHash())
	m.Headers.AddHeader("Authorization", value)
	log.Println(value)
	return m
}

func (m *Message) SetAllow(actions []string) *Message {
	m.Headers.AddHeader("Allow", strings.Join(actions, ", "))
	return m
}

func (m *Message) AddHeader(name string, value string) *Message {
	m.Headers.AddHeader(name, value)
	return m
}

func (m *Message) GetType() MessageType {
	return m.MessageType
}

func (m *Message) String() string {
	s := ""
	s += m.Headline.ToString() + "\n"
	for _, crtHeader := range m.Headers.Lines {
		crtLine := crtHeader.Name + ": " + crtHeader.Value + "\n"
		s += crtLine
	}
	s += "\n"
	s += string(m.Body)
	return s
}
