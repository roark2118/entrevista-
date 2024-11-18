package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"slices"
	"sync"
)

func main() {
	numChainsFlag :=
		flag.Uint("chains", 1_000_000, "number of chains to generate")
	addressFlag :=
		flag.String("socket", "localhost:3000", "socket address")
	flag.Parse()

	socket, err := net.Dial("tcp", *addressFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()
	socketReader := bufio.NewReader(socket)
	socketWriter := bufio.NewWriter(socket)

	chainsFile, err := os.Create("chains.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer chainsFile.Close()
	chainsWriter := bufio.NewWriter(chainsFile)
	defer chainsWriter.Flush()

	resultsFile, err := os.Create("results.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer resultsFile.Close()
	resultsWriter := bufio.NewWriter(resultsFile)
	defer resultsWriter.Flush()

	waitForCopy := sync.WaitGroup{}
	waitForCopy.Add(1)
	//copy from socket to results file in another thread
	go func() {
		defer waitForCopy.Done()
		_, err := io.Copy(resultsWriter, socketReader)
		if err != nil {
			log.Println("error coppying from socket to results file", err)
		}
	}()
	multiWriter := io.MultiWriter(chainsWriter, socketWriter)
	for range *numChainsFlag {
		if _, err := multiWriter.Write(generateChain()); err != nil {
			log.Fatal("Error writing chain: ", err)
		}
		if _, err := multiWriter.Write([]byte{'\n'}); err != nil {
			log.Fatal("Error writing newline: ", err)
		}
	}
	socketWriter.WriteByte('\n')
	socketWriter.Flush()
	waitForCopy.Wait()
}

func randomRange(start, end int) int {
	return rand.N(end+1-start) + start
}

func generateChain() []byte {
	const validChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const maxValidCharIndex = len(validChars) - 1
	chainLength := randomRange(50, 100)
	chain := make([]byte, chainLength)
	for _, index := range generateSpaceIndexes(chainLength) {
		chain[index] = ' '
	}
	for i, c := range chain {
		if c == ' ' {
			continue
		}
		chain[i] = validChars[randomRange(1, maxValidCharIndex)]
	}
	return chain
}

func generateSpaceIndexes(chainLength int) []int {
	numSpaces := randomRange(3, 5)
	validSpaces := make([]int, 0, numSpaces)
	validSpaces = append(validSpaces, randomRange(1, chainLength-2))
	for range numSpaces {
		for {
			choice := randomRange(1, chainLength-2)
			if !(slices.Contains(validSpaces, choice) ||
				slices.Contains(validSpaces, choice-1) ||
				slices.Contains(validSpaces, choice+1)) {
				validSpaces = append(validSpaces, choice)
				break
			}
		}
	}
	return validSpaces
}
