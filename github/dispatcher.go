package github

import (
	"errors"
	"fmt"
	"github.com/google/go-github/v56/github"
	"lightspeed-bot-controller/github/events"
	"net/http"
)

func handleGithubWebhookRequest(req *http.Request) error {
	payload, err := github.ValidatePayload(req, nil)
	if err != nil {
		return err
	}

	event, err := github.ParseWebHook(github.WebHookType(req), payload)
	if err != nil {
		return err
	}

	switch event.(type) {
	case *github.InstallationEvent:
		ev := event.(*github.InstallationEvent)
		err = events.HandleGithubInstallationEvent(ev)
		if err != nil {
			return err
		}
	case *github.RepositoryEvent:
		ev := event.(*github.RepositoryEvent)
		err = events.HandleGithubRepositoryEvent(ev)
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("unhandled webhook event: %v", event))
	}

	return nil
}
