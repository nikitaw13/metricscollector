package repository

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			exec := mock.ExpectExec(qInsertGauge).
				WithArgs("test_gauge", 42.5)
			if tt.execErr != nil {
				exec.WillReturnError(tt.execErr)
			} else {
				exec.WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err := ps.SetGauge("test_gauge", 42.5)

			if tt.execErr != nil {
				assert.ErrorIs(t, err, tt.execErr)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			exec := mock.ExpectExec(qInsertCounter).
				WithArgs("test_counter", int64(5))
			if tt.execErr != nil {
				exec.WillReturnError(tt.execErr)
			} else {
				exec.WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err := ps.AddCounter("test_counter", 5)

			if tt.execErr != nil {
				assert.ErrorIs(t, err, tt.execErr)
				assert.ErrorContains(t, err, "error writing counter")
			} else {
				assert.NoError(t, err)
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
			name:    "missing gauge maps sql.ErrNoRows to ErrMetricNotFound",
			rows:    sqlmock.NewRows([]string{"value"}),
			wantErr: ErrMetricNotFound,
		},
		{
			name:       "scan failure is wrapped",
			rows:       sqlmock.NewRows([]string{"value"}).AddRow("not-a-number"),
			wantErrMsg: "error scanning row",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			mock.ExpectQuery(qSelectGauge).
				WithArgs("test_gauge").
				WillReturnRows(tt.rows)

			value, err := ps.GetGauge("test_gauge")

			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Zero(t, value)
			case tt.wantErrMsg != "":
				assert.ErrorContains(t, err, tt.wantErrMsg)
				assert.Zero(t, value)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.want, value)
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
			name:    "missing counter maps sql.ErrNoRows to ErrMetricNotFound",
			rows:    sqlmock.NewRows([]string{"delta"}),
			wantErr: ErrMetricNotFound,
		},
		{
			name:       "scan failure is wrapped",
			rows:       sqlmock.NewRows([]string{"delta"}).AddRow("not-a-number"),
			wantErrMsg: "error scanning row",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			mock.ExpectQuery(qSelectCounter).
				WithArgs("test_counter").
				WillReturnRows(tt.rows)

			value, err := ps.GetCounter("test_counter")

			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Zero(t, value)
			case tt.wantErrMsg != "":
				assert.ErrorContains(t, err, tt.wantErrMsg)
				assert.Zero(t, value)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.want, value)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			mock.ExpectQuery(qSelectAllGauges).WillReturnRows(tt.rows)

			result, err := ps.GetAllGauges()

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ps, mock := newMockStorage(t)
			mock.ExpectQuery(qSelectAllCounters).WillReturnRows(tt.rows)

			result, err := ps.GetAllCounters()

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
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

	result, err := ps.GetAllGauges()

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "error encountered during iteration")
}

func TestPostgresStorage_GetAllCounters_QueryError(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)
	mock.ExpectQuery(qSelectAllCounters).
		WillReturnError(errors.New("db is down"))

	result, err := ps.GetAllCounters()

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "error executing query")
}

func TestPostgresStorage_GetAllGauges_ScanError(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)
	mock.ExpectQuery(qSelectAllGauges).
		WillReturnRows(sqlmock.NewRows([]string{"name", "value"}).
			AddRow("g1", "not-a-number"))

	result, err := ps.GetAllGauges()

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "error scanning row")
}

func TestPostgresStorage_GetAllCounters_ScanError(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)
	mock.ExpectQuery(qSelectAllCounters).
		WillReturnRows(sqlmock.NewRows([]string{"name", "delta"}).
			AddRow("c1", "not-a-number"))

	result, err := ps.GetAllCounters()

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "error scanning row")
}

func TestPostgresStorage_GetAllGauges_QueryError(t *testing.T) {
	t.Parallel()

	ps, mock := newMockStorage(t)
	mock.ExpectQuery(qSelectAllGauges).
		WillReturnError(errors.New("db is down"))

	result, err := ps.GetAllGauges()

	assert.Nil(t, result)
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

	result, err := ps.GetAllCounters()

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "error encountered during iteration")
}

// TestPostgres_AddCounter_Integration runs only when TEST_DATABASE_DSN is set
// and verifies real upsert semantics (delta accumulation) that sqlmock cannot check.
func TestPostgres_AddCounter_Integration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set - skipping integration test")
	}

	ps, err := NewPostgresStorageFromDSN(dsn, "../../migrations")
	require.NoError(t, err)
	defer ps.Close()

	// Unique name keeps the test repeatable: previous runs never affect the sum.
	name := fmt.Sprintf("integration_test_counter_%d", time.Now().UnixNano())

	require.NoError(t, ps.AddCounter(name, 5))
	require.NoError(t, ps.AddCounter(name, 3))

	delta, err := ps.GetCounter(name)
	require.NoError(t, err)
	assert.Equal(t, int64(8), delta, "counter must accumulate via ON CONFLICT upsert")
}
