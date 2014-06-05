package main

// This file contains the definitions needed for a multiplayer Life server

import (
	"encoding/json"
	"errors"
	"github.com/garyburd/go-websocket/websocket"
	"io/ioutil"
	"life"
	"log"
	"net/http"
	"time"
)

type Message struct {
	Command string
}

type WorldMessage struct {
	Message
	World     *life.World
	SendCount int
}

type Server struct {
	Clients               []*Client     // All clients connected to the server
	Worlds                []*life.World // All worlds hosted by the server
	NumClients, NumWorlds int
}

type Client struct {
	RemoteAddr string          // IP of the remote client
	Header     http.Header     // Headers from the HTTP request that was upgraded
	Websocket  *websocket.Conn // The client's websocket
	World      *life.World     // The client's currently subscribed world
	SendCount  int
}

var DefaultServer = &Server{make([]*Client, 10), make([]*life.World, 10), 0, 0}

// Upgrades all HTTP connections to websocket and begins serving Game of Life.
func (server *Server) httpHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	// First validate the HTTP request.
	if r.Method != "GET" {
		err = errors.New("Method Not Allowed")
		http.Error(w, err.Error(), 405)
		log.Println(err)
		return
	}
	if r.Header.Get("Origin") != "http://"+r.Host {
		err = errors.New("Origin Not Allowed")
		http.Error(w, err.Error(), 403)
		log.Println(err)
		return
	}

	// Everything checks out; upgrade to WebSocket protocol.
	var ws *websocket.Conn
	ws, err = websocket.Upgrade(w, r.Header, nil, 1024, 1024)

	// Handle a handshake error or a general error.
	if _, ok := err.(websocket.HandshakeError); ok {
		err = errors.New("Not a websocket handshake")
		http.Error(w, err.Error(), 400)
		log.Println(err)
		return
	} else if err != nil {
		err = errors.New("Internal Server Error")
		http.Error(w, err.Error(), 500)
		log.Println(err)
		return
	}

	// No errors, register a new client.
	client := &Client{}
	client.RemoteAddr = r.RemoteAddr
	client.Header = r.Header
	client.Websocket = ws

	server.Clients = append(server.Clients, client)
	server.NumClients++

	// Create a World for the client and register it with the Server too.
	world := life.New(50, 50)
	client.World = world

	server.Worlds = append(server.Worlds, world)
	server.NumWorlds++

	err = ws.WriteMessage(websocket.OpText, []byte("Welcome to game of shitty life, where nobody loves you!"))

	// Start the session listener.
	go server.ListenToClient(client)

	// TODO: This should be moved into methods that modify NumClient and NumWorlds.
	log.Println(server.NumClients, "clients and", server.NumWorlds, "worlds")
}

// ListenToClient starts a loop which handles websocket messages from a client.
func (s *Server) ListenToClient(c *Client) {
	defer func() {
		// Panic errors crash the session with a log message.
		r := recover()
		if err, ok := r.(error); ok {
			log.Println("Listener panic:", err)
		}
	}()

	log.Println("Launched listener for", c.RemoteAddr)

	// Declare a buffer for incoming messages.
	var msg []byte

	// Do an infinite read loop.
	for {
		// Get the next message reader.
		msgOpCode, msgReader, err := c.Websocket.NextReader()
		if err != nil {
			panic(err)
		}

		// Ignore non-text message types.
		if msgOpCode != websocket.OpText {
			log.Printf("Recieved op code %v from %v", msgOpCode, c.RemoteAddr)
			continue
		}

		// Build the full message from multiple packets.
		msg, err = ioutil.ReadAll(msgReader)
		if err != nil {
			panic(err)
		}

		// Route the message to the proper handler.
		baseMsg := Message{}

		if err = json.Unmarshal(msg, &baseMsg); err != nil {
			panic(err)
		}

		switch baseMsg.Command {
		case "set":
			// Permit the client to set the world state.
			worldMsg := WorldMessage{}

			if err = json.Unmarshal(msg, &worldMsg); err != nil {
				panic(err)
			}

			log.Printf("Setting world state for %s", c.RemoteAddr)

			log.Println("Alive cells =", worldMsg.World.NumAlive())

			if worldMsg.World == nil {
				log.Println("No World attribute recieved")
				return
			}

			c.World = worldMsg.World

			// Start the session world updater.
			go s.SendWorldUpdatesToClient(c)

		default:
			// The command was unrecognized.
			log.Println("No match (recieved: " + baseMsg.Command + ")")
		}

		log.Println("Listened to", c.RemoteAddr)
		time.Sleep(1)
	}
}

func (s *Server) SendWorldUpdatesToClient(c *Client) {
	defer func() {
		// Panic errors crash the session with a log message.
		r := recover()
		if err, ok := r.(error); ok {
			log.Println("World Sender panic:", err)
		}
	}()

	log.Println("Launched world sender for", c.RemoteAddr)

	// Subscribe to a channel of updates from the World.
	worlds, stop := c.World.Stream()

	var nextWorld life.World

	defer func() {
		stop <- true
	}()

	var worldMsg WorldMessage
	var jsonMsg []byte
	var err error

	worldMsg.Command = "update"

	// Regularly send the world to the client.
	for {
		nextWorld = <-worlds
		worldMsg.World = &nextWorld

		c.SendCount++
		worldMsg.SendCount = c.SendCount

		jsonMsg, err = json.Marshal(worldMsg)
		if err != nil {
			panic(err)
		}

		if err = c.Websocket.WriteMessage(websocket.OpText, jsonMsg); err != nil {
			panic(err)
		}

		//log.Println("Sent world to", c.RemoteAddr)
		time.Sleep(75 * time.Millisecond)
	}
}
