//Filename: cmd/api/main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"remindMe/federicorosado.net/internal/data"
	"remindMe/federicorosado.net/internal/jsonlog"
	"time"

	_ "github.com/lib/pq"
)

//The application version number
const version = "1.0.0"

//The conifiguration settings
type config struct {
	port int
	env  string //development,   staging, production, etc.
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64 //request/secod
		burst   int
		enabled bool
	}
}

//Dependency Injection
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	// Declare an instance of the config struct.
	var cfg config
	//read in the flags tahat are need to populateouuuuur config
	flag.IntVar(&cfg.port, "port", 3000, "API server pport")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development | staging | production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("RM_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conss", 20, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conss", 20, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()
	//create a logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	//Create the connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	//Log the successful connection pool
	logger.PrintInfo("database connection pool established", nil)
	//create an instance of our application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	//Call app.serve() to start server
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

//openDB() function returns a *sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	//Create a context with a 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
