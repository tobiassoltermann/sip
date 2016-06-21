package sip

import (
	"errors"
	"log"
	"net"
	"strconv"
)

type Connectinfo struct {
	Transport string
	Host      string
	Port      int
}

type Outbound struct {
	Conn   Connectinfo
	Socket net.Conn
}

func Dial(transport string, host string, port int) (Outbound, error) {
	o := Outbound{}

	o.Conn = Connectinfo{
		transport,
		host,
		port,
	}
	socket, err := net.Dial(transport, host+":"+strconv.Itoa(port))
	if err != nil {
		return Outbound{}, errors.New("Could not connect:" + err.Error())
	}
	o.Socket = socket
	return o, nil
}

type Listener struct {
	Connectinfo
	conns           []net.Conn
	listener        net.Listener
	running         bool
	stoppingChannel chan bool
}

func (l *Listener) Stop() {
	l.running = false
	l.listener.Close()
	_ = <-l.stoppingChannel
}

func CreateListener(transport string, host string, port int, sipClient *SipClient, dialogListener func(d *Dialog)) *Listener {
	var err error
	var l Listener = Listener{}
	l.stoppingChannel = make(chan bool, 1)
	l.Host = host
	l.Port = port
	l.Transport = transport
	l.listener, err = net.Listen(transport, host+":"+strconv.Itoa(port))
	if err != nil {
		log.Println("Error listening for ", transport, host, port, " due to ", err)
	}
	l.running = true

	go func() {
		for l.running {
			conn, errListen := l.listener.Accept()
			if errListen != nil {
				break
			}
			d := CreateDialog(conn, sipClient)
			dialogListener(d)
		}
		l.stoppingChannel <- true
	}()

	return &l
}
