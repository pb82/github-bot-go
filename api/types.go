package api

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type AppInstallation struct {
	AppID          int64 `json:"appID"`
	InstallationID int64 `json:"installationID"`
}

type RepositorySchedule struct {
	RepositoryID int64  `json:"id"`
	Schedule     string `json:"schedule"`
	LastScan     int64  `json:"lastScan"`
	Override     bool   `json:"override"`
}

type RepositoryConfigSchedule struct {
	Interval string `yaml:"interval"`
}

type RepositoryConfig struct {
	Schedule *RepositoryConfigSchedule `yaml:"schedule,omitempty"`
}

func (i *RepositorySchedule) Elapsed() bool {
	if i.LastScan == 0 {
		return true
	}

	now := time.Now().UTC()
	lastScanTime := time.Unix(i.LastScan, 0)

	switch strings.ToLower(strings.TrimSpace(i.Schedule)) {
	case "hourly":
		delta, _ := time.ParseDuration("1h")
		return lastScanTime.Add(delta).Before(now)
	case "daily":
		delta, _ := time.ParseDuration(fmt.Sprintf("%dh", 1*24))
		return lastScanTime.Add(delta).Before(now)
	case "weekly":
		delta, _ := time.ParseDuration(fmt.Sprintf("%dh", 7*24))
		return lastScanTime.Add(delta).Before(now)
	case "monthly":
		delta, _ := time.ParseDuration(fmt.Sprintf("%dh", 30*24))
		return lastScanTime.Add(delta).Before(now)
	default:
		log.Printf("invalid schedule format: %s", i.Schedule)
		return false
	}
}
