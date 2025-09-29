package main

import (
	sentry "github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/kelseyhightower/envconfig"
	"github.com/gin-gonic/gin"
)

/*
	Validate the env variables
*/
type SentryEnv struct {
	SentryDSN           string `envconfig:"SENTRY_DSN"`
	SentryEnvironment   string `envconfig:"SENTRY_ENVIRONMENT"`
	SentryEnableTracing string `envconfig:"SENTRY_ENABLE_TRACING"`
}

var sentryEnv SentryEnv

func init() {
	if err := envconfig.Process("", &sentryEnv); err != nil {
		panic(err)
	}
}

/*
	Initialize Sentry
*/
func initSentry(app *gin.Engine) (err error) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:                sentryEnv.SentryDSN,
		Environment:        sentryEnv.SentryEnvironment,
		EnableTracing:      sentryEnv.SentryEnableTracing == "true",
		TracesSampleRate:   1.0,
		IgnoreTransactions: []string{"/lazy-lagoon/healthz/ready", "/lazy-lagoon/healthz/live"},
	}); err != nil {
		return err
	}
	app.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))
	return nil
}
