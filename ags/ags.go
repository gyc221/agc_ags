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
	svrAddr := getServerAddr(client)
	if len(svrAddr) == 0 {
		fmt.Println("can not find server addr[]!")
		return
	}

	server, err := net.Dial("tcp", svrAddr)
	if err != nil {
		fmt.Printf("net.Dial failed![%s]\n", svrAddr)
		return
	}
	defer server.Close()

	go myUpCopy(server, client)
	myDownCopy(client, server)

}

func getServerAddr(client net.Conn) string {
	var b [40960]byte
	n, err := myRecvByLenAndDecode(client, b[0:])
	if err != nil && n <= 0 {
		return ""
	}
	return string(b[0:n])
}

func myUpCopy(dst net.Conn, src net.Conn) {
	var b [40960]byte
	for {
		n, err := myRecvByLenAndDecode(src, b[0:])
		if err != nil || n <= 0 {
			return
		}

		_, err = mySend(dst, b[0:n])
		if err != nil {
			break
		}
	}
}

func myDownCopy(dst net.Conn, src net.Conn) {
	var b [40960]byte
	for {
		n, err := src.Read(b[0:])
		if err != nil || n <= 0 {
			fmt.Println("myDownCopy src.Read error")
			return
		}
		_, err = mySendWithLenAndEnCode(dst, b[0:n])
		if err != nil {
			fmt.Println("myDownCopy mySend error")
			return
		}
	}
}
