package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	laddr, raddr := loadConfig("config.ini")
	if len(laddr) == 0 || len(raddr) == 0 {
		fmt.Println("load config.ini error!")
		return
	}

	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		fmt.Printf("net.Listen err: %v\n", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("listener.Accept err: %v\n", err)
			continue
		}
		go process(conn, raddr)
	}
}

func process(client net.Conn, svrAddr string) {
	defer client.Close()

	server, err := net.Dial("tcp", svrAddr)
	if err != nil {
		fmt.Printf("net.Dial failed[%s], err : %v\n", svrAddr, err.Error())
		return
	}
	defer server.Close()

	req := MyRequestNew()
	if !req.ReadAll(client) {
		return
	}

	if req.IsHTTPSRequest() {
		if !writeHTTPSSuccess(client, server, req) {
			return
		}
		go myHTTPSUpCopy(server, client)
		myDownCopy(client, server)
	} else {
		writeFirstHTTPToServer(server, req)
		go myHTTPUpCopy(server, client)
		myDownCopy(client, server)
	}
}

func writeHTTPSSuccess(client net.Conn, server net.Conn, r *MyRequest) bool {
	_, err := mySend(client, []byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		fmt.Println("https first wirte response error!")
		return false
	}
	bsHead := []byte(r.addr)
	_, err = mySendWithLenAndEnCode(server, bsHead)
	return err == nil

}

func writeFirstHTTPToServer(server net.Conn, r *MyRequest) bool {
	bsHead := []byte(r.addr)
	mySendWithLenAndEnCode(server, bsHead)

	bsBody := r.GetBuffer()
	n, err := mySendWithLenAndEnCode(server, bsBody)
	return (err == nil && n > 0)
}

func myHTTPUpCopy(dst net.Conn, src net.Conn) {
	for {
		req := MyRequestNew()
		if !req.ReadAll(src) {
			break
		}

		_, err := mySendWithLenAndEnCode(dst, req.GetBuffer())
		if err != nil {
			break
		}
	}
}

func myHTTPSUpCopy(dst net.Conn, src net.Conn) {
	var b [40960]byte
	for {
		n, err := src.Read(b[0:])
		if err != nil || n <= 0 {
			return
		}

		_, err = mySendWithLenAndEnCode(dst, b[0:n])
		if err != nil {
			if io.EOF != err {
				fmt.Printf("server connetion write error[%v]", err)
			} else {
				fmt.Println("server connection write closed!")
			}
			return
		}
	}
}

func myDownCopy(dst net.Conn, src net.Conn) {
	var b [40960]byte
	for {
		n, err := myRecvByLenAndDecode(src, b[0:])
		if err != nil || n <= 0 {
			break
		}
		_, err = mySend(dst, b[0:n])
		if err != nil {
			break
		}
	}
}
