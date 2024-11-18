package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"
	"unicode"
)

func main() {
	address := flag.String("socket", "localhost:3000", "socket address")
	logFilename := flag.String("log", "server.log", "log file")
	flag.Parse()
	stat, err := os.Stat(*logFilename)
	if err == nil {
		err = os.Rename(*logFilename,
			fmt.Sprintf("%s_%s", *logFilename, stat.ModTime().Format("2006_01_02_15_04")))
		if err != nil {
			log.Fatal("error while renaming "+*logFilename, err)
		}
	}
	logFile, err := os.OpenFile(*logFilename,
		os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(logFile, nil)))
	ln, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("server started at " + *address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(socket net.Conn) {
	slog.Info(fmt.Sprintf("connection from %s", socket.RemoteAddr()))
	start := time.Now()
	socketWriter := bufio.NewWriter(socket)
	socketReader := bufio.NewReader(socket)
	defer func() {
		socketWriter.Flush()
		slog.Info(fmt.Sprintf("process from %s take %s", socket.RemoteAddr(), time.Since(start)))
		socket.Close()
	}()
	for {
		chain, err := socketReader.ReadString('\n')
		if err != nil {
			slog.Error(err.Error())
			return
		}
		chain = strings.TrimSpace(chain)
		if len(chain) == 0 {
			return
		}
		weigth := getChainWeigth(chain)
		if weigth == 1000 {
			continue
		}
		message := fmt.Sprintf("%s : %.2f\n", chain, weigth)
		socketWriter.WriteString(message)
	}
}

func getChainWeigth(chain string) float64 {
	var digits, letters, spaces int

	for i, r := range chain {
		if i > 0 && unicode.ToLower(r) == 'a' &&
			unicode.ToLower(rune(chain[i-1])) == 'a' {
			message := fmt.Sprintf("Double 'a' rule detected >> '%s'", chain)
			slog.Warn(message)
			return 1000
		}
		switch {
		case r == ' ':
			spaces++
		case unicode.IsDigit(r):
			digits++
		case unicode.IsLetter(r):
			letters++
		default:
			message := fmt.Sprintf("invalid character %s in chain %s",
				string(r), chain)
			slog.Error(message)
			return 1000
		}
	}
	if spaces == 0 {
		message := fmt.Sprintf("chain %s has 0 spaces",
			chain)
		slog.Error(message)
		return -1
	}
	return (float64(letters)*1.5 + float64(digits)*2) / float64(spaces)
}
