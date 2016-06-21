package sip

import (
	"net"
)

type Dialog struct {
	Conn     net.Conn
	CallID   string
	Remote   string
	Local    string
	CSeq     uint32
	Messages []Message
	Parser   *Parser

	via          string
	viaBranch    string
	client       *SipClient
	lastRequest  *Message
	lastResponse *Message
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

func CreateDialog(socket net.Conn, sipClient *SipClient) *Dialog {
	d := Dialog{}
	d.CSeq = 100
	d.Conn = socket
	d.Remote = ""
	d.Local = ""
	d.viaBranch = RandSeq(10)
	d.Parser = NewParser(d.Conn)
	d.Parser.StartParsing()
	d.client = sipClient
	return &d
}

func (d *Dialog) OnMessage(callback Callback) {
	d.Parser.SetCallback(func(m *Message) {
		d.Remote = m.GetFrom()
		d.CallID = m.GetCallId()
		d.via = m.GetVia()
		switch m.GetType() {
		case RESPONSE:
			d.lastResponse = m
		case REQUEST:
			d.lastRequest = m
		}
		callback(m)
	})
}

func (d *Dialog) sendMessage(m *Message) {
	d.Conn.Write([]byte(m.String()))
}

func (d *Dialog) Reply100Trying() {
	to := d.lastRequest.GetTo()
	from := d.lastRequest.GetFrom()

	c := CreateResponse(100, "Trying")
	c.SetViaValue(d.via)
	c.SetFromValue(from)
	c.SetToValue(to)
	c.SetCallId(d.CallID)

	cseqNum, cseqVerb := d.lastRequest.GetCSeq()
	c.SetCSeq(cseqNum, cseqVerb)
	c.SetContentLength(0)
	d.sendMessage(&c)
}

func (d *Dialog) Reply180Ringing() {
	to := d.lastRequest.GetTo()
	from := d.lastRequest.GetFrom()

	c := CreateResponse(180, "Ringing")
	c.SetViaValue(d.via)
	c.SetFromValue(from)
	c.SetToValue(to)
	c.SetCallId(d.CallID)

	cseqNum, cseqVerb := d.lastRequest.GetCSeq()
	c.SetCSeq(cseqNum, cseqVerb)
	c.SetContentLength(0)
	d.sendMessage(&c)
}

func (d *Dialog) Reply200Ok() {
	to := d.lastRequest.GetTo()
	from := d.lastRequest.GetFrom()

	c := CreateResponse(200, "OK")
	c.SetViaValue(d.via)
	c.SetFromValue(from)
	c.SetToValue(to)
	c.SetCallId(d.CallID)

	cseqNum, cseqVerb := d.lastRequest.GetCSeq()
	c.SetCSeq(cseqNum, cseqVerb)
	c.SetContentLength(0)
	d.sendMessage(&c)
}

func (d *Dialog) SendRegister(registerInfo *RegisterInfo, authInfo *AuthInformation) {
	d.client.registerInfo = registerInfo
	clientCI := registerInfo.Client
	registrarCI := registerInfo.Registrar
	userName := registerInfo.Username

	c := CreateRequest("REGISTER", "sip:"+registerInfo.Registrar.Host)

	d.CallID = RandSeq(10)
	crtCSeq := d.cseq()

	c.SetVia(clientCI.Transport, clientCI.Host, clientCI.Port, d.viaBranch)
	if d.Local == "" {
		fromTag := RandSeq(10)
		c.SetFrom("sip", userName, registrarCI.Host, fromTag)
		d.Local = c.GetFrom()
	} else {
		c.SetFromValue(d.Local)
	}

	if d.Remote == "" {
		c.SetTo("sip", userName, registrarCI.Host, "")
		d.Remote = c.GetTo()
	} else {
		c.SetToValue(d.Remote)
	}
	c.SetCallId(d.CallID)
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
