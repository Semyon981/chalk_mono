package main

import (
	"chalk/internal/repo"
	thttp "chalk/internal/transport/http"
	"chalk/internal/usecases"
	"chalk/pkg/config"
	"chalk/pkg/log"
	"chalk/pkg/mailer"
	"chalk/pkg/migrator"
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.SetLevel(log.DEBUG)

	// parse flags
	configPath := flag.String("config", "/etc/chalk/config/config.yaml", "path to config file")
	migrationsPath := flag.String("migrations", "/etc/chalk/migrations", "path to the directory with migration files")
	flag.Parse()

	// load config
	cfg := config.LoadConfig(*configPath)

	// postgres connect
	pcli, err := pgx.Connect(context.Background(), cfg.PostgresURI)
	if err != nil {
		log.Errorf("postgres connect: %v", err)
		return
	}
	defer pcli.Close(context.Background())

	// migrations
	migrator := migrator.NewPgx(pcli, *migrationsPath)
	err = migrator.Up()
	if err != nil {
		log.Errorf("up migrations error: %v", err)
	}

	// redis connect
	opts, err := redis.ParseURL(cfg.RedisURI)
	if err != nil {
		log.Errorf("parse redis uri: %v", err)
		return
	}
	rcli := redis.NewClient(opts)

	// s3 connect
	endpoint := "play.min.io"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := false
	miniocli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Errorf("minio connect: %v", err)
		return
	}

	fmt.Println("miniocli", miniocli)

	// repositories
	acrepo := repo.NewAuthCodeRepo(rcli)
	srepo := repo.NewSessionsRepo(pcli)
	urepo := repo.NewUsersRepo(pcli)

	// mailer
	amailer := mailer.New(
		cfg.Auth.Mailer.SmtpHost,
		cfg.Auth.Mailer.SmtpPort,
		cfg.Auth.Mailer.SmtpSASLUsername,
		cfg.Auth.Mailer.SmtpSASLPassword,
	)

	// usecases
	auc := usecases.NewAuthUseCase(
		acrepo,
		urepo,
		srepo,
		cfg.Auth.CodeTTL,
		amailer,
		cfg.Auth.EmailFromAddr,
		cfg.Auth.EmailFromName,
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.RefreshTokenTTL,
	)

	handler := thttp.NewHandler(auc)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: handler,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Errorf("server startup error: %v", err)
		return
	}
}
