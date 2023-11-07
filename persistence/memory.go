package persistence

import (
	"lightspeed-bot-controller/api"
	"sync"
	"time"
)

var installations []api.AppInstallation
var schedules []api.RepositorySchedule
var lock sync.Mutex

func AddInstallation(installation api.AppInstallation) {
	lock.Lock()
	defer lock.Unlock()

	for _, i := range installations {
		if i.InstallationID == installation.InstallationID {
			return
		}
	}

	installations = append(installations, installation)
}

func AddOrUpdateSchedule(schedule *api.RepositorySchedule) {
	lock.Lock()
	defer lock.Unlock()

	for index, i := range schedules {
		if i.RepositoryID == schedule.RepositoryID {
			schedules[index].Schedule = schedule.Schedule
			schedules[index].Override = schedule.Override
			return
		}
	}

	schedules = append(schedules, *schedule)
}

func RemoveInstallation(installationID int64) {
	lock.Lock()
	defer lock.Unlock()

	var filteredInstallations []api.AppInstallation

	for _, i := range installations {
		if i.InstallationID == installationID {
			continue
		}
		filteredInstallations = append(filteredInstallations, i)
	}

	installations = filteredInstallations
}

func RemoveSchedule(repositoryID int64) {
	lock.Lock()
	defer lock.Unlock()

	var filteredSchedules []api.RepositorySchedule

	for _, i := range schedules {
		if i.RepositoryID == repositoryID {
			continue
		}
		filteredSchedules = append(filteredSchedules, i)
	}

	schedules = filteredSchedules
}

func GetInstallations() []api.AppInstallation {
	return installations
}

func GetSchedule(repositoryID int64) *api.RepositorySchedule {
	for _, schedule := range schedules {
		s := schedule
		if s.RepositoryID == repositoryID {
			return &s
		}
	}
	return nil
}

func UpdateLastScanned(repositoryID int64) {
	for index, schedule := range schedules {
		if schedule.RepositoryID == repositoryID {
			schedules[index].LastScan = time.Now().UTC().Unix()
		}
	}
}
