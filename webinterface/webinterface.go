package webinterface

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/gopistolet/gopistolet/helpers"
	"github.com/gopistolet/gopistolet/mta"
	"html/template"
	"log"
	"net/http"
	"syscall"
	"os"
)

var config mta.Config

func addKey(obj interface{}, key string) interface{} {
	return struct {
		name   string
		object interface{}
	}{
		name:   key,
		object: obj,
	}
}

func handler(w http.ResponseWriter, r *http.Request) {

	// Parse template file.
	t, _ := template.ParseFiles("webinterface/index.html")
	t.Execute(w, nil)

	//html := ""
	//log.Fprintf(w, html)
}

type Msg struct {
	Method  string
	Content interface{}
}

func createJsonMessage(m string, c interface{}) (string, error) {
	goMsg := Msg{Method: m, Content: c}
	jsonMsg, err := json.MarshalIndent(&goMsg, "", "    ")
	if err != nil {
		return "", err
	} else {
		return string(jsonMsg), err
	}

}

func handleSocket(ws *websocket.Conn) {

	// Send config to client
	msg, err := createJsonMessage("getConfig", config)
	if err != nil {
		log.Println("Can't marshal: " + err.Error())
	}
	err = websocket.Message.Send(ws, string(msg))
	if err != nil {
		log.Println("Can't send: " + err.Error())
	}
	log.Println("Sent config to client: " + string(msg))

	// Keep listening on socket
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			log.Println("Can't receive: " + err.Error())
			break
		}

		if reply == "restart" {
			log.Println("Restart issued...")
			syscall.Exec(os.Args[0], os.Args, os.Environ())
		}

		// In this naive version we handle all received
		// on the socket as a mta.Config object
		log.Println("Received from client: " + reply)

		err = json.Unmarshal([]byte(reply), &config)
		if err != nil {
			log.Println(err)
		}

		err := helpers.EncodeFile("config.json", &config)
		notify := ""
		if err != nil {
			notify = "Config not saved: " + err.Error()
		} else {
			notify = "Config saved. Restart GoPistolet to apply changes."
		}

		msg, err := createJsonMessage("getNotify", notify)
		if err != nil {
			log.Println("Can't marshal: " + err.Error())
		}
		err = websocket.Message.Send(ws, string(msg))
		if err != nil {
			log.Println("Can't send: " + err.Error())
		}

	}
}

func Webinterface() {

	// Load config from JSON file
	err := helpers.DecodeFile("config.json", &config)
	if err != nil {
		log.Println(err)
	}

	http.HandleFunc("/", handler)
	http.Handle("/ws", websocket.Handler(handleSocket))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("webinterface/assets"))))
	http.ListenAndServe(":8080", nil)
}
