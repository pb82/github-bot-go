package events

import (
	"errors"
	"fmt"
	"github.com/google/go-github/v56/github"
	"lightspeed-bot-controller/api"
	"lightspeed-bot-controller/persistence"
	"log"
)

func hasTopic(repo *github.Repository) bool {
	for _, topic := range repo.Topics {
		if topic == "ansible-code-bot-scan" {
			return true
		}
	}
	return false
}

func handleGithubRepositoryEdited(ev *github.RepositoryEvent) error {
	if !hasTopic(ev.Repo) {
		return nil
	}
	log.Println(fmt.Sprintf("repository scan event received for %s", *ev.Repo.FullName))
	schedule := persistence.GetSchedule(*ev.Repo.ID)
	if schedule != nil {
		schedule.Override = true
		persistence.AddOrUpdateSchedule(schedule)
	} else {
		persistence.AddOrUpdateSchedule(&api.RepositorySchedule{
			RepositoryID: *ev.Repo.ID,
			Schedule:     "",
			LastScan:     0,
			Override:     true,
		})
	}
	return nil
}

func HandleGithubRepositoryEvent(ev *github.RepositoryEvent) error {
	switch *ev.Action {
	case "edited":
		return handleGithubRepositoryEdited(ev)
	default:
		return errors.New(fmt.Sprintf("unbknown action: %s", *ev.Action))
	}
}
