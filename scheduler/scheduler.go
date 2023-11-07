package scheduler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v56/github"
	"gopkg.in/yaml.v3"
	"lightspeed-bot-controller/api"
	"lightspeed-bot-controller/controller"
	"lightspeed-bot-controller/persistence"
	"log"
	"net/http"
	"time"
)

func downloadSchedule(ctx context.Context, client *github.Client, repo *github.Repository) (*api.RepositorySchedule, error) {
	configFileUrl := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/.github/ansible-code-bot.yml",
		*repo.Owner.Login, *repo.Name, *repo.DefaultBranch)

	req, err := client.NewRequest("GET", configFileUrl, nil)
	if err != nil {
		return nil, err
	}

	var resp bytes.Buffer
	r, err := client.Do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var config api.RepositoryConfig
	err = yaml.Unmarshal(resp.Bytes(), &config)
	if err != nil {
		return nil, err
	}

	if config.Schedule == nil {
		return nil, nil
	}

	schedule := &api.RepositorySchedule{
		RepositoryID: *repo.ID,
		Schedule:     config.Schedule.Interval,
		LastScan:     0,
		Override:     false,
	}

	return schedule, nil
}

func getOrDownloadSchedule(ctx context.Context, client *github.Client, repo *github.Repository) (*api.RepositorySchedule, error) {
	schedule := persistence.GetSchedule(*repo.ID)
	if schedule != nil {
		return schedule, nil
	}

	log.Println(fmt.Sprintf("no schedule found for %s", *repo.FullName))
	newSchedule, err := downloadSchedule(ctx, client, repo)
	if err != nil {
		// no configuration file found
		return nil, nil
	}
	if newSchedule != nil {
		persistence.AddOrUpdateSchedule(newSchedule)
		log.Printf(fmt.Sprintf("retrieved schdeule for repository %s: %s", *repo.FullName, newSchedule.Schedule))
		return newSchedule, nil
	}

	// no schedule file found
	return nil, nil
}

func checkRepositories(ctx context.Context) {
	for _, installation := range persistence.GetInstallations() {
		itr, _ := ghinstallation.NewKeyFromFile(http.DefaultTransport, installation.AppID, installation.InstallationID, "key.pem")
		client := github.NewClient(&http.Client{Transport: itr})

		repos, _, err := client.Apps.ListRepos(ctx, &github.ListOptions{})
		if err != nil {
			log.Println(err.Error())
			continue
		}

		for _, repo := range repos.Repositories {
			log.Println(fmt.Sprintf("processing repository %s", *repo.FullName))
			schedule, err := getOrDownloadSchedule(ctx, client, repo)
			if err != nil {
				log.Println(err.Error())
				continue
			}

			if schedule == nil {
				log.Println(fmt.Sprintf("repository %s does not have a schedule, will be ignored", *repo.FullName))
				continue
			}

			if !schedule.Elapsed() && !schedule.Override {
				log.Printf("%s schedule for %s has not yet elapsed", schedule.Schedule, *repo.FullName)
				continue
			}
			schedule.Override = false

			scanInProgress, err := controller.IsScanInProgress(ctx, *repo.ID)
			if err != nil {
				log.Println(err.Error())
				continue
			}

			if scanInProgress {
				log.Println(fmt.Sprintf("scan for repository %s is already in progress", *repo.FullName))
			}

			token, err := itr.Token(ctx)
			if err != nil {
				log.Println(err.Error())
			}

			log.Println(fmt.Sprintf("starting new scan job for %s", *repo.FullName))
			err = controller.CreateScanJob(ctx, *repo.ID, *repo.Name, *repo.Owner.Login, token)
			if err != nil {
				log.Println(err.Error())
			}
		}
	}
}

func Run(ctx context.Context) error {
	duration, err := time.ParseDuration("10s")
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("shutting down scheduler")
				return
			case <-time.After(duration):
				checkRepositories(ctx)
			}
		}
	}()

	return nil
}
