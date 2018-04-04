package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Activity struct {
	Id              int
	Name            string
	Started_time    time.Time
	End_time        time.Time
	Duration        time.Duration
	DurationSeconds int64
	App_id          int
	AppName         string
}

type App struct {
	Id         int
	Name       string
	Created_at time.Time
}

type ActivityList struct {
	Activities []*Activity
}

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

type Page struct{}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
	// p := Page{}
	p := get_last_n_events(10000)
	t.Execute(w, p)
}

func get_last_n_events(n int) (activity_list *ActivityList) {
	// rows, err := db.Query("SELECT * FROM activity JOIN apps ON apps.id = activity.app_id LIMIT ?", n)
	rows, err := db.Query("SELECT activity.id, activity.name, activity.started_time, activity.end_time, activity.duration, activity.app_id, apps.name FROM activity JOIN apps ON apps.id = activity.app_id WHERE duration IS NOT NULL AND date(activity.started_time) = date('now', '-1 day') ORDER BY activity.started_time DESC LIMIT ?", n)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	activity_list = &ActivityList{}
	for rows.Next() {
		var activity = Activity{}
		rows.Scan(&activity.Id, &activity.Name, &activity.Started_time, &activity.End_time, &activity.Duration, &activity.App_id, &activity.AppName)
		fmt.Println(activity)
		activity.DurationSeconds, _ = strconv.ParseInt(strconv.FormatFloat(activity.Duration.Seconds(), 'f', 0, 64), 10, 64)
		activity_list.Activities = append(activity_list.Activities, &activity)
	}
	return activity_list
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
	http.HandleFunc("/", index)

	log.Fatal(http.ListenAndServe(":5000", nil))

}
