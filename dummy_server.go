package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html"
	"io"
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
	// buf := new(bytes.Buffer)
	// buf.ReadFrom(r.Body)
	// s := buf.String()
	// fmt.Println(s)

	// var dat map[string]interface{}
	dec := json.NewDecoder(r.Body)
	for {
		var m PostBody
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", m.Window.Title, m.AppKey)
		app_id := get_or_create_app(m.Window.AppName)

		insert_activity(m.Window, app_id)
	}
}

func insert_activity(w *Window, app_id int64) {
	_, err := db.Exec("INSERT INTO activity (name, started_time, end_time, duration, app_id) VALUES ($1, $2, $3, $4, $5)",
		w.Title, w.Start_time, w.End_time, w.Duration, app_id)
	if err != nil {
		log.Fatal(err)
	}
}

func get_or_create_app(appName string) (id int64) {
	row := db.QueryRow("SELECT id FROM apps WHERE name = $1", appName)
	err := row.Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	} else if err == sql.ErrNoRows {
		result, err := db.Exec("INSERT INTO apps (name) VALUES ($1)", appName)
		if err != nil {
			log.Fatal(err)
		}
		id, _ = result.LastInsertId()
	}
	return id
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./x.db")
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()

	fmt.Println("waiting...")
	http.HandleFunc("/v1/event/add", printRequest)

	log.Fatal(http.ListenAndServe(":5000", nil))

}
