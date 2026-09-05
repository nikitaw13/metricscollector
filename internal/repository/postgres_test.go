package repository

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Query patterns match the SQL statements built in postgres.go. sqlmock matches
// by regular expression, so \s+ absorbs the raw-string indentation of queries.
const (
	qInsertGauge       = `INSERT\s+INTO\s+gauges`
	qInsertCounter     = `INSERT\s+INTO\s+counters`
	qSelectGauge       = `SELECT\s+value\s+FROM\s+gauges`
	qSelectCounter     = `SELECT\s+delta\s+FROM\s+counters`
	qSelectAllGauges   = `SELECT\s+name,\s+value\s+FROM\s+gauges`
	qSelectAllCounters = `SELECT\s+name,\s+delta\s+FROM\s+counters`
)

// newMockStorage returns a PostgresStorage backed by a sqlmock driver instead
// of a real database. All sqlmock expectations are verified on cleanup.
func newMockStorage(t *testing.T) (*PostgresStorage, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled sqlmock expectations")
		_ = db.Close()
	})

	return NewPostgresStorage(db), mock
}

func TestPostgresStorage_SetGauge(t *testing.T) {
	t.Parallel()

	dbFailure := errors.New("connection refused")

	tests := []struct {
		name    string
		execErr error
	}{
		{name: "upsert succeeds"},
		{name: "database error is wrapped", execErr: dbFailure},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			exec := mock.ExpectExec(qInsertGauge).
				WithArgs("test_gauge", 42.5)
			if tc.execErr != nil {
				exec.WillReturnError(tc.execErr)
			} else {
				exec.WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err := ps.SetGauge("test_gauge", 42.5)

			if tc.execErr != nil {
				assert.ErrorIs(t, err, tc.execErr)
				assert.ErrorContains(t, err, "error writing gauge")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostgresStorage_AddCounter(t *testing.T) {
	t.Parallel()

	dbFailure := errors.New("connection refused")

	tests := []struct {
		name    string
		execErr error
	}{
		{name: "increment upsert succeeds"},
		{name: "database error is wrapped", execErr: dbFailure},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			// AddCounter uses QueryRow (INSERT ... RETURNING delta),
			// so the mock must expect a query, not an exec.
			query := mock.ExpectQuery(qInsertCounter).
				WithArgs("test_counter", int64(5))
			if tc.execErr != nil {
				query.WillReturnError(tc.execErr)
			} else {
				query.WillReturnRows(sqlmock.NewRows([]string{"delta"}).AddRow(int64(10)))
			}

			newDelta, err := ps.AddCounter("test_counter", 5)

			if tc.execErr != nil {
				assert.ErrorIs(t, err, tc.execErr)
				assert.ErrorContains(t, err, "error writing counter")
				assert.Zero(t, newDelta)
			} else {
				require.NoError(t, err)
				assert.Equal(t, int64(10), newDelta, "RETURNING delta must be propagated")
			}
		})
	}
}

func TestPostgresStorage_GetGauge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rows       *sqlmock.Rows
		want       float64
		wantErr    error  // matched with errors.Is
		wantErrMsg string // substring of the wrapped error message
	}{
		{
			name: "existing gauge returns value",
			rows: sqlmock.NewRows([]string{"value"}).AddRow(42.5),
			want: 42.5,
		},
		{
			name:    "missing gauge maps sql.ErrNoRows to model.ErrMetricNotFound",
			rows:    sqlmock.NewRows([]string{"value"}),
			wantErr: model.ErrMetricNotFound,
		},
		{
			name:       "scan failure is wrapped",
			rows:       sqlmock.NewRows([]string{"value"}).AddRow("not-a-number"),
			wantErrMsg: "error scanning row",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			mock.ExpectQuery(qSelectGauge).
				WithArgs("test_gauge").
				WillReturnRows(tc.rows)

			value, err := ps.GetGauge("test_gauge")

			switch {
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Zero(t, value)
			case tc.wantErrMsg != "":
				assert.ErrorContains(t, err, tc.wantErrMsg)
				assert.Zero(t, value)
			default:
				require.NoError(t, err)
				assert.Equal(t, tc.want, value)
			}
		})
	}
}

func TestPostgresStorage_GetCounter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rows       *sqlmock.Rows
		want       int64
		wantErr    error  // matched with errors.Is
		wantErrMsg string // substring of the wrapped error message
	}{
		{
			name: "existing counter returns delta",
			rows: sqlmock.NewRows([]string{"delta"}).AddRow(int64(7)),
			want: 7,
		},
		{
			name:    "missing counter maps sql.ErrNoRows to model.ErrMetricNotFound",
			rows:    sqlmock.NewRows([]string{"delta"}),
			wantErr: model.ErrMetricNotFound,
		},
		{
			name:       "scan failure is wrapped",
			rows:       sqlmock.NewRows([]string{"delta"}).AddRow("not-a-number"),
			wantErrMsg: "error scanning row",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			mock.ExpectQuery(qSelectCounter).
				WithArgs("test_counter").
				WillReturnRows(tc.rows)

			delta, err := ps.GetCounter("test_counter")

			switch {
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Zero(t, delta)
			case tc.wantErrMsg != "":
				assert.ErrorContains(t, err, tc.wantErrMsg)
				assert.Zero(t, delta)
			default:
				require.NoError(t, err)
				assert.Equal(t, tc.want, delta)
			}
		})
	}
}

func TestPostgresStorage_GetAllGauges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rows *sqlmock.Rows
		want map[string]float64
	}{
		{
			name: "multiple gauges",
			rows: sqlmock.NewRows([]string{"name", "value"}).
				AddRow("g1", 1.5).
				AddRow("g2", -2.25),
			want: map[string]float64{"g1": 1.5, "g2": -2.25},
		},
		{
			name: "empty table returns empty map",
			rows: sqlmock.NewRows([]string{"name", "value"}),
			want: map[string]float64{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			mock.ExpectQuery(qSelectAllGauges).WillReturnRows(tc.rows)

			gauges, err := ps.GetAllGauges()

			require.NoError(t, err)
			assert.Equal(t, tc.want, gauges)
		})
	}
}

func TestPostgresStorage_GetAllCounters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rows *sqlmock.Rows
		want map[string]int64
	}{
		{
			name: "multiple counters",
			rows: sqlmock.NewRows([]string{"name", "delta"}).
				AddRow("c1", int64(3)).
				AddRow("c2", int64(-1)),
			want: map[string]int64{"c1": 3, "c2": -1},
		},
		{
			name: "empty table returns empty map",
			rows: sqlmock.NewRows([]string{"name", "delta"}),
			want: map[string]int64{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			mock.ExpectQuery(qSelectAllCounters).WillReturnRows(tc.rows)

			counters, err := ps.GetAllCounters()

			require.NoError(t, err)
			assert.Equal(t, tc.want, counters)
		})
	}
}

// Rows with an injected error on the second row exercise the rows.Err() branch.
func TestPostgresStorage_GetAllGauges_IterationError(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)

	rows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("g1", 1.5).
		AddRow("g2", 2.5)
	rows.RowError(1, errors.New("iteration failed")) // reading the second row fails

	mock.ExpectQuery(qSelectAllGauges).WillReturnRows(rows)

	gauges, err := ps.GetAllGauges()

	assert.Nil(t, gauges)
	assert.ErrorContains(t, err, "error encountered during iteration")
}

func TestPostgresStorage_GetAllCounters_QueryError(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)
	mock.ExpectQuery(qSelectAllCounters).
		WillReturnError(errors.New("db is down"))

	counters, err := ps.GetAllCounters()

	assert.Nil(t, counters)
	assert.ErrorContains(t, err, "error executing query")
}

func TestPostgresStorage_GetAllGauges_ScanError(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)
	mock.ExpectQuery(qSelectAllGauges).
		WillReturnRows(sqlmock.NewRows([]string{"name", "value"}).
			AddRow("g1", "not-a-number"))

	gauges, err := ps.GetAllGauges()

	assert.Nil(t, gauges)
	assert.ErrorContains(t, err, "error scanning row")
}

func TestPostgresStorage_GetAllCounters_ScanError(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)
	mock.ExpectQuery(qSelectAllCounters).
		WillReturnRows(sqlmock.NewRows([]string{"name", "delta"}).
			AddRow("c1", "not-a-number"))

	counters, err := ps.GetAllCounters()

	assert.Nil(t, counters)
	assert.ErrorContains(t, err, "error scanning row")
}

func TestPostgresStorage_GetAllGauges_QueryError(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)
	mock.ExpectQuery(qSelectAllGauges).
		WillReturnError(errors.New("db is down"))

	gauges, err := ps.GetAllGauges()

	assert.Nil(t, gauges)
	assert.ErrorContains(t, err, "error executing query")
}

func TestPostgresStorage_GetAllCounters_IterationError(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)

	rows := sqlmock.NewRows([]string{"name", "delta"}).
		AddRow("c1", int64(1)).
		AddRow("c2", int64(2))
	rows.RowError(1, errors.New("iteration failed")) // reading the second row fails

	mock.ExpectQuery(qSelectAllCounters).WillReturnRows(rows)

	counters, err := ps.GetAllCounters()

	assert.Nil(t, counters)
	assert.ErrorContains(t, err, "error encountered during iteration")
}

// ---------- UpdateMetrics ----------

// TestPostgresStorage_UpdateMetrics_Success verifies that a mixed batch of
// gauge and counter metrics is written inside a single transaction and committed.
func TestPostgresStorage_UpdateMetrics_Success(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)

	gaugeVal := 42.5
	counterVal := int64(7)
	metrics := []model.Metric{
		{ID: "test_gauge", Type: model.Gauge, Value: &gaugeVal},
		{ID: "test_counter", Type: model.Counter, Delta: &counterVal},
	}

	mock.ExpectBegin()
	mock.ExpectExec(qInsertGauge).
		WithArgs("test_gauge", 42.5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(qInsertCounter).
		WithArgs("test_counter", int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := ps.UpdateMetrics(t.Context(), metrics)

	require.NoError(t, err)
}

// TestPostgresStorage_UpdateMetrics_RollbackOnError verifies that a failed
// statement inside the batch aborts the whole transaction with a rollback.
func TestPostgresStorage_UpdateMetrics_RollbackOnError(t *testing.T) {
	t.Parallel()

	dbFailure := errors.New("connection refused")

	tests := []struct {
		name        string
		failCounter bool // fail the counter statement instead of the gauge one
		wantMsg     string
	}{
		{name: "gauge exec failure rolls back", wantMsg: "error executing setGauge"},
		{name: "counter exec failure rolls back", failCounter: true, wantMsg: "error executing addCounter"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)

			gaugeVal := 1.5
			counterVal := int64(3)
			metrics := []model.Metric{
				{ID: "test_gauge", Type: model.Gauge, Value: &gaugeVal},
				{ID: "test_counter", Type: model.Counter, Delta: &counterVal},
			}

			mock.ExpectBegin()
			if tc.failCounter {
				mock.ExpectExec(qInsertGauge).
					WithArgs("test_gauge", 1.5).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(qInsertCounter).
					WithArgs("test_counter", int64(3)).
					WillReturnError(dbFailure)
			} else {
				mock.ExpectExec(qInsertGauge).
					WithArgs("test_gauge", 1.5).
					WillReturnError(dbFailure)
			}
			mock.ExpectRollback()

			err := ps.UpdateMetrics(t.Context(), metrics)

			assert.ErrorIs(t, err, dbFailure)
			assert.ErrorContains(t, err, tc.wantMsg)
		})
	}
}

// TestPostgresStorage_UpdateMetrics_TxLifecycleErrors verifies wrapping of
// Begin and Commit failures; a metric with an unknown type is skipped silently.
func TestPostgresStorage_UpdateMetrics_TxLifecycleErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(sqlmock.Sqlmock)
		wantMsg string
	}{
		{
			name: "begin failure is wrapped",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("conn closed"))
			},
			wantMsg: "error beginning a transaction",
		},
		{
			name: "commit failure is wrapped",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit().WillReturnError(errors.New("conn closed"))
			},
			wantMsg: "error committing a transaction",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			tc.setup(mock)

			// Unknown metric type produces no statement; the transaction still runs.
			err := ps.UpdateMetrics(t.Context(), []model.Metric{
				{ID: "unknown", Type: "random"},
			})

			assert.ErrorContains(t, err, tc.wantMsg)
		})
	}
}

// TestPostgresStorage_AddCounter_Integration runs only when TEST_DATABASE_DSN is set
// and verifies real upsert semantics (delta accumulation) that sqlmock cannot check.
func TestPostgresStorage_AddCounter_Integration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set - skipping integration test")
	}

	ps, err := NewPostgresStorageFromDSN(dsn, "../../migrations")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, ps.Close()) })

	// Unique name keeps the test repeatable: previous runs never affect the sum.
	name := fmt.Sprintf("integration_test_counter_%d", time.Now().UnixNano())

	newDelta, err := ps.AddCounter(name, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), newDelta)

	newDelta, err = ps.AddCounter(name, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(10), newDelta)

	delta, err := ps.GetCounter(name)
	require.NoError(t, err)
	assert.Equal(t, int64(10), delta, "counter must accumulate via ON CONFLICT upsert")
}

// TestPostgresStorage_UpdateMetrics_Integration runs only when TEST_DATABASE_DSN is set
// and verifies real batch upsert semantics (counter accumulation, gauge overwrite)
// within a single transaction.
func TestPostgresStorage_UpdateMetrics_Integration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set - skipping integration test")
	}

	ps, err := NewPostgresStorageFromDSN(dsn, "../../migrations")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, ps.Close()) })

	// Unique names keep the test repeatable: previous runs never affect the sum.
	suffix := time.Now().UnixNano()
	gaugeName := fmt.Sprintf("integration_test_gauge_%d", suffix)
	counterName := fmt.Sprintf("integration_test_counter_%d", suffix)

	gaugeVal := 12.5
	counterVal := int64(5)
	batch := []model.Metric{
		{ID: gaugeName, Type: model.Gauge, Value: &gaugeVal},
		{ID: counterName, Type: model.Counter, Delta: &counterVal},
	}

	require.NoError(t, ps.UpdateMetrics(t.Context(), batch))
	// The same batch again: the counter must accumulate, the gauge must overwrite.
	require.NoError(t, ps.UpdateMetrics(t.Context(), batch))

	value, err := ps.GetGauge(gaugeName)
	require.NoError(t, err)
	assert.InDelta(t, 12.5, value, 0.001)

	delta, err := ps.GetCounter(counterName)
	require.NoError(t, err)
	assert.Equal(t, int64(10), delta, "counter must accumulate via batched ON CONFLICT upsert")
}
