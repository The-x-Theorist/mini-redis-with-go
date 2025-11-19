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

type StoreData struct {
	value string
	ttl time.Time
}

type Store struct {
	mu sync.RWMutex
	data map[string]StoreData
}

func (s *Store) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = StoreData{
		value: value,
		ttl: time.Now().Add(time.Millisecond * 5000),
	}
}

func (s *Store) Get(key string) (string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storeData, ok := s.data[key]

	if !ok {
		return "ERR data doesn't exist"
	}

	if time.Now().After(storeData.ttl) {
		delete(s.data, key)
		return "ERR data expired"
	}

	return storeData.value
}

func (s *Store) Del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Store) Exists(key string) (bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.data[key]
	return exists
}

func (s *Store) Execute(command string, args []string) (string) {
	switch command {
    case "PING":
        return "PONG"
    case "SET":
		if len(args) != 2 {
			return "ERR wrong number of arguments for 'set' command"
		}
		s.Set(args[0], args[1])
		return "OK"
	case "GET":
		if len(args) != 1 {	
			return "ERR wrong number of arguments for 'get' command"
		}
		variable := s.Get(args[0])
		if variable == "" {
			return "ERR property doesn't exist in store"
		}
		return variable
	case "DEL":
		if len(args) != 1 {
			return "ERR wrong number of arguments for 'del' command"
		}
		s.Del(args[0])
		return "OK"
	case "EXISTS":
		if len(args) != 1 {
			return "ERR wrong number of arguments for 'exists' command"
		}
		exists := s.Exists(args[0])
		if exists {
			return "Yes"
		}
		return "No"
    default:
        return "ERR unknown command"
    }

}

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

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn, store)
	}
}