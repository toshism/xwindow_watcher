package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"time"
)

type PostBody struct {
	Window *Window
	AppKey string
}

type Window struct {
	Title      string         `json:"title"`
	AppName    string         `json:"app_name"`
	Start_time *time.Time     `json:"start_time"`
	End_time   *time.Time     `json:"end_time"`
	Duration   *time.Duration `json:"duration"`
}

func structRequest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t PostBody
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	log.Println(t.Window)
}

func printRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	s := buf.String()
	fmt.Println(s)
}

func main() {
	fmt.Println("waiting...")
	http.HandleFunc("/v1/event/add", printRequest)

	log.Fatal(http.ListenAndServe(":5000", nil))
}
