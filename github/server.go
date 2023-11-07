package github

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type githubWebhookRouter struct{}

func (g githubWebhookRouter) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch fmt.Sprintf("%v %v", req.Method, req.URL.Path) {
	case "POST /":
		err := handleGithubWebhookRequest(req)
		if err != nil {
			log.Println(err.Error())
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		resp.WriteHeader(http.StatusNotFound)
	}

	resp.WriteHeader(http.StatusOK)
}

func StartGithubWebhookHandler(ctx context.Context, port *int) (*http.Server, error) {
	if port == nil {
		return nil, errors.New("port number required")
	}

	if *port <= 0 || *port > 65536 {
		return nil, errors.New(fmt.Sprintf("%v is not a valid port number", *port))
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%v", *port),
		Handler: githubWebhookRouter{},
	}

	go func() {
		log.Printf("listening for github webhook requests on port %v", *port)
		server.ListenAndServe()
	}()

	return server, nil
}
