package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StoreData struct {
	value string
	expiresAt time.Time
}

type Store struct {
	mu sync.RWMutex
	data map[string]StoreData
}

func (s *Store) Set(key string, value string) (string) {
	s.mu.Lock()
	s.data[key] = StoreData{
		value: value,
	}
	s.mu.Unlock()
	s.Expire(key, 5)
	return "OK"
}

func (s *Store) Get(key string) (string) {
	s.mu.RLock()
	
	storeData, ok := s.data[key]
	
	s.mu.RUnlock()

	if !ok {
		return "ERR data doesn't exist"
	}

	expired := s.TTL(key)

	if expired == "-1" {
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

func (s *Store) Expire(key string, seconds int) (string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.data[key]

	if !ok {
		return "ERR data not found"
	}

	value.expiresAt = time.Now().Add(time.Second * time.Duration(seconds))
	s.data[key] = value

	return "OK"
}

func (s *Store) TTL(key string) (string) {
	s.mu.RLock()
	
	value, ok := s.data[key]

	s.mu.RUnlock()

	if !ok {
		return  "-1"
	}

	if value.expiresAt.IsZero() {
		return "Data never expires"
	}

	diff := time.Until(value.expiresAt)

	if diff <= 0 {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.data, key)
		return "-1"
	}

	return strconv.Itoa(int(diff.Seconds()))
}

func (s *Store) StartJanitor(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for range ticker.C {
			s.cleanup()
		}
	}()
}

func (s *Store) cleanup() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.data {
		if !v.expiresAt.IsZero() && now.After(v.expiresAt) {
			delete(s.data, k)
		}
	}
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

	store.StartJanitor(time.Duration(time.Second * 3))

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn, store)
	}
}