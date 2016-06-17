package sip

import (
	"log"
	"net"
	"sip/sipbase"
	"strconv"
	"strings"
)

type State int

const (
	UNREGISTERED State = iota
	REGISTERED
)

type PendingRegisterRequest struct {
	Message  sipbase.Message
	Response chan RegisterResponse
}

type RegisterResponse struct {
	ReponseMessage sipbase.Message
}

type SipClient struct {
	State State

	socket net.Conn
	parser sipbase.Parser

	ProxyHost string
	ProxyPort int
	OwnIP     net.IP
	CSeq      uint32
	Transport string

	Username string
	Password string

	PendingRequests map[uint32]PendingRegisterRequest
}

func CreateClient() SipClient {
	s := SipClient{}
	s.PendingRequests = make(map[uint32]PendingRegisterRequest)
	// DEFAULTS:
	s.CSeq = 100
	s.Transport = "TCP"
	s.OwnIP = sipbase.GetLocalIP()
	return s
}

func (s *SipClient) handleConnection() {

}

func (s *SipClient) ensureSocketInit() {
	if s.socket == nil {
		var err error
		s.socket, err = net.Dial(strings.ToLower(s.Transport), s.ProxyHost+":"+strconv.Itoa(s.ProxyPort))
		log.Println("Socket: ", s.socket)

		s.parser = sipbase.NewParser(s.socket)
		s.parser.SetCallback(func(m sipbase.Message) {
			log.Println("Message arrived:" + m.String())
			cseqNumber, verb := m.GetCSeq()
			if verb == "" {
				log.Println("Error parsing CSEQ: ", err)
			} else {
				log.Println("CSeq is ", cseqNumber)
				crtRequest, ok := s.PendingRequests[cseqNumber]
				if ok {
					log.Println("Found request")
					if m.GetType() == sipbase.RESPONSE {
						if headLine, ok := m.Headline.(sipbase.ResponseHeadline); ok {
							response := RegisterResponse{}
							response.ReponseMessage = m
							crtRequest.Response <- response

							if headLine.Code > 199 {
								// Final response!
								close(crtRequest.Response)
							}
						}
					}
				}
			}
		})
		s.parser.StartParsing()

		if err != nil {
			log.Println("Error:", err)
		}
	}
}

func (s *SipClient) cseq() uint32 {
	val := s.CSeq
	if s.CSeq <= (1<<31)-1 {
		s.CSeq++
	} else {
		s.CSeq = 100
	}
	return val
}

func (s *SipClient) SetProxy(proxyHost string, proxyPort int) {
	s.ProxyHost = proxyHost
	s.ProxyPort = proxyPort
}

func (s *SipClient) SetSignallingTransport(transport string) {
	s.Transport = transport
}

func (s *SipClient) SetAuthenticationBasic(username string, password string) {
	s.Username = username
	s.Password = password
}

func (s *SipClient) SetOwnIP(ownIP net.IP) {
	s.OwnIP = ownIP
}
func (s *SipClient) Register() PendingRegisterRequest {
	s.ensureSocketInit()
	c := sipbase.CreateRequest("REGISTER", "sip:"+s.ProxyHost)
	viaBranch := sipbase.RandSeq(10)
	fromTag := sipbase.RandSeq(10)
	callID := sipbase.RandSeq(10)
	crtCSeq := s.cseq()

	c.SetVia(s.Transport, s.OwnIP.String(), 5060, viaBranch)
	c.SetFrom("sip", s.Username, s.ProxyHost, fromTag)
	c.SetTo("sip", s.Username, s.ProxyHost, "")
	c.SetCallId(callID)
	c.SetCSeq(crtCSeq, "REGISTER")
	c.SetContact("sip", s.Username, s.OwnIP.String(), 5060)
	c.SetExpires(300)
	c.SetUserAgent("sipbell/0.1")
	c.SetAllow([]string{"PRACK", "INVITE", "ACK", "BYE", "CANCEL", "UPDATE", "INFO", "SUBSCRIBE", "NOTIFY", "OPTIONS", "REFER", "MESSAGE"})
	c.SetContentLength(0)

	req := PendingRegisterRequest{}
	req.Response = make(chan RegisterResponse)
	req.Message = c

	s.PendingRequests[crtCSeq] = req
	s.socket.Write([]byte(c.String()))

	return req
}

func (s *SipClient) TryRegister() {
	pendingRequest := s.Register()
	for {
		log.Println("Wait for next response")
		response, more := <-pendingRequest.Response
		log.Println("Received response")
		message := response.ReponseMessage
		if !more {
			break
		}
		log.Println("message", message)
	}
	log.Println("Responses done")
}
