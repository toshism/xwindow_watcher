package main

import (
	"bytes"
	"encoding/json"
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

type Window struct {
	Title      string         `json:"title"`
	AppName    string         `json:"app_name"`
	Start_time *time.Time     `json:"start_time"`
	End_time   *time.Time     `json:"end_time"`
	Duration   *time.Duration `json:"duration"`
}

type WindowTracker struct {
	activeWindow   *Window
	previousWindow *Window
}

type PostBody struct {
	Window *Window
	AppKey string
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

func postJson(jsonStr []byte) {
	url := postUrl

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func (wt *WindowTracker) NewActive(title string, AppName string) {
	var now = time.Now().UTC()
	var duration = time.Since(*wt.activeWindow.Start_time)
	new_window := &Window{Title: title, Start_time: &now, AppName: AppName}

	wt.previousWindow = wt.activeWindow
	wt.previousWindow.End_time = &now
	wt.previousWindow.Duration = &duration
	wt.activeWindow = new_window
}

func (wt *WindowTracker) Poll(title string, AppName string) {
	if title != wt.activeWindow.Title {
		wt.NewActive(title, AppName)
		fmt.Println(wt.activeWindow.Title)
		window_json := wt.previousWindow.ToJson()
		go postJson(window_json)
	}
}

func runCmd(cmd *exec.Cmd) string {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("stdout error : %v", err)
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("start error : %v", err)
	}
	b, _ := ioutil.ReadAll(stdout)

	if err := cmd.Wait(); err != nil {
		fmt.Printf("wait error : %v", err)
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
		title := getWindowTitle()
		AppName := getAppName()
		window_tracker.Poll(title, AppName)
	}
}
