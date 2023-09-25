package main

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	"log"
)

type Store interface {
	CreateDatabase() error
	GetPrediction(string, string, string) (*[]PredictionDto, error)
	InsertPrediction(PredictionTableEntry) (int, error)
	InsertPredictionValue(PredictionValueTableEntry, int) error
	InsertPredictionModel(PredictionModelTableEntry, int) (int, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	//connStr := "host=db port=5432 user=postgres dbname=predictions password=pass sslmode=disable"
	connStr := "user=postgres dbname=predictions password=pass sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) CreateDatabase() error {
	queries := []string{

		`CREATE TABLE IF NOT EXISTS predictions.Prediction (
			id serial PRIMARY KEY,
			wf_name varchar(20) NOT NULL,
			prediction_date timestamptz NOT NULL,
			prediction_from timestamptz NOT NULL,
			prediction_to timestamptz NOT NULL
		);

		CREATE TABLE IF NOT EXISTS predictions.Model (
			id serial PRIMARY KEY,
			prediction_id int, 
			FOREIGN KEY (prediction_id) REFERENCES predictions.Prediction(id),
			model_uuid varchar(50),
			wtg_code varchar(20) NOT NULL,
			CONSTRAINT unique_prediction_model_wtg UNIQUE (prediction_id, model_uuid, wtg_code)
		);

		CREATE TABLE IF NOT EXISTS predictions.PredictionValue (
			id serial PRIMARY KEY,
			model_id int,
			FOREIGN KEY (model_id) REFERENCES predictions.Model(id),
			prediction_for timestamptz NOT NULL,
			value int
		);`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) InsertPrediction(entry PredictionTableEntry) (int, error) {
	insertSQL := `
		INSERT INTO predictions.Prediction (wf_name, prediction_date, prediction_from, prediction_to)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	var id int
	err := s.db.QueryRow(insertSQL, entry.WFName, entry.PredictionDate, entry.PredictionPeriodFrom, entry.PredictionPeriodTo).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (s *PostgresStore) InsertPredictionValue(entry PredictionValueTableEntry, modelId int) error {
	insertSQL := `
		INSERT INTO predictions.PredictionValue (model_id, prediction_for, value)
		VALUES ($1, $2, $3)`

	_, err := s.db.Exec(insertSQL, modelId, entry.PredictionFor, entry.Value)
	if err != nil {
		log.Printf("Error inserting prediction value: %v", err)
		return err
	}

	return nil
}

func (s *PostgresStore) InsertPredictionModel(entry PredictionModelTableEntry, predictionId int) (int, error) {
	insertSQL := `
		WITH inserted_row AS (
    INSERT INTO predictions.Model (prediction_id, model_uuid, wtg_code)
    VALUES ($1, $2, $3)
    ON CONFLICT (prediction_id, model_uuid, wtg_code) DO NOTHING
    RETURNING id
	)
	SELECT id FROM inserted_row
	UNION ALL
	SELECT id FROM predictions.Model WHERE prediction_id = $1 AND model_uuid = $2 AND wtg_code = $3
	LIMIT 1;`

	var id int
	err := s.db.QueryRow(insertSQL, predictionId, entry.ModelUUID, entry.WTGCode).Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error inserting prediction model: %v", err)
		return -1, err
	}

	return id, nil
}

func (s *PostgresStore) GetPrediction(wfName string, timestamp string, wtCodes string) (*[]PredictionDto, error) {
	query := `
		SELECT
			p.wf_name AS "wfName",
			p.prediction_date AS "predictionDate",
			p.prediction_from AS "predictionPeriod.from",
			p.prediction_to AS "predictionPeriod.to",
			m.wtg_code AS "predictionValues.wtgCode",
			pv.prediction_for AS "predictionValues.predictions.predictionFor",
			pv.value AS "predictionValues.predictions.predictedValue_kWh"
		FROM
			predictions.prediction AS p
		JOIN
			predictions.model AS m ON p.id = m.prediction_id
		JOIN
			predictions.predictionvalue AS pv ON m.id = pv.model_id
		WHERE
			p.wf_name = $1
			AND prediction_date = $2
			--AND wtg_code IN (wtCodes)`

	rows, err := s.db.Query(query, wfName, timestamp)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	var predictions []PredictionDto

	for rows.Next() {
		prediction := PredictionDto{}

		err := rows.Scan(
			&prediction.WfName,
			&prediction.PredictionDate,
			&prediction.From,
			&prediction.To,
			&prediction.WtgCode,
			&prediction.PredictionFor,
			&prediction.ValueKwh,
		)

		if err != nil {
			return nil, err
		}
		predictions = append(predictions, prediction)
	}

	return &predictions, nil
}
