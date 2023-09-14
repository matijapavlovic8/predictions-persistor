package main

type PredictionResponse struct {
	Prediction struct {
		WFName           string `json:"wfName"`
		PredictionDate   string `json:"predictionDate"`
		PredictionPeriod struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		PredictionValues []struct {
			WtgCode     string `json:"wtgCode"`
			ModelID     string `json:"modelUuid"`
			Predictions []struct {
				PredictionFor string `json:"predictionFor"`
				ValueKwh      int    `json:"predictedValue_kWh"`
			} `json:"predictions"`
		} `json:"predictionValues"`
	} `json:"prediction"`
}

type PredictionTableEntry struct {
	WFName               string
	PredictionDate       string
	PredictionPeriodFrom string
	PredictionPeriodTo   string
}

type PredictionValueTableEntry struct {
	PredictionFor string
	Value         int
}

type PredictionModelTableEntry struct {
	ModelUUID string
	WTGCode   string
}

type ModelValuePair struct {
	Model PredictionModelTableEntry
	Value PredictionValueTableEntry
}
