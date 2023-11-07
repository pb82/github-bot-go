package events

import (
	"errors"
	"fmt"
	"github.com/google/go-github/v56/github"
	"lightspeed-bot-controller/api"
	"lightspeed-bot-controller/persistence"
	"log"
)

func handleGithubInstallationCreated(ev *github.InstallationEvent) error {
	installation := api.AppInstallation{
		AppID:          *ev.Installation.AppID,
		InstallationID: *ev.Installation.ID,
	}
	persistence.AddInstallation(installation)
	log.Println(fmt.Sprintf("installation %d created", installation.InstallationID))
	return nil
}

func handleGithubInstallationDeleted(ev *github.InstallationEvent) error {
	persistence.RemoveInstallation(*ev.Installation.ID)
	log.Println(fmt.Sprintf("installation %d deleted", *ev.Installation.ID))
	return nil
}

func HandleGithubInstallationEvent(ev *github.InstallationEvent) error {
	switch *ev.Action {
	case "created":
		return handleGithubInstallationCreated(ev)
	case "deleted":
		return handleGithubInstallationDeleted(ev)
	default:
		return errors.New(fmt.Sprintf("unbknown action: %s", *ev.Action))
	}
}
