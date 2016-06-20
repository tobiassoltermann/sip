package sip

import (
	"errors"
	"fmt"
	"log"
	"net"
	sipbase "sip/sipbase"
	sipstack "sip/sipstack"
	_ "strconv"
)

type RegistrationResult int

const (
	OKAY RegistrationResult = iota
	UNAUTHORIZED
	ERROR
)

type SipClient struct {
	socket    net.Conn
	Listeners map[string]sipstack.Listener

	OwnIP net.IP

	DefaultTransport string

	done chan int
}

func CreateClient() SipClient {
	s := SipClient{}

	// DEFAULTS:
	s.OwnIP = sipbase.GetLocalIP()
	s.Listeners = make(map[string]sipstack.Listener)
	return s
}

func (s *SipClient) Listen(transport string, host string, port int) error {
	id := fmt.Sprintf("%s_%s_%d", transport, host, port)
	_, alreadyThere := s.Listeners[id]
	if alreadyThere {
		return errors.New("Already accepting connections on this Listener")
	}

	var err error
	l := sipstack.CreateListener(transport, host, port)

	s.Listeners[id] = l
	if err != nil {
		log.Println("Error:", err)
	}

	return nil
}

/*
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

						response := RegisterResponse{}
						response.ResponseMessage = m

						crtRequest.Response <- response

						if response.IsFinal() {
							// Final response!
							close(crtRequest.Response)
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
*/

func (s *SipClient) SetDefaultTransport(transport string) {
	s.DefaultTransport = transport
}

func (s *SipClient) SetOwnIP(ownIP net.IP) {
	s.OwnIP = ownIP

}

func (s *SipClient) TryRegister(registerInfo sipstack.RegisterInfo) (RegistrationResult, error) {
	connectInfo := registerInfo.Registrar
	if connectInfo.Transport == "" {
		connectInfo.Transport = "tcp"
	}
	dialog, err := sipstack.CreateDialog(connectInfo.Transport, connectInfo.Host, connectInfo.Port)
	if err != nil {
		return ERROR, err
	}

	unauthRegResult := make(chan RegistrationResult)
	var auth sipbase.WWWAuthenticate
	var innerErr error
	dialog.OnMessage(func(m sipbase.Message) {
		if m.GetType() == sipbase.RESPONSE {
			responseHeader, ok := m.Headline.(sipbase.ResponseHeadline)
			if !ok {
				log.Println("Error. Type is RESPONSE, but headline is REQUEST")
				log.Println("Details: ", m)
				unauthRegResult <- ERROR
				innerErr = errors.New("Error. Type is RESPONSE, but headline is REQUEST")
				return
			}
			if responseHeader.IsFinal() {
				finalCode := responseHeader.Code
				if err != nil {
					log.Println("Couldn't determine the response code")
					unauthRegResult <- ERROR
					return
				}
				switch finalCode {
				case 200:
					unauthRegResult <- OKAY
					return
				case 401:
					authLine, err := m.Headers.FindHeaderByName("WWW-Authenticate")
					auth, err = sipbase.ParseWWWAuthenticate(authLine)
					if err != nil {
						log.Println("Error parsing wwwauthenticate: ", err)
					}
					unauthRegResult <- UNAUTHORIZED
					return
				}

			}
		}

	})
	dialog.SendRegister(registerInfo, nil)

	res := <-unauthRegResult
	if res == OKAY {
		return OKAY, nil
	}
	if res == ERROR {
		return ERROR, innerErr
	}

	if registerInfo.UserInfo.GetType() == "UNAUTHORIZED" {
		return UNAUTHORIZED, errors.New("Authorization required but not provided")
	}
	// Retry registration with authorization information
	authRegResult := make(chan RegistrationResult)

	dialog.OnMessage(func(m sipbase.Message) {
		if m.GetType() == sipbase.RESPONSE {
			responseHeader, ok := m.Headline.(sipbase.ResponseHeadline)
			if !ok {
				log.Println("Error. Type is RESPONSE, but headline is REQUEST")
				log.Println("Details: ", m)
				authRegResult <- ERROR
				return
			}
			if responseHeader.IsFinal() {
				finalCode := responseHeader.Code
				if err != nil {
					log.Println("Couldn't determine the response code")
					authRegResult <- ERROR
					return
				}
				switch finalCode {
				case 200:
					authRegResult <- OKAY
					return
				case 401:
					authRegResult <- UNAUTHORIZED
					return
				}

			}
		}

	})

	digestAuth, ok1 := auth.(sipbase.DigestWWWAuthenticate)
	userInfo := registerInfo.UserInfo
	digestAuthInfo, ok2 := userInfo.(*sipstack.DigestUserInfoImpl)
	if ok1 && ok2 {
		authInfo := sipbase.AuthInformation{
			digestAuth,
			userInfo.GetUsername(),
			digestAuthInfo.GetPassword(),
			"sip:" + connectInfo.Host,
		}
		dialog.SendRegister(registerInfo, &authInfo)
		//dialog.SendRegister(&authInfo, connectInfo)
	}

	res2 := <-authRegResult
	return res2, innerErr
}

/*
func (s *SipClient) TryRegister() (bool, string) {
	unauthorized := false
	var finalCode int
	var reply string
	var auth sipbase.WWWAuthenticate
	pendingRequest := s.Register(nil)
	for {
		log.Println("Wait for next response")
		response, more := <-pendingRequest.Response
		if more {
			log.Println("Received response. More = ", more)

			message := response.ResponseMessage
			if response.IsFinal() {
				delete(s.PendingRequests, s.CSeq)
				finalCode, err := response.Code()
				if err != nil {
					log.Println("Couldn't determine the response code")
					continue
				}
				reply, err := response.Reply()
				if err != nil {
					log.Println("Couldn't get the reply text")
				}
				registrationOK := finalCode == 200
				if registrationOK {
					s.State = REGISTERED
				} else {

					s.State = UNREGISTERED
				}

				if finalCode == 401 {
					unauthorized = true
					log.Println("Unauthorized! Let's find the WWW-Authenticate")
					authLine, err := response.ResponseMessage.Headers.FindHeaderByName("WWW-Authenticate")
					auth, err = sipbase.ParseWWWAuthenticate(authLine)
					if err != nil {
						log.Println("Error parsing wwwauthenticate: ", err)
					}
					break
				}
				log.Println("reply", reply)
			}
			log.Println("message", message)
		} else {
			break
		}

	}
	log.Println("Responses done")
	log.Println("State registered?")
	if s.State == REGISTERED {
		return true, ""
	}
	if !unauthorized {
		return false, "Couldn't register: " + strconv.Itoa(finalCode) + ", " + reply
	}

	log.Println("Unauthorized. Retry with auth!")
	log.Println("=================================")
	log.Println("Auth: ", auth.GetMechanism())
	authInfo := sipbase.AuthInformation{}

	switch auth.GetMechanism() {
	case "DIGEST":
		if digestAuth, ok := auth.(sipbase.DigestWWWAuthenticate); ok {
			log.Println(digestAuth.Algorithm)

			authInfo.Username = s.Username
			authInfo.Password = s.Password
			authInfo.Wwwauth = digestAuth
			authInfo.URL = s.ProxyHost

		}
	}
	pendingAuthRequest := s.Register(&authInfo)
	for {
		log.Println("Wait for next response")
		response, more := <-pendingAuthRequest.Response
		if more {
			log.Println("Received response. More = ", more)

			message := response.ResponseMessage
			log.Println("message", message)
			if response.IsFinal() {
				delete(s.PendingRequests, s.CSeq)
				finalCode, err := response.Code()
				if err != nil {
					log.Println("Couldn't determine the response code")
					continue
				}

				registrationOK := finalCode == 200
				if registrationOK {
					log.Println("Registered.")

					s.State = REGISTERED
					expiresIn := response.ResponseMessage.GetExpires()
					log.Println("Registration is valid only for ", expiresIn, "seconds.")
				} else {
					s.State = UNREGISTERED
				}

			}

		} else {
			break
		}

	}
	if s.State == REGISTERED {
		return true, ""
	}
	return false, "Were there any?"
}
*/
func (s *SipClient) WaitAll() {
	for {
		value := <-s.done
		if value < 0 {
			return
		}
	}
}
