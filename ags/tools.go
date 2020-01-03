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
