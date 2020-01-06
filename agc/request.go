package main

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"strings"
)

//Pair Pair
type Pair struct {
	key, val string
}

//MyRequest MyRequest
type MyRequest struct {
	method   string
	url      string
	protocol string
	addr     string
	headers  []*Pair
	body     []byte
}

//IsHTTPSRequest IsHTTPSRequest
func (r *MyRequest) IsHTTPSRequest() bool {
	return (r.method == "CONNECT" || r.method == "connect")
}

//MyRequestNew MyRequestNew
func MyRequestNew() *MyRequest {
	p := &MyRequest{}
	p.headers = make([]*Pair, 0)
	return p
}

//AddHeader AddHeader
func (r *MyRequest) AddHeader(k, v string) {
	r.headers = append(r.headers, &Pair{k, v})
}

//PrintInfo PrintInfo
func (r *MyRequest) PrintInfo() {
	fmt.Println("\n\nstart------------------------------------------------------")
	fmt.Printf("->addr:%s\n", r.addr)
	fmt.Printf("->%s %s %s\n", r.method, r.url, r.protocol)
	for i := 0; i < len(r.headers); i++ {
		fmt.Printf("->%s:%s\n", r.headers[i].key, r.headers[i].val)
	}
	fmt.Println("end------------------------------------------------------------")
}

func getFirstSubStr(b []byte, endChar byte) (string, int) {
	nlen := len(b)
	startIdx := 0
	isCutBlank := true

	for i := 0; i < nlen; i++ {
		if isCutBlank {
			if b[i] == ' ' || b[i] == '\r' || b[i] == '\n' || b[i] == ':' {
				startIdx++
				continue
			}
			isCutBlank = false
		}

		if b[i] == endChar {
			return string(b[startIdx:i]), i
		}
	}
	return "", -1
}

//GetBuffer GetBuffer
func (r *MyRequest) GetBuffer() []byte {
	var buffer bytes.Buffer

	var addr string
	u, err := url.Parse(r.url)
	if err != nil {
		return nil
	}
	addr = u.Path
	if len(u.RawQuery) > 0 {
		addr += ("?" + u.RawQuery)
	}

	buffer.WriteString(fmt.Sprintf("%s %s %s\r\n", r.method, addr, r.protocol))

	for i := 0; i < len(r.headers); i++ {
		realKey := r.headers[i].key
		if strings.Index(realKey, "Proxy-") >= 0 {
			realKey = strings.ReplaceAll(realKey, "Proxy-", "")
		}
		buffer.WriteString(fmt.Sprintf("%s: %s\r\n", realKey, r.headers[i].val))
	}
	buffer.WriteString("\r\n")
	return buffer.Bytes()
}

//ParseRequest ParseRequest
func (r *MyRequest) ParseRequest(req []byte) bool {
	startIdx := 0

	//parse request method
	str, curIndex := getFirstSubStr(req[0:], ' ')
	if curIndex < 0 {
		return false
	}
	startIdx += curIndex
	r.method = str

	//parse request url
	str, curIndex = getFirstSubStr(req[startIdx:], ' ')
	if curIndex < 0 {
		return false
	}
	startIdx += curIndex
	r.url = str

	//parse addr for server connection
	var addr string
	if r.IsHTTPSRequest() {
		addr = r.url
		if strings.LastIndex(addr, ":") < 0 {
			addr += ":443"
		}
	} else {
		u, err := url.Parse(r.url)
		if err != nil {
			return false
		}
		addr = u.Host
		if strings.LastIndex(addr, ":") < 0 {
			addr += ":80"
		}
	}
	r.addr = addr

	//pase http/1.1 or http/1.0
	str, curIndex = getFirstSubStr(req[startIdx:], '\r')
	if curIndex < 0 {
		return false
	}
	startIdx += curIndex
	r.protocol = str

	//pase key:value
	for {
		str, curIndex = getFirstSubStr(req[startIdx:], ':')
		if curIndex < 0 {
			break
		}
		startIdx += curIndex
		k := str

		str, curIndex = getFirstSubStr(req[startIdx:], '\r')
		if curIndex < 0 {
			return false
		}
		startIdx += curIndex
		v := str

		r.AddHeader(k, v)
	}
	return true
}

//ReadAll ReadAll
func (r *MyRequest) ReadAll(client net.Conn) bool {
	var b [40960]byte
	haveRecved := 0
	idx := 0

	for {
		n, err := client.Read(b[haveRecved:])
		if err != nil || n <= 0 {
			return false
		}
		haveRecved += n

		strReq := string(b[0:haveRecved])
		idx = strings.Index(strReq, "\r\n\r\n")
		if idx <= 0 {
			if haveRecved < len(b) {
				continue
			} else {
				fmt.Println("ReadAll::can not find rnrn")
				return false
			}
		}
		idx += 3
		break
	}

	return r.ParseRequest(b[0:idx])
}
