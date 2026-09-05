package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	gaugeTableName   = "gauges"
	counterTableName = "counters"
)

type gauge struct {
	name  string
	value float64
}

type counter struct {
	name  string
	delta int64
}

// PostgresStorage wraps a sql.DB connection for PostgreSQL operations.
type PostgresStorage struct {
	db *sql.DB
}

// Close closes the underlying database connection.
func (ps *PostgresStorage) Close() error {
	return ps.db.Close()
}

var (
	addCounterExec = fmt.Sprintf(`
    INSERT INTO %s (name, delta)
    VALUES ($1, $2)
    ON CONFLICT (name) DO UPDATE
    SET delta = %s.delta + EXCLUDED.delta
	RETURNING delta;`, counterTableName, counterTableName)

	setGaugeExec = fmt.Sprintf(`
	INSERT INTO %s (name, value) 
	VALUES ($1, $2) 
	ON CONFLICT (name) DO UPDATE 
	SET value = $2;`, gaugeTableName)

	getCounterQuery = fmt.Sprintf(`
	SELECT delta 
	FROM %s
	WHERE NAME = $1;`, counterTableName)

	getGaugeQuery = fmt.Sprintf(`
	SELECT value 
	FROM %s
	WHERE NAME = $1;`, gaugeTableName)

	getAllCountersQuery = fmt.Sprintf(`
	SELECT name, delta 
	FROM %s;`, counterTableName)

	getAllGaugesQuery = fmt.Sprintf(`
	SELECT name, value 
	FROM %s;`, gaugeTableName)
)

// NewPostgresStorage creates a new PostgresStorage instance.
func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		db: db,
	}
}

// NewPostgresStorageFromDSN opens a PostgreSQL connection using the given DSN, runs migrations, and returns a PostgresStorage.
func NewPostgresStorageFromDSN(dsn, migrationsPath string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open standard sql DB: %w", err)
	}

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"pgx",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrator: %w", err)
	}

	log.Println("Applying migrations...")
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No new migrations to apply.")
		} else {
			return nil, fmt.Errorf("migration failed: %w", err)
		}
	} else {
		log.Println("Migrations applied successfully.")
	}

	postgresStorage := NewPostgresStorage(db)
	return postgresStorage, nil
}

// PingContext verifies the database connection is alive using the provided context.
func (ps *PostgresStorage) PingContext(ctx context.Context) error {
	return ps.db.PingContext(ctx)
}

// SetGauge sets the named gauge metric to the specified value, overwriting any previous value.
func (ps *PostgresStorage) SetGauge(name string, value float64) error {
	_, err := ps.db.Exec(setGaugeExec, name, value)

	if err != nil {
		return fmt.Errorf("error writing gauge: %w", err)
	}

	return nil
}

// AddCounter increments the named counter metric by the specified delta.
func (ps *PostgresStorage) AddCounter(name string, value int64) (newDelta int64, err error) {
	err = ps.db.QueryRow(addCounterExec, name, value).Scan(&newDelta)

	if err != nil {
		return 0, fmt.Errorf("error writing counter: %w", err)
	}
	return newDelta, nil
}

// GetGauge returns the value of the named gauge metric.
// Returns an error if the metric does not exist.
func (ps *PostgresStorage) GetGauge(name string) (value float64, err error) {
	row := ps.db.QueryRow(getGaugeQuery, name)
	var result gauge
	err = row.Scan(&result.value)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("gauge %s %w", name, ErrMetricNotFound)
	}

	if err != nil {
		return 0, fmt.Errorf("error scanning row: %w", err)
	}

	return result.value, nil
}

// GetCounter returns the value of the named counter metric.
// Returns an error if the metric does not exist.
func (ps *PostgresStorage) GetCounter(name string) (value int64, err error) {
	row := ps.db.QueryRow(getCounterQuery, name)
	var result counter
	err = row.Scan(&result.delta)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("counter %s %w", name, ErrMetricNotFound)
	}

	if err != nil {
		return 0, fmt.Errorf("error scanning row: %w", err)
	}

	return result.delta, nil
}

// GetAllGauges returns a shallow copy of all gauge metrics to prevent external mutation.
func (ps *PostgresStorage) GetAllGauges() (result map[string]float64, err error) {
	rows, err := ps.db.Query(getAllGaugesQuery)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	result = make(map[string]float64)

	for rows.Next() {
		var g gauge
		err = rows.Scan(&g.name, &g.value)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		result[g.name] = g.value
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error encountered during iteration: %w", err)
	}

	return result, nil
}

// GetAllCounters returns a shallow copy of all counter metrics to prevent external mutation.
func (ps *PostgresStorage) GetAllCounters() (result map[string]int64, err error) {
	rows, err := ps.db.Query(getAllCountersQuery)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	result = make(map[string]int64)

	for rows.Next() {
		var c counter
		err = rows.Scan(&c.name, &c.delta)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		result[c.name] = c.delta
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error encountered during iteration: %w", err)
	}

	return result, nil
}

func (ps *PostgresStorage) UpdateMetrics(ctx context.Context, metrics []model.Metric) (err error) {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning a transaction: %w", err)
	}
	defer tx.Rollback()

	for _, metric := range metrics {
		switch metric.Type {
		case model.Counter:
			_, err = tx.ExecContext(ctx, addCounterExec, metric.ID, *metric.Delta)
			if err != nil {
				return fmt.Errorf("error executing addCounter: %w", err)
			}
		case model.Gauge:
			_, err = tx.ExecContext(ctx, setGaugeExec, metric.ID, *metric.Value)
			if err != nil {
				return fmt.Errorf("error executing setGauge: %w", err)
			}
		default:
			log.Printf("unknown metric type: %s\n", metric.Type)
		}

	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing a transaction: %w", err)
	}

	return nil
}
