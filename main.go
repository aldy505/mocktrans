package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"

	// Database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Dependencies struct {
	DB               *sql.DB
	ServerKey        string
	CallbackUrl      string
	DatabaseProvider string
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "5000"
	}

	serverKey, ok := os.LookupEnv("SERVER_KEY")
	if !ok {
		serverKey = "SB-Mid-server-abc123cde456"
	}

	callbackUrl, ok := os.LookupEnv("CALLBACK_URL")
	if !ok {
		callbackUrl = "localhost"
	}

	databaseProvider, ok := os.LookupEnv("DATABASE_PROVIDER")
	if !ok {
		databaseProvider = "sqlite3"
	}

	databaseUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		databaseUrl = "./database.db"
	}

	db, err := sql.Open(databaseProvider, databaseUrl)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	var maximumOpenConns int = 20
	var maximumIdleConns int = 5

	if databaseProvider == "sqlite3" || databaseProvider == "sqlite" {
		maximumOpenConns = 1
		maximumIdleConns = 1
	}

	db.SetConnMaxLifetime(time.Second * 60)
	db.SetMaxOpenConns(maximumOpenConns)
	db.SetMaxIdleConns(maximumIdleConns)

	dependencies := &Dependencies{
		DB:               db,
		ServerKey:        serverKey,
		CallbackUrl:      callbackUrl,
		DatabaseProvider: databaseProvider,
	}

	app := chi.NewRouter()

	// Check for Authorization
	app.Get("/", dependencies.UserConfirmation)
	app.Put("/confirm", dependencies.Confirm)
	app.Use(dependencies.Authorization)
	app.Get("/healthz", dependencies.Healthz)
	app.Get("/charge", dependencies.Charge)

	server := &http.Server{
		Handler:      app,
		Addr:         ":" + port,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		IdleTimeout:  time.Second * 15,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	go func() {
		log.Printf("Server started on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listening server: %s\n", err)
		}
	}()

	<-sig

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %s\n", err)
	}
}
