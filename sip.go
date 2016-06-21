package sip

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
)

type RegistrationResult int

const (
	OKAY RegistrationResult = iota
	UNAUTHORIZED
	ERROR
)

type SipClient struct {
	socket           net.Conn
	Listeners        map[string]*Listener
	DefaultTransport string
	registerInfo     *RegisterInfo
	done             chan int

	callCallback   CallCallback
	cancelCallback CallCallback
}
type Call struct {
	From string
}
type CallCallback func(*Call)

func CreateClient() SipClient {
	s := SipClient{}

	// DEFAULTS:
	s.Listeners = make(map[string]*Listener)
	return s
}

func (s *SipClient) Listen(transport string, host string, port int) error {
	id := fmt.Sprintf("%s_%s_%d", transport, host, port)
	_, alreadyThere := s.Listeners[id]
	if alreadyThere {
		return errors.New("Already accepting connections on this Listener")
	}

	//	var err error
	var dialogListener func(d *Dialog) = func(d *Dialog) {
		d.OnMessage(func(m *Message) {
			switch m.GetType() {
			case REQUEST:
				requestHeadline, ok := m.Headline.(RequestHeadline)
				if ok {
					switch requestHeadline.Method {
					case "INVITE":
						d.Reply100Trying()
						d.Reply180Ringing()

						c := Call{}
						c.From = m.GetFrom()

						if s.callCallback != nil {
							s.callCallback(&c)
						}
					case "CANCEL":
						d.Reply200Ok()
						c := Call{}
						c.From = m.GetFrom()
						if s.cancelCallback != nil {
							s.cancelCallback(&c)
						}
					default:

						log.Println("Message is:", requestHeadline.Method)
					}
				}

			}
		})
	}

	l := CreateListener(transport, host, port, s, dialogListener)

	s.Listeners[id] = l

	return nil
}

func (s *SipClient) OnIncomingCall(callback CallCallback) {
	s.callCallback = callback
}
func (s *SipClient) OnCancel(callback CallCallback) {
	s.cancelCallback = callback
}

func (s *SipClient) StopListeningAll() {
	numRemaining := len(s.Listeners)
	done := make(chan bool)
	for i, crtListener := range s.Listeners {
		go func() {
			crtListener.Stop()
			numRemaining--
			if numRemaining == 0 {
				done <- true
			}
		}()
		crtListener.Stop()
		delete(s.Listeners, i)
	}
	_ = <-done
}

func (s *SipClient) SetDefaultTransport(transport string) {
	s.DefaultTransport = transport
}

func (s *SipClient) TryRegister(registerInfo *RegisterInfo) (RegistrationResult, error) {
	connectInfo := registerInfo.Registrar
	if connectInfo.Transport == "" {
		connectInfo.Transport = "tcp"
	}

	socket, err := net.Dial(connectInfo.Transport, connectInfo.Host+":"+strconv.Itoa(connectInfo.Port))
	if err != nil {
		return ERROR, err
	}
	dialog := CreateDialog(socket, s)
	log.Printf("[P] Created: %p\n", dialog)
	unauthRegResult := make(chan RegistrationResult)
	var auth WWWAuthenticate
	var innerErr error
	dialog.OnMessage(func(m *Message) {
		if m.GetType() == RESPONSE {
			responseHeader, ok := m.Headline.(ResponseHeadline)
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
					auth, err = ParseWWWAuthenticate(authLine)
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

	dialog.OnMessage(func(m *Message) {
		if m.GetType() == RESPONSE {
			responseHeader, ok := m.Headline.(ResponseHeadline)
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

	digestAuth, ok1 := auth.(DigestWWWAuthenticate)
	userInfo := registerInfo.UserInfo
	digestAuthInfo, ok2 := userInfo.(*DigestUserInfoImpl)
	if ok1 && ok2 {
		authInfo := AuthInformation{
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

func (s *SipClient) WaitAll() {
	for {
		value := <-s.done
		if value < 0 {
			return
		}
	}
}
