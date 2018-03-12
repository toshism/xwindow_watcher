package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

var d, _ = time.ParseDuration("10s")
var activeThreshold = d.Seconds()
var p, _ = time.ParseDuration("10s")
var pollInterval = p.Seconds()

func postJson(jsonStr []byte) {
	url := "http://127.0.0.1:5000/v1/event/add"
	fmt.Println("URL:>", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// panic(err)
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

type Window struct {
	Title      string         `json:"title"`
	Start_time *time.Time     `json:"start_time"`
	End_time   *time.Time     `json:"end_time"`
	Duration   *time.Duration `json:"duration"`
}

func (w Window) runningDuration() (duration *time.Duration) {
	if w.Duration != nil {
		return w.Duration
	} else {
		t := time.Since(*w.Start_time)
		return &t
	}
}

func (w *Window) ToJson() (windowJson []byte) {
	windowJson, err := json.Marshal(*w)
	if err != nil {
		log.Fatal(err)
		return
	}
	return windowJson
}

type WindowTracker struct {
	activeWindow   *Window
	previousWindow *Window
	currentWindow  *Window
}

func (wt *WindowTracker) NewActive(title string) {
	var now = time.Now().UTC()
	var duration = time.Since(*wt.activeWindow.Start_time)
	new_window := &Window{Title: title, Start_time: &now}

	wt.previousWindow = wt.activeWindow
	wt.previousWindow.End_time = &now
	wt.previousWindow.Duration = &duration
	wt.activeWindow = new_window
}

func (wt *WindowTracker) Poll(title string) {
	if title != wt.activeWindow.Title {
		wt.NewActive(title)
		fmt.Println(wt.activeWindow.Title)
		window_json := wt.previousWindow.ToJson()
		postJson(window_json)
	}
}

func runCmd(cmd *exec.Cmd) string {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	b, _ := ioutil.ReadAll(stdout)

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	return strings.TrimSuffix(string(b), "\n")
}

func getProcessNameFromPID(pid string) (app_name string) {
	proc_string := "/proc/" + pid + "/comm"
	cmd := exec.Command("cat", proc_string)
	return runCmd(cmd)
}

func getActivePid() (pid string) {
	cmd := exec.Command("/usr/bin/xdotool", "getactivewindow", "getwindowpid")
	return runCmd(cmd)
}

func getAppName() string {
	return getProcessNameFromPID(getActivePid())
}

func getWindowTitle() (title string) {
	cmd := exec.Command("/usr/bin/xdotool", "getactivewindow", "getwindowname")
	return runCmd(cmd)
}

func main() {
	var now = time.Now().UTC()
	title := getAppName() + " || " + getWindowTitle()
	var window = Window{Title: title, Start_time: &now}
	var window_tracker = WindowTracker{activeWindow: &window}
	for {
		time.Sleep(1000 * time.Millisecond)
		title := getAppName() + " || " + getWindowTitle()
		window_tracker.Poll(title)
	}
}
