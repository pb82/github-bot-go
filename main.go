package main

import (
	"context"
	"flag"
	"lightspeed-bot-controller/controller"
	"lightspeed-bot-controller/github"
	"lightspeed-bot-controller/scheduler"
	"os"
	"os/signal"
	"syscall"
)

var (
	port  *int
	appid *int64
)

func main() {
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE, syscall.SIGABRT)
	defer stop()

	server, err := github.StartGithubWebhookHandler(ctx, port)
	if err != nil {
		panic(err)
	}

	controller.CreateKubernetesClient(ctx)
	controller.RunListener(ctx)

	err = scheduler.Run(ctx)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
	server.Shutdown(ctx)
}

func init() {
	port = flag.Int("port", 8080, "port to receive webhook events")
	appid = flag.Int64("appid", 0, "github app id")
}
