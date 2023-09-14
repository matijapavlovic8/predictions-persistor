package main

import (
	"log"
)

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}
	if err := store.createDatabase(); err != nil {
		log.Fatal(err)
	}

	prediction, err := parse()

	predictionId, err := store.InsertPrediction(prediction.mapToPredictionTableEntries())

	if err != nil {
		log.Fatal(err)
	}

	pairs := prediction.mapToPredictionValueAndModelEntries()

	for _, pair := range pairs {
		if err != nil {
			log.Fatal(err)
		}
		predictionValueId, err := store.InsertPredictionValue(pair.Value, int(predictionId))
		if err != nil {
			log.Fatal(err)
		}

		err = store.InsertPredictionModel(pair.Model, int(predictionValueId))
		if err != nil {
			log.Fatal(err)
		}
	}

}
