package main

import (
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type server struct {
	db      *sql.DB
	router  *mux.Router
	logger  *zap.Logger
	authKey *rsa.PublicKey
}

func newServer() *server {
	var err error

	s := server{}
	s.router = mux.NewRouter()

	if s.logger, err = zap.NewProduction(); err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}

	res, err := http.Get("http://trip-pin-points.sa-east-1.elasticbeanstalk.com/key")
	if err != nil {
		s.logger.Fatal("Failed to get authorization key", zap.Error(err))
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		s.logger.Fatal("Failed to read public key from auth service", zap.Error(err))
	}
	block, _ := pem.Decode(bodyBytes)

	if s.authKey, err = x509.ParsePKCS1PublicKey(block.Bytes); err != nil {
		s.logger.Fatal("Failed to parse public key", zap.Error(err))
	}

	s.startDatabase()

	return &s
}

func (s *server) startDatabase() {
	var host, port, user, password, dbname string
	var ok bool
	var err error

	if host, ok = os.LookupEnv("RDS_HOSTNAME"); !ok {
		s.logger.Fatal("Failed to find host environment variable")
	}

	if port, ok = os.LookupEnv("RDS_PORT"); !ok {
		s.logger.Fatal("Failed to find port environment variable")
	}

	if user, ok = os.LookupEnv("RDS_USERNAME"); !ok {
		s.logger.Fatal("Failed to find user environment variable")
	}

	if password, ok = os.LookupEnv("RDS_PASSWORD"); !ok {
		s.logger.Fatal("Failed to find password environment variable")
	}

	if dbname, ok = os.LookupEnv("RDS_DB_NAME"); !ok {
		s.logger.Fatal("Failed to find dbname environment variable")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	s.logger.Info("Trying to connect to", zap.String("connstring", psqlInfo))

	if s.db, err = sql.Open("postgres", psqlInfo); err != nil {
		s.logger.Fatal("Failed to initialize databse")
	}

	err = s.db.Ping()
	if err != nil {
		s.logger.Fatal("Failed to ping database")
	}

	s.logger.Info("Database connected !")

	_, err = s.db.Exec(`
	CREATE TABLE IF NOT EXISTS markers 
	(
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL,
		lat DOUBLE PRECISION NOT NULL,
		long DOUBLE PRECISION NOT NULL,
		note TEXT
	);`)

	if err != nil {
		s.logger.Fatal("Could not create initial table", zap.Error(err))
	}

	return
}

func (s *server) finalize() {
	s.logger.Sync()
	s.db.Close()
}

func main() {
	s := newServer()
	defer s.finalize()

	s.routes()

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "3000"
	}

	http.ListenAndServe(":"+port, s.router)
}
