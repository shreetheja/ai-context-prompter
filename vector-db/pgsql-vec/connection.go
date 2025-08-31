package pgsqlvec

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Connect(rdsUsername, rdsPassword, rdsHostname, rdsDatabase string) (*pgxpool.Pool, error) {
	// create the connection string
	connection := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", rdsUsername, rdsPassword,
		rdsHostname, rdsDatabase)
	pgxConfig, err := pgxpool.ParseConfig(connection)
	if err != nil {
		return nil, err
	}

	// Configure connection pool settings
	pgxConfig.MaxConns = 80
	pgxConfig.MinConns = 10
	pgxConfig.MaxConnLifetime = time.Hour * 2
	pgxConfig.MaxConnIdleTime = time.Minute * 15
	pgxConfig.HealthCheckPeriod = time.Minute * 2

	// Configure connection timeouts
	pgxConfig.ConnConfig.ConnectTimeout = time.Second * 30

	runtimeParams := pgxConfig.ConnConfig.RuntimeParams
	runtimeParams["application_name"] = "bandit"
	runtimeParams["idle_in_transaction_session_timeout"] = "300000" // 5 minutes
	runtimeParams["statement_timeout"] = "120000"                   // 2 minutes
	runtimeParams["lock_timeout"] = "30000"                         // 30 seconds

	runtimeParams["tcp_keepalives_idle"] = "300"
	runtimeParams["tcp_keepalives_interval"] = "30"
	runtimeParams["tcp_keepalives_count"] = "3"
	// establish the connection
	pool, err := pgxpool.ConnectConfig(context.Background(), pgxConfig)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return pool, nil
}
