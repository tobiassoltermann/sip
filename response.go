package sip

import (
	"fmt"
)

// ---------------
type ResponseHeadline struct {
	Version string
	Code    int
	Reply   string
}

func CreateResponseHeadline(version string, code int, reply string) (line ResponseHeadline) {
	r := ResponseHeadline{version, code, reply}
	return r
}
func (r *ResponseHeadline) IsFinal() bool {
	return r.Code > 199
}
func (r ResponseHeadline) ToString() string {
	return fmt.Sprintf("%s %d %s", r.Version, r.Code, r.Reply)
}

func CreateResponse(code int, reply string) Message {
	r := Message{}
	r.MessageType = RESPONSE
	r.Body = []byte("")
	r.Headline = CreateResponseHeadline("SIP/"+sipversion, code, reply)
	r.SetRequestId(RandSeq(10))
	return r
}
