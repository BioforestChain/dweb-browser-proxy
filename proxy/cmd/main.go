package main

import (
	"dweb-proxy/ws"
	"encoding/json"
	"fmt"
	"ipc"
	"log"
	"net/http"
)

var hub *ws.Hub

func main() {
	hub = ws.NewHub()
	go hub.Run()

	http.HandleFunc("/ipc/test", serveApi)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	err := http.ListenAndServe("127.0.0.1:8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveApi(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/ipc/test" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowd", http.StatusMethodNotAllowed)
		return
	}

	client := hub.GetClient("test")
	if client == nil {
		fmt.Fprintf(w, `{"msg": "%s"}`, "The service is unavailable")
		return
	}

	clientIpc := client.GetIpc()
	req := clientIpc.Request("https://www.example.com/search?p=feng", ipc.RequestArgs{
		Method: "GET",
		Header: map[string]string{"Content-Type": "application/json"},
	})
	res, err := clientIpc.Send(req)
	if err != nil {
		log.Println("ipc response err: ", err)
		fmt.Fprintf(w, `{"msg": "%s"}`, err.Error())
		return
	}
	for k, v := range res.Header {
		w.Header().Set(k, v)
	}

	resStr, err := json.Marshal(res)
	if err != nil {
		fmt.Fprintf(w, `{"msg": "%s"}`, err.Error())
		return
	}

	fmt.Fprintf(w, `{"msg": %s}`, string(resStr))
}
