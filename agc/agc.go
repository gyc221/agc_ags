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
	}
	//runtime.GOMAXPROCS(1)
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
		fmt.Printf("connect failed, err : %v\n", err.Error())
		return
	}
	defer server.Close()

	go myDownCopy(client, server)
	myUpCopy(server, client)
}

func myUpCopy(dst net.Conn, src net.Conn) {
	var b [40960]byte
	var b2 [2]byte
	for {
		n, err := src.Read(b[0:])
		if err != nil || n <= 0 {
			return
		}

		b2[0] = byte(n / 256)
		b2[1] = byte(n % 256)
		mySend(dst, b2[0:])

		myencode(b[0:n])
		_, err = mySend(dst, b[0:n])

		// mydecode(b[0:n])
		// str := string(b[0:n])
		// fmt.Println(str)

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
	var b2 [2]byte
	for {
		n, err := myRecv(src, 2, b2[0:])
		if err != nil || n != 2 {
			return
		}

		needLen := int(b2[0])*256 + int(b2[1])
		n1, err := myRecv(src, needLen, b[0:])
		if err != nil || n1 != needLen {
			return
		}

		mydecode(b[0:n1])
		//fmt.Println(string(b[0:n1]))
		_, err = mySend(dst, b[0:n1])
		if err != nil {
			return
		}
	}
}
