package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	listenAddr string
	store      Store
}

func NewServer(listenAddr string, store Store) *Server {
	return &Server{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *Server) Run() {

	router := mux.NewRouter()

	router.HandleFunc("/predictions/{wfName}", makeHTTPHandleFunc(s.handlePredictions))
	router.HandleFunc("/persist", makeHTTPHandleFunc(s.handlePersist))
	router.HandleFunc("/persist/rabbit", makeHTTPHandleFunc(s.handlePersistRabbit))
	router.HandleFunc("/get/rabbit", makeHTTPHandleFunc(s.handleGetRabbit))

	log.Println("JSON API server running on port", s.listenAddr)
	err := http.ListenAndServe(s.listenAddr, router)
	if err != nil {
		log.Fatal(err)
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			err := WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			if err != nil {
				return
			}
		}
	}
}

func (s *Server) handlePredictions(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("method not allowed")
	}

	vars := mux.Vars(r)
	wfName := vars["wfName"]
	wtCodes := r.URL.Query().Get("wtCodes")
	timestamp := r.URL.Query().Get("timestamp")
	timestamp = strings.Replace(timestamp, "Z", "+", -1)
	prediction, err := s.store.GetPrediction(wfName, timestamp, wtCodes)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, prediction)
}

func (s *Server) handlePersist(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed")
	}

	var predictionResponse PredictionJson
	err := json.NewDecoder(r.Body).Decode(&predictionResponse)
	if err != nil {
		http.Error(w, "Failed to parse JSON data", http.StatusBadRequest)
		return err
	}

	predictionId, err := s.store.InsertPrediction(predictionResponse.mapToPredictionTableEntries())

	if err != nil {
		log.Fatal(err)
	}

	pairs := predictionResponse.mapToPredictionValueAndModelEntries()

	for _, pair := range pairs {
		predictionModelId, err := s.store.InsertPredictionModel(pair.Model, predictionId)
		if err != nil {
			log.Printf("Error inserting prediction model: %v", err)
			return err
		}
		err = s.store.InsertPredictionValue(pair.Value, predictionModelId)
		if err != nil {
			log.Printf("Error inserting prediction value: %v", err)
			return err
		}
	}

	return WriteJSON(w, http.StatusOK, "data successfully persisted")

}

func (s *Server) handlePersistRabbit(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed")
	}

	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	RabbitSend(message)
	return nil
}

func (s *Server) handleGetRabbit(w http.ResponseWriter, r *http.Request) error {
	RabbitRecieve()
	return nil
}
