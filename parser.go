package sip

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"log"
)

type ParserState int

const (
	FIRST_LINE ParserState = iota
	HEADERS
	BODY
)

type Parser struct {
	state     bool
	reader    io.Reader
	bufReader *bufio.Reader

	callback Callback
}

type Callback func(*Message)

func NewParser(reader io.Reader) *Parser {
	p := &Parser{}
	p.state = false
	p.reader = reader
	p.bufReader = bufio.NewReader(p.reader)

	return p
}

func (p *Parser) SetCallback(newCallback Callback) {
	p.callback = newCallback
}

func (p *Parser) StartParsing() {
	go p.parse()
}

func (p *Parser) parse() {
	go func() {
		state := FIRST_LINE // 0 = first line, 1 = headers, 2 = content
		var message Message
		var toRead int
		for {
			switch state {
			case FIRST_LINE:
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
				state = HEADERS
			case HEADERS:
				line, err := Readln(p.bufReader)
				if err != nil {
					log.Println("Error: ", err)
				}
				if line == "" {
					state = BODY
					continue
				}

				headerLine := strings.Split(line, ": ")
				headerName := headerLine[0]
				headerValue := headerLine[1]
				if headerName == "Content-Length" {
					toRead, _ = strconv.Atoi(headerValue)
				}
				message.Headers.AddHeader(headerName, headerValue)
			case BODY:
				if toRead > 0 {
					crtByte, err := p.bufReader.ReadByte()
					if err != nil {
						log.Println("Error: ", err)
					}
					message.Body = append(message.Body, crtByte)
					toRead--
				} else {
					p.callback(&message)
					state = FIRST_LINE
				}

			}

		}
	}()
}
