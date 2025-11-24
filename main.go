package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func handleConnection(conn net.Conn, store *Store) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
	
		parts := strings.Fields(line)
		cmd := strings.ToUpper(parts[0])
		args := parts[1:]
	
		resp := store.Execute(cmd, args)
		fmt.Fprintln(conn, resp)
	}
	
}

func main() {
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	store := &Store{
		mu: sync.RWMutex{},
		data: make(map[string]StoreData),
	}

	store.StartJanitor(time.Duration(time.Second * 3))

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn, store)
	}
}