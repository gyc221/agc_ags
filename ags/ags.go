package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8888")
	if err != nil {
		fmt.Printf("net.Listen fail, err: %v\n", err)
		return
	}
	for {
		client, err := listener.Accept()
		if err != nil {
			fmt.Printf("listener.Accept fail, err: %v\n", err)
			continue
		}
		go process(client)
	}
}

func process(client net.Conn) {
	defer client.Close()
	for {
		req := MyRequestNew()
		if !req.ReadAll(client) {
			break
		}

		server, err := net.Dial("tcp", req.addr)
		if err != nil {
			fmt.Printf("net.Dial failed![%s]\n", req.addr)
			break
		}
		//req.PrintInfo()

		if req.IsHTTPSRequest() {
			if !writeHTTPSSuccess(client) {
				return
			}

			go myHTTPSUpCopy(server, client)
			myDownCopy(client, server)
		} else {
			mySend(server, req.GetBuffer())
			go myHTTPUpCopy(server, client)
			myDownCopy(client, server)
		}
	}
}

func writeHTTPSSuccess(client net.Conn) bool {
	retStr := "HTTP/1.1 200 Connection established\r\n\r\n"
	bs := []byte(retStr)

	nlen := len(bs)
	var b2 [2]byte
	b2[0] = byte(nlen / 256)
	b2[1] = byte(nlen % 256)

	_, err := mySend(client, b2[0:])
	if err != nil {
		fmt.Println("https first wirte response error!")
		return false
	}

	myencode(bs[0:])
	_, err = mySend(client, bs)
	if err != nil {
		fmt.Println("https first wirte response error!")
		return false
	}
	return true
}

func myHTTPUpCopy(dst net.Conn, src net.Conn) {
	for {
		req := MyRequestNew()
		if !req.ReadAll(src) {
			break
		}
		_, err := mySend(dst, req.GetBuffer())
		if err != nil {
			break
		}
	}
}

func myHTTPSUpCopy(dst net.Conn, src net.Conn) {
	var b [40960]byte
	var b2 [2]byte
	for {
		n, err := myRecv(src, 2, b2[0:])
		if err != nil || n != 2 {
			return
		}
		needLen := int(b2[0])*256 + int(b2[1])

		n1, err := myRecv(src, needLen, b[0:])
		if err != nil || n1 != needLen {
			fmt.Println("myUpCopy myRecv error")
			return
		}
		mydecode(b[0:needLen])

		_, err = mySend(dst, b[0:needLen])
		if err != nil {
			fmt.Println("myUpCopy mySend error")
			return
		}
	}
}

func myDownCopy(dst net.Conn, src net.Conn) {
	var b [40960]byte
	var b2 [2]byte
	for {
		n, err := src.Read(b[0:])
		if err != nil || n <= 0 {
			fmt.Println("myDownCopy src.Read error")
			return
		}

		b2[0] = byte(n / 256)
		b2[1] = byte(n % 256)
		_, err = mySend(dst, b2[0:])

		myencode(b[0:n])
		_, err = mySend(dst, b[0:n])
		if err != nil {
			fmt.Println("myDownCopy mySend error")
			return
		}
	}
}
