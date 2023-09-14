package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {

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

func (s *PostgresStore) createDatabase() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS predictions.Prediction (
			id serial PRIMARY KEY,
			wf_name varchar(20) NOT NULL,
			prediction_date timestamptz NOT NULL,
			prediction_from timestamptz NOT NULL,
			prediction_to timestamptz NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS predictions.PredictionValue (
			id serial PRIMARY KEY,
			prediction_id int, 
			FOREIGN KEY(prediction_id) REFERENCES predictions.Prediction(id), -- Reference Prediction table
			prediction_for timestamptz NOT NULL,
			value int
		);`,
		`CREATE TABLE IF NOT EXISTS predictions.Model (
			id serial PRIMARY KEY,
			prediction_value_id int,
			FOREIGN KEY(prediction_value_id) REFERENCES predictions.PredictionValue(id), -- Reference PredictionValue table
			model_uuid varchar(50),
			wtg_code varchar(20) NOT NULL
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

func (s *PostgresStore) InsertPredictionValue(entry PredictionValueTableEntry, predictionId int) (int, error) {
	insertSQL := `
		INSERT INTO predictions.PredictionValue (prediction_id, prediction_for, value)
		VALUES ($1, $2, $3)
		RETURNING id`

	var id int
	err := s.db.QueryRow(insertSQL, predictionId, entry.PredictionFor, entry.Value).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (s *PostgresStore) InsertPredictionModel(entry PredictionModelTableEntry, predictionValueId int) error {
	insertSQL := `
		INSERT INTO predictions.Model (prediction_value_id, model_uuid, wtg_code)
		VALUES ($1, $2, $3)
		RETURNING id`

	_, err := s.db.Exec(insertSQL, predictionValueId, entry.ModelUUID, entry.WTGCode)
	if err != nil {
		return err
	}
	return nil
}
