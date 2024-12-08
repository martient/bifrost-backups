package main

import (
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
	"github.com/martient/bifrost-backups/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	dsn     = "undefined"
)

func main() {
	if dsn != "undefined" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: dsn,
			// Set TracesSampleRate to 1.0 to capture 100%
			// of transactions for tracing.
			// We recommend adjusting this value in production,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}
	cmd.Execute(fmt.Sprintf("version %s, commit %s, built at %s\n", version, commit, date), version)
}
