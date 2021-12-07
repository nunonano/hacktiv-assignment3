package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/nunonano/hacktiv-assignment3/model"
)

const (
	PAGE_REFRESH_SECOND int    = 15
	UPDATE_FILE_SECOND  int    = 5
	JSON_PATH           string = "json/data.json"
)

var (
	statusMu sync.RWMutex
	status   model.Status
)

func main() {
	fmt.Println("==============Start================")

	fmt.Println("Config:")
	fmt.Printf("Page Refresh every %v second\n", PAGE_REFRESH_SECOND)
	fmt.Printf("Update File every %v second\n", UPDATE_FILE_SECOND)
	fmt.Println("===================================")

	// Start the goroutine that will sum the current time
	go runDataLoop()

	// Create a handler that will read-lock the mutext and
	// write the summed time to the client
	tmpl := template.Must(template.ParseFiles("view/page.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		statusMu.RLock()
		defer statusMu.RUnlock()

		log.Println("=========Read Data Start==========")
		// Open jsonFile
		jsonFile, err := os.Open("json/data.json")

		if err != nil {
			fmt.Println("error:", err.Error())
		}
		log.Println("Successfully Opened data.json")

		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		// read our opened jsonFile as a byte array.
		byteValue, _ := ioutil.ReadAll(jsonFile)

		data := struct {
			PageRefreshSecond int
			PageTitle         string
			Status            model.Status `json:"status"`
			Message           struct {
				Water string
				Wind  string
			}
		}{
			PageRefreshSecond: PAGE_REFRESH_SECOND,
			PageTitle:         "Water and Wind Status",
			Status:            status,
			Message: struct {
				Water string
				Wind  string
			}{Water: getWaterMessage(status.Water), Wind: getWindMessage(status.Wind)},
		}

		err = json.Unmarshal(byteValue, &data)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
		log.Println("ResponseWriter:", data)
		tmpl.Execute(w, data)
		log.Println("=========Read Data End==========")
	})

	// http://127.0.0.1:8080/
	http.ListenAndServe(":8080", nil)
}

// Within an infinite loop, lock the mutex and
// increment our value, then sleep for 1 second until
// the next time we need to get a value.
func runDataLoop() {
	for {
		time.Sleep(time.Duration(UPDATE_FILE_SECOND) * time.Second)

		statusMu.Lock()
		status = model.Status{
			Water: setRandomNumber(1, 10),
			Wind:  setRandomNumber(1, 20)}

		newJson, err := json.Marshal(struct {
			Status model.Status `json:"status"`
		}{Status: status})
		if err != nil {
			fmt.Println("error:", err.Error())
		}

		err = ioutil.WriteFile(JSON_PATH, newJson, 0644)
		if err != nil {
			fmt.Println("error:", err.Error())
		} else {
			fmt.Println("Successfully Updated data.json")
			fmt.Println(string(newJson))
		}
		statusMu.Unlock()
	}
}

func setRandomNumber(min int, max int) int {
	return min + rand.Intn(max-min+1)
}

func getWaterMessage(v int) (msg string) {
	switch {
	case v <= 5:
		msg = "Safe"
	case v <= 8:
		msg = "High Alert"
	case (v > 8) && (v <= 10):
		msg = "Danger"
	default:
		msg = "Unknown"
	}
	return
}

func getWindMessage(v int) (msg string) {
	switch {
	case v <= 6:
		msg = "Safe"
	case v <= 15:
		msg = "High Alert"
	case (v > 15) && (v <= 20):
		msg = "Danger"
	default:
		msg = "Unknown"
	}
	return
}
