package sipstack

import (
	"errors"
	"log"
	"net"
	"sip/sipbase"
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
	conns    []net.Conn
	listener net.Listener
	running  bool
}

func (l *Listener) Stop() {
	l.running = true
}

func CreateListener(transport string, host string, port int) Listener {
	var err error
	var l Listener = Listener{}
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
			conn, _ := l.listener.Accept()
			l.conns = append(l.conns, conn)
			parser := sipbase.NewParser(conn)
			parser.SetCallback(func(m sipbase.Message) {
				log.Println("Message arrived: " + m.String())
			})
			parser.StartParsing()
		}
	}()

	return l
}
