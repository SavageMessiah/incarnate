package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/savagemessiah/incarnate/messages"
	"strings"
	"sync"
)

var upgrader = websocket.Upgrader{}

type server struct {
	folks map[string]chan messages.Command
	m     sync.Mutex
}

func (s *server) do(name, command string) {
	switch {
	case strings.HasPrefix(command, "say "):
		s.broadcast(name, name+" bleats out \""+command[4:]+"\"")
		s.respond(name, "wheat", "you say a thing")
	default:
		s.respond(name, "piss", "your brain is dumb")
	}
}

func (s *server) respond(name, ctx, body string) {
	s.m.Lock()
	var cmd messages.Command
	cmd.Command.Response = &messages.Response{
		Context: ctx,
		Body:    body,
	}
	s.folks[name] <- cmd
	s.m.Unlock()
}

func (s *server) broadcast(skip, stuff string) {
	s.m.Lock()
	for n, c := range s.folks {
		if n == skip {
			continue
		}
		var cmd messages.Command
		cmd.Broadcast = &stuff
		c <- cmd
	}
	s.m.Unlock()
}

func (s *server) connect(name string, c chan messages.Command) {
	log.Println("connecting:", name)
	s.broadcast(name, "watch out, "+name+" has dragged his worthless ass in")
	var cmd messages.Command
	cmd.Command.Response = &messages.Response{
		Context: "welcum to imp zone",
		Body:    name + ", regrettably you haveentered imp zone",
	}
	c <- cmd
	s.m.Lock()
	s.folks[name] = c
	s.m.Unlock()
}

func (s *server) disconnect(name string) {
	s.broadcast(name, "hooray, "+name+" has dragged his worthless ass out")
	s.m.Lock()
	delete(s.folks, name)
	s.m.Unlock()
}

func (s *server) loginhandler(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("name")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	ch := make(chan messages.Command, 1)
	s.connect(user, ch)

	go func() {
		log.Println("writer starting:", user)
		for {
			cmd := <-ch
			err = c.WriteJSON(cmd)
			if err != nil {
				log.Println("write:", err)
				s.disconnect(user)
				close(ch)
				break
			}
		}
	}()

	for {
		var req messages.Command
		err := c.ReadJSON(&req)
		if err != nil {
			log.Println("read:", err)
			s.disconnect(user)
			break
		}
		log.Printf("cmd: %s", req.Command.Request)
		s.do(user, *req.Command.Request)
	}
}

func main() {
	var serve string
	flag.StringVar(&serve, "s", ":8080", "Interface and host to serve on.")
	flag.Parse()

	server := &server{folks: make(map[string]chan messages.Command)}
	http.HandleFunc("/login", server.loginhandler)
	log.Fatal(http.ListenAndServe(serve, nil))
}
