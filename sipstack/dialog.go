package sipstack

import (
	"log"
	"sip/sipbase"
)

type Dialog struct {
	Conn     Outbound
	CSeq     uint32
	Messages []sipbase.Message
	Parser   *sipbase.Parser
}

func (d *Dialog) cseq() uint32 {
	val := d.CSeq
	if d.CSeq <= (1<<31)-1 {
		d.CSeq++
	} else {
		d.CSeq = 100
	}
	return val
}

func CreateDialog(transport string, host string, port int) (Dialog, error) {
	d := Dialog{}
	d.CSeq = 100
	if transport == "" {
		transport = "tcp"
	}
	conn, err := Dial(transport, host, port)
	if err != nil {
		return Dialog{}, err
	}
	d.Conn = conn

	d.Parser = sipbase.NewParser(d.Conn.Socket)
	d.Parser.StartParsing()

	return d, nil
}

func (d *Dialog) OnMessage(callback sipbase.Callback) {
	log.Println("dialog.OnMessage: ", callback)
	d.Parser.SetCallback(callback)
}

func (d *Dialog) sendMessage(m *sipbase.Message) {
	d.Conn.Socket.Write([]byte(m.String()))
}

func (d *Dialog) SendRegister(registerInfo RegisterInfo, authInfo *sipbase.AuthInformation) {
	clientCI := registerInfo.Client
	registrarCI := registerInfo.Registrar
	userName := registerInfo.Username

	c := sipbase.CreateRequest("REGISTER", "sip:"+d.Conn.Conn.Host)
	viaBranch := sipbase.RandSeq(10)
	fromTag := sipbase.RandSeq(10)
	callID := sipbase.RandSeq(10)
	crtCSeq := d.cseq()

	c.SetVia(clientCI.Transport, clientCI.Host, clientCI.Port, viaBranch)
	c.SetFrom("sip", userName, registrarCI.Host, fromTag)
	c.SetTo("sip", userName, registrarCI.Host, "")
	c.SetCallId(callID)
	c.SetCSeq(crtCSeq, "REGISTER")
	c.SetContact("sip", userName, clientCI.Host, clientCI.Port)
	c.SetExpires(300)
	if authInfo != nil {
		c.SetDigestAuthorizationHeader(*authInfo)
	}
	c.SetUserAgent("sipbell/0.1")
	c.SetAllow([]string{"PRACK", "INVITE", "ACK", "BYE", "CANCEL", "UPDATE", "INFO", "SUBSCRIBE", "NOTIFY", "OPTIONS", "REFER", "MESSAGE"})
	c.SetContentLength(0)

	d.sendMessage(&c)
}
