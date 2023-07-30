package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// global db variable, yiiikkeees
	databaseConn *sql.DB
)

// DatabaseConfig is a struct holding all relevant info for connecting to db
type DatabaseConfig struct {
	username string
	password string
	hostname string
	port     string
	dbName   string
}

// connectDatabase will connect the software to the globall database variable (yikes)
// only do like this in a example
func connectDatabase() error {
	log.Println("trying to connect to database")
	db, err := sql.Open("mysql", createDSN(true))
	if err != nil {
		return fmt.Errorf("failed to open mysql connection: %w", err)
	}
	databaseConn = db

	if err := createDatabase(os.Getenv("DATABASE_NAME")); err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}
	// Reconnect, this time to the database created
	db, err = sql.Open("mysql", createDSN(false))
	if err != nil {
		return fmt.Errorf("failed to open mysql connection using databasename: %w", err)
	}
	log.Println("connected to database")
	databaseConn = db

	return nil
}

// createDatabase will create a database if it does not exist
func createDatabase(dbname string) error {
	log.Println("creating database")
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	tx, err := databaseConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Create database
	res, err := tx.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
	if err != nil {
		return err
	}
	no, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if no == 0 {
		return errors.New("failed to create database, no row affected")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

// createDSN creates a mysql dsn string
// skipDbName can be passed as true to skip using a certain db
func createDSN(skipDbName bool) string {
	dbCfg := getDatabaseConfig()
	if skipDbName {
		return fmt.Sprintf("%s:%s@tcp(%s)/%s", dbCfg.username, dbCfg.password, dbCfg.hostname, "")
	}
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", dbCfg.username, dbCfg.password, dbCfg.hostname, dbCfg.dbName)
}

// getDatabaseConfig fetches database configs from env vars
func getDatabaseConfig() DatabaseConfig {
	// These env vars are manually added in the 02_hellogopher.yml
	username := os.Getenv("DATABASE_USERNAME")
	password := os.Getenv("DATABASE_PASSWORD")
	database := os.Getenv("DATABASE_NAME")
	// these env vars are automatically added by Kubernetes when pod is created if Service named Mysql Exists
	hostname := os.Getenv("MYSQL_SERVICE_HOST")
	port := os.Getenv("MYSQL_SERVICE_PORT")

	return DatabaseConfig{
		username: username,
		password: password,
		hostname: hostname,
		port:     port,
		dbName:   database,
	}
}
