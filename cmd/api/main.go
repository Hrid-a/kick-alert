package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Hrid-a/kick-alert/internal/database"
	"github.com/Hrid-a/kick-alert/internal/mailer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

// TODO:  retest all the endpoints
// TODO:  make a Nike scraper product
// TODO: add a webhook to the implementation
// TODO: test the whole flow again
// TODO: push this into github and submit bootdev

type config struct {
	port                    string
	db_dsn                  string
	env                     string
	db_maxOpenConns         int
	db_maxIdleConns         int
	db_maxIdleTime          time.Duration
	front_end_activationURL string
	jwt_secret              string
	limiter                 struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp  smtp
	apify apifyConfig
}

type apifyConfig struct {
	token string
}

type smtp struct {
	host     string
	port     int
	username string
	password string
	sender   string
}

type application struct {
	config config
	logger *slog.Logger
	db     *database.Queries
	wg     sync.WaitGroup
	mailer *mailer.Mailer
}

func main() {

	godotenv.Load()

	db_dsn := os.Getenv("KICK_ALERT_DB_DSN")

	if db_dsn == "" {
		log.Fatal("DB URL must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Port must be set")
	}

	env := os.Getenv("ENV")

	if env == "" {
		log.Fatal("env must be set")
	}

	front_end_activationURL := os.Getenv("FRONTEND_ACTIVATION_URL")

	if front_end_activationURL == "" {
		log.Fatal("env must be set")
	}

	db_maxOpenConns, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))

	if db_maxOpenConns == 0 || err != nil {
		log.Fatal("Postgres max open connections must be set")
	}
	db_maxIdleConns, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))

	if db_maxIdleConns == 0 || err != nil {
		log.Fatal("Postgres max Idle connections must be set")
	}
	db_maxIdleTime, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_TIME"))

	if db_maxIdleTime == 0 || err != nil {
		log.Fatal("Postgres max connection idle time must be set")
	}

	jwt_secret := os.Getenv("JWT_SECRET")

	if jwt_secret == "" {
		log.Fatal("JWT secret must be set")
	}

	limiterEnabled, _ := strconv.ParseBool(os.Getenv("LIMITER_ENABLED"))

	limiterRPS, err := strconv.ParseFloat(os.Getenv("LIMITER_RPS"), 64)
	if err != nil || limiterRPS <= 0 {
		limiterRPS = 2
	}

	limiterBurst, err := strconv.Atoi(os.Getenv("LIMITER_BURST"))
	if err != nil || limiterBurst <= 0 {
		limiterBurst = 4
	}

	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		log.Fatal("SMTP_HOST must be set")
	}

	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil || smtpPort == 0 {
		log.Fatal("SMTP_PORT must be set and be a valid integer")
	}

	smtpUsername := os.Getenv("SMTP_USERNAME")
	if smtpUsername == "" {
		log.Fatal("SMTP_USERNAME must be set")
	}

	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		log.Fatal("SMTP_PASSWORD must be set")
	}

	smtpSender := os.Getenv("SMTP_SENDER")
	if smtpSender == "" {
		log.Fatal("SMTP_SENDER must be set")
	}

	smtpdata := smtp{
		host:     smtpHost,
		port:     smtpPort,
		sender:   smtpSender,
		username: smtpUsername,
		password: smtpPassword,
	}

	apifyToken := os.Getenv("APIFY_TOKEN")
	if apifyToken == "" {
		log.Fatal("APIFY_TOKEN must be set")
	}

	cfg := config{
		db_dsn:                  db_dsn,
		port:                    port,
		env:                     env,
		front_end_activationURL: front_end_activationURL,
		db_maxOpenConns:         db_maxOpenConns,
		db_maxIdleConns:         db_maxIdleConns,
		db_maxIdleTime:          time.Duration(db_maxIdleTime) * time.Minute,
		jwt_secret:              jwt_secret,
		smtp:                    smtpdata,
		apify: apifyConfig{
			token: apifyToken,
		},
	}
	cfg.limiter.enabled = limiterEnabled
	cfg.limiter.rps = limiterRPS
	cfg.limiter.burst = limiterBurst

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbConn, err := openDb(cfg)

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer dbConn.Close()
	logger.Info("Database connection pool established")

	dbQueries := database.New(dbConn)

	mailer, err := mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		config: cfg,
		logger: logger,
		db:     dbQueries,
		mailer: mailer,
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDb(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db_dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db_maxOpenConns)
	db.SetMaxIdleConns(cfg.db_maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db_maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
