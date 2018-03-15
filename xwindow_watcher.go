package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

var pollInterval int
var postUrl string
var appKey string

func postJson(jsonStr []byte) {
	url := postUrl
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

type PostBody struct {
	Window *Window
	AppKey string
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
	postBody := PostBody{Window: w, AppKey: appKey}
	windowJson, err := json.Marshal(postBody)
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
	windowHistory  []*Window
}

func (wt *WindowTracker) Push(w *Window) {
	wt.windowHistory = append(wt.windowHistory, w)
}

func (wt *WindowTracker) Pop() (w *Window, err error) {
	if len(wt.windowHistory) == 0 {
		return &Window{}, errors.New("Window queue is empty")
	}
	w = wt.windowHistory[0]
	wt.windowHistory = wt.windowHistory[1:]
	return w, nil
}

func (wt *WindowTracker) NewActive(title string) {
	var now = time.Now().UTC()
	var duration = time.Since(*wt.activeWindow.Start_time)
	new_window := &Window{Title: title, Start_time: &now}

	// add queue to log previously seen windows
	old_window, err := wt.Pop()
	if err != nil {
		old_window = nil
	}
	fmt.Println(old_window)
	wt.Push(new_window)

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

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	pollInterval = viper.GetInt("pollInterval")
	postUrl = viper.GetString("postUrl")
	appKey = viper.GetString("appKey")
}

func main() {
	var now = time.Now().UTC()
	title := getAppName() + " || " + getWindowTitle()
	var window = Window{Title: title, Start_time: &now}
	var window_tracker = WindowTracker{activeWindow: &window}
	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)
		title := getAppName() + " || " + getWindowTitle()
		window_tracker.Poll(title)
	}
}
