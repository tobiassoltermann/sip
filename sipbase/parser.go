package sipbase

import (
	"bufio"
	"io"
	"log"
	"strconv"
	"strings"
)

type Parser struct {
	state     bool
	reader    io.Reader
	bufReader *bufio.Reader

	callback Callback
}

type Callback func(Message)

func NewParser(reader io.Reader) Parser {
	p := Parser{}
	p.state = false
	p.reader = reader
	p.bufReader = bufio.NewReader(p.reader)

	return p
}

func (p *Parser) SetCallback(callback Callback) {
	p.callback = callback
}

func (p *Parser) StartParsing() {
	if p.callback == nil {
		log.Println("Error: Callback function must be set")
	}
	go func() {
		state := 0 // 0 = first line, 1 = headers, 2 = content
		var message Message
		var expectedContentLength int
		var toRead int = expectedContentLength
		for {
			switch state {
			case 0:
				line, err := Readln(p.bufReader)
				if err != nil {
					log.Println("Error: ", err)
				}
				elements := strings.Split(line, " ")
				if elements[0] == "SIP/2.0" {
					code, err := strconv.Atoi(elements[1])
					if err != nil {
						log.Println("Error. Code was " + elements[1] + ". Attempt to continue")
						code = 400
					}
					reply := elements[2]
					message = CreateResponse(code, reply)
				} else {
					method := elements[0]
					uri := elements[1]
					message = CreateRequest(method, uri)
				}
				state = 1
			case 1:
				line, err := Readln(p.bufReader)
				if err != nil {
					log.Println("Error: ", err)
				}
				if line == "" {
					state = 2
					continue
				}

				headerLine := strings.Split(line, ": ")
				headerName := headerLine[0]
				headerValue := headerLine[1]
				message.Headers.AddHeader(headerName, headerValue)
			case 2:
				if toRead > 0 {
					crtByte, err := p.bufReader.ReadByte()
					if err != nil {
						log.Println("Error: ", err)
					}
					message.Body = append(message.Body, crtByte)
					toRead--
				} else {
					log.Println("Message done. Emit.")
					p.callback(message)
					state = 0
				}

			}

		}
	}()
}
