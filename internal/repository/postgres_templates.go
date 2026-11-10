package repository

import "fmt"

const (
	gaugeTableName   = "gauges"
	counterTableName = "counters"
)

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
