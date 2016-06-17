package sipbase

import (
	"fmt"
)

// ---------------
type RequestHeadline struct {
	Method  string
	Uri     SipUri
	Version string
}

func CreateRequestHeadline(method string, uri SipUri, version string) (line RequestHeadline) {
	r := RequestHeadline{method, uri, version}
	return r
}

func (r RequestHeadline) ToString() string {
	return fmt.Sprintf("%s %s %s", r.Method, r.Uri.String(), r.Version)
}

func CreateRequest(method string, uri string) Message {
	r := Message{}
	r.MessageType = REQUEST
	r.Body = []byte("")
	r.Headline = CreateRequestHeadline(method, ParseSipUri(uri), "SIP/"+sipversion)
	r.SetRequestId(RandSeq(10))
	return r
}
