package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Rooms struct {
	mtx *sync.RWMutex
	m   map[string][]*websocket.Conn
}

func (r *Rooms) Remove(path string, conn *websocket.Conn) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	conns, ok := r.m[path]
	if !ok {
		return
	}
	newConns := make([]*websocket.Conn, 0, len(conns))
	for _, c := range conns {
		if c == conn {
			conn.Close()
		} else {
			newConns = append(newConns, c)
		}
	}
	if len(newConns) == 0 {
		delete(r.m, path)
	}
}

func (r *Rooms) Add(path string, conn *websocket.Conn) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	conns, ok := r.m[path]
	if !ok {
		conns = []*websocket.Conn{}
	}
	r.m[path] = append(conns, conn)
}

func (r *Rooms) ReadMessage(path string, hostConn *websocket.Conn, recast bool) error {
	messageType, p, err := hostConn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read from connection: %w", err)
	}

	r.mtx.RLock()
	defer r.mtx.RUnlock()

	conns, ok := r.m[path]
	if !ok {
		return fmt.Errorf("no room for path %s", path)
	}
	for _, conn := range conns {
		if conn == hostConn && !recast {
			continue
		}
		err := conn.WriteMessage(messageType, p)
		if !ok {
			log.Printf("could not write to connection: %v", err)
		}
	}
	return nil
}

func JoinRoom() http.Handler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	rooms := &Rooms{
		mtx: &sync.RWMutex{},
		m:   map[string][]*websocket.Conn{},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recast := r.URL.Query().Has("reacst")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print(err)
			fmt.Fprintln(w, err.Error())
			w.WriteHeader(500)
			return
		}

		rooms.Add(r.URL.Path, conn)

		defer func() {
			log.Printf("close %s", r.URL.Path)
			conn.Close()
			rooms.Remove(r.URL.Path, conn)
		}()

		for {
			rooms.ReadMessage(r.URL.Path, conn, recast)
		}
	})
}
