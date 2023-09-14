package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

func parse() (*PredictionResponse, error) {
	var predictionResponse PredictionResponse

	path := "predictions.json"

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = json.Unmarshal(jsonData, &predictionResponse)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &predictionResponse, nil
}

func (response *PredictionResponse) mapToPredictionTableEntries() PredictionTableEntry {
	wfName := response.Prediction.WFName
	predictionDate := response.Prediction.PredictionDate
	predictionFrom := response.Prediction.PredictionPeriod.From
	predictionTo := response.Prediction.PredictionPeriod.To

	return PredictionTableEntry{
		WFName:               wfName,
		PredictionDate:       predictionDate,
		PredictionPeriodFrom: predictionFrom,
		PredictionPeriodTo:   predictionTo,
	}
}

func (response *PredictionResponse) mapToPredictionValueAndModelEntries() []ModelValuePair {
	var modelValuePairs []ModelValuePair

	for _, val := range response.Prediction.PredictionValues {
		model := PredictionModelTableEntry{ModelUUID: val.ModelID, WTGCode: val.WtgCode}

		for _, prediction := range val.Predictions {
			value := PredictionValueTableEntry{PredictionFor: prediction.PredictionFor, Value: int(prediction.ValueKwh)}

			pair := ModelValuePair{Value: value, Model: model}
			modelValuePairs = append(modelValuePairs, pair)
		}
	}
	return modelValuePairs
}
