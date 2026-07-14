package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", handleWS)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	addr := ":8080"
	fmt.Printf("SCP visualizer server running on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("webstock upgrade error: %v", err)
		return
	}

	defer conn.Close()

	log.Println("client connected, starting simulation muehehehehe")

	emit := func(step VisualizerStep) {
		data, err := json.Marshal(step)
		if err != nil {
			log.Printf("marshal error: %v", err)
		}

		if err := conn.WriteMessage(1, data); err != nil {
			log.Printf("write error: %v", err)
		}
	}

	RunSimulation(emit)

	done, _ := json.Marshal(map[string]any{
		"type": "simulation_done",
	})
	conn.WriteMessage(1, done)

	log.Println("simulation complete")
}
