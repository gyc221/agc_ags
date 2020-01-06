package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func myencode(b []byte) {
	blen := len(b)
	for i := 0; i < blen-1; i++ {
		b[i] ^= b[i+1]
	}
}

func mydecode(b []byte) {
	blen := len(b)
	for i := blen - 2; i >= 0; i-- {
		b[i] ^= b[i+1]
	}
}

func loadConfig(fileName string) (string, string) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", ""
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if io.EOF != err {
				break
			}
			if len(line) <= 0 {
				break
			}
		}
		line = strings.ReplaceAll(line, "\r", "")
		line = strings.ReplaceAll(line, "\n", "")

		if strings.HasPrefix(line, "#") {
			continue
		}
		bss := strings.Split(line, ",")
		if len(bss) != 2 {
			fmt.Println("config.ini format error!")
			return "", ""
		}
		return bss[0], bss[1]
	}
	return "", ""
}

func myRecv(src net.Conn, needLen int, s []byte) (int, error) {
	receivedLen := 0
	for {
		n, err := src.Read(s[receivedLen:needLen])
		if err != nil {
			return receivedLen, err
		}
		receivedLen += n
		if receivedLen >= needLen {
			break
		}
	}
	return receivedLen, nil
}

func myRecvByLen(src net.Conn, s []byte) (int, error) {
	var b2 [2]byte
	n, err := myRecv(src, 2, b2[0:])
	if err != nil || n != 2 {
		return n, err
	}

	needLen := int(b2[0])*256 + int(b2[1])
	n1, err := myRecv(src, needLen, s)
	return n1, err
}

func myRecvByLenAndDecode(src net.Conn, s []byte) (int, error) {
	var b2 [2]byte
	n, err := myRecv(src, 2, b2[0:])
	if err != nil || n != 2 {
		return n, err
	}

	needLen := int(b2[0])*256 + int(b2[1])
	n1, err := myRecv(src, needLen, s)
	if err == nil && n1 == needLen {
		mydecode(s[0:n1])
	}
	return n1, err
}

func mySend(dst net.Conn, s []byte) (int, error) {
	needLen := len(s)
	sendedLen := 0

	for {
		n, err := dst.Write(s[sendedLen:])
		if err != nil {
			return sendedLen, err
		}
		sendedLen += n
		if sendedLen >= needLen {
			break
		}
	}
	return sendedLen, nil
}

func mySendWithLen(dst net.Conn, s []byte) (int, error) {
	var b2 [2]byte
	b2[0] = byte(len(s) / 256)
	b2[1] = byte(len(s) % 256)
	mySend(dst, b2[0:])
	return mySend(dst, s)
}

func mySendWithLenAndEnCode(dst net.Conn, s []byte) (int, error) {
	var b2 [2]byte
	b2[0] = byte(len(s) / 256)
	b2[1] = byte(len(s) % 256)
	mySend(dst, b2[0:])

	myencode(s)
	return mySend(dst, s)
}
