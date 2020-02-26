package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/robfig/cron"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
)

type State int

const (
	Closed State = iota + 1
	Opening
	Open
	Closing
	Stuck
)

var STUCK_DOOR_SECONDS int
var DISCORD_WEBHOOK_URL string
var AUTO_OPEN_CRON string
var AUTO_CLOSE_CRON string
const LOWER_HALL_EFFECT_SENSOR = 0
const UPPER_HALL_EFFECT_SENSOR = 1

var motorPins *gpiod.Lines
var hallEffectPins *gpiod.Lines

var currentState State
var stateStartTime *time.Time

func main() {
	pullConfigFromEnvironmentVariables()

	chip, err := gpiod.NewChip("gpiochip0")
	die(err)
	defer chip.Close()

	motorPins, err = chip.RequestLines([]int{rpi.GPIO19, rpi.GPIO26}, gpiod.AsOutput(0, 0))
	die(err)
	hallEffectPins, err = chip.RequestLines([]int{rpi.GPIO6, rpi.GPIO13}, gpiod.AsInput)
	die(err)

	defer func() {
		motorPins.Reconfigure(gpiod.AsInput)
		motorPins.Close()
		hallEffectPins.Close()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	currentState = determineInitialDoorState()
	stateStartTime = nil

	go doControlLoop()
	registerCronJobs()

	http.HandleFunc("/status", getStatusHandler)
	http.HandleFunc("/open", openDoorHandler)
	http.HandleFunc("/close", closeDoorHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

	<-quit
}

func pullConfigFromEnvironmentVariables() {
	DISCORD_WEBHOOK_URL = os.Getenv("DISCORD_WEBHOOK_URL")
	AUTO_OPEN_CRON = os.Getenv("AUTO_OPEN_CRON")
	AUTO_CLOSE_CRON = os.Getenv("AUTO_CLOSE_CRON")
	var err error
	STUCK_DOOR_SECONDS, err = strconv.Atoi(os.Getenv("STUCK_DOOR_SECONDS"))
	die(err)
}

func determineInitialDoorState() State {
	if isDoorOpen() {
		return Open
	} else if isDoorClosed() {
		return Closed
	}

	return Stuck
}

func doControlLoop() {
	for {
		time.Sleep(10 * time.Millisecond)

		switch currentState {
		case Closing:
			if isDoorClosed() {
				currentState = Closed
				log.Println("Door is fully closed.")
				sendDiscordMessage("Door is fully closed.")
				stateStartTime = nil
				stopDoor()
			} else if secondsInState() > STUCK_DOOR_SECONDS {
				handleStuckDoor()
			} else {
				lowerDoor()
			}
		case Opening:
			if isDoorOpen() {
				currentState = Open
				log.Println("Door is fully open.")
				sendDiscordMessage("Door is fully open.")
				stopDoor()
			} else if secondsInState() > STUCK_DOOR_SECONDS {
				handleStuckDoor()
			} else {
				raiseDoor()
			}
		}
	}
}

func isDoorOpen() bool {
	return isDoorInPosition(UPPER_HALL_EFFECT_SENSOR)
}

func isDoorClosed() bool {
	return isDoorInPosition(LOWER_HALL_EFFECT_SENSOR)
}

func isDoorInPosition(sensorIndex int) bool {
	hallEffectValues := []int{0, 0}
	hallEffectPins.Values(hallEffectValues)
	return hallEffectValues[sensorIndex] == 0
}

func handleStuckDoor() {
	currentState = Stuck
	sendDiscordMessage("Door is stuck!")
	log.Println("STUCK")
	stateStartTime = nil
	stopDoor()
}

func raiseDoor() {
	motorPins.SetValues([]int{0, 1})
}

func lowerDoor() {
	motorPins.SetValues([]int{1, 0})
}

func stopDoor() {
	motorPins.SetValues([]int{0, 0})
}

func registerCronJobs() {
	c := cron.New()

	if AUTO_OPEN_CRON != "" {
		c.AddFunc(AUTO_OPEN_CRON, func() { triggerRaiseDoor() })
		fmt.Println("Added handler for auto-open cron")
	}

	if AUTO_CLOSE_CRON != "" {
		c.AddFunc(AUTO_CLOSE_CRON, func() { triggerLowerDoor() })
		fmt.Println("Added handler for auto-close cron")
	}

	c.Start()
}

func openDoorHandler(w http.ResponseWriter, r *http.Request) {
	triggerRaiseDoor()

	fmt.Fprintf(w, "Requesting door open.")
}

func closeDoorHandler(w http.ResponseWriter, r *http.Request) {
	triggerLowerDoor()

	fmt.Fprintf(w, "Requesting door close.")
}

func triggerLowerDoor() {
	log.Println("Lowering door")
	currentState = Closing
	stateStartTime = getCurrentTime()
}

func triggerRaiseDoor() {
	log.Println("Raising door")
	currentState = Opening
	stateStartTime = getCurrentTime()
}

func getCurrentTime() *time.Time {
	time := time.Now()
	return &time
}

func getStatusHandler(w http.ResponseWriter, r *http.Request) {
	statusDisplay := ""

	switch currentState {
	case Closed:
		statusDisplay = "Closed"
	case Open:
		statusDisplay = "Open"
	case Closing:
		statusDisplay = fmt.Sprintf("Closing (%d seconds)", secondsInState())
	case Opening:
		statusDisplay = fmt.Sprintf("Opening (%d seconds)", secondsInState())
	case Stuck:
		statusDisplay = "Stuck"
	}

	fmt.Fprintf(w, statusDisplay)
}

func secondsInState() int {
	return int(time.Now().Sub(*stateStartTime).Seconds())
}

func sendDiscordMessage(message string) {
	if DISCORD_WEBHOOK_URL == "" {
		return
	}

	requestBody, err := json.Marshal(map[string]string{
		"content": message,
	})
	die(err)

	resp, err := http.Post(DISCORD_WEBHOOK_URL, "application/json", bytes.NewBuffer(requestBody))
	die(err)

	defer resp.Body.Close()
}

func die(err error) {
	if err != nil {
		panic(err)
	}
}
