package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yryz/ds18b20"
)

type sensorResponse struct {
	Sensors map[string]string `json:"sensors"`
}

func getTemperatures(w http.ResponseWriter, r *http.Request) {
	sensors, err := ds18b20.Sensors()
	if err != nil {
		panic(err);
	}

	fmt.Printf("sensor IDs: %v\n", sensors)

	response := sensorResponse{}
	response.Sensors = make(map[string]string)

	for _, sensor := range sensors {
		t, err := ds18b20.Temperature(sensor)
		if err == nil {
			fmt.Printf("sensor: %s temperature: %.2f F\n", sensor, convertCToF(t))
			response.Sensors[sensor] = fmt.Sprintf("%.2f", convertCToF(t))
		} else {
			panic(err)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/temperatures/all", getTemperatures)
	log.Fatal(http.ListenAndServe(":8002", router))
}

func convertCToF(celsiusValue float64) float64 {
	return (celsiusValue * 9.0 / 5.0) + 32
}
