package main

import (
	"flag"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/savagemessiah/incarnate/messages"
	"gopkg.in/readline.v1"
	"io"
	"log"
	"net/url"
)

func main() {
	var server, name string
	flag.StringVar(&server, "s", "localhost:8080", "The server host.")
	flag.StringVar(&name, "n", "", "Your name")
	flag.Parse()
	if server == "" || name == "" {
		log.Fatal("You must provide both the server url and your name")
	}

	u := url.URL{Scheme: "ws", Host: server, Path: "/login", RawQuery: fmt.Sprintf("name=%s", name)}
	log.Printf("connecting to %s", u)

	pc := make(chan string)
	line, err := readline.New("> ")
	if err != nil {
		log.Fatal("line:", err)
	}
	defer line.Close()
	log.SetOutput(line.Stderr())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	go func() {
		for {
			var cmd messages.Command
			err := c.ReadJSON(&cmd)
			if err != nil {
				log.Print("read:", err)
				continue
			}
			switch {
			case cmd.Broadcast != nil:
				fmt.Fprintln(line.Stdout(), *cmd.Broadcast)
			case cmd.Command.Response != nil:
				fmt.Fprintln(line.Stdout(), cmd.Command.Response.Body)
				pc <- cmd.Command.Response.Context
			}
		}
	}()

	for {
		prompt := <-pc
		line.SetPrompt(prompt + " >> ")
		l, err := line.Readline()
		if err == io.EOF {
			break
		}
		var cmd messages.Command
		cmd.Command.Request = &l
		c.WriteJSON(cmd)
		if l == "quit" {
			break
		}
	}
}
