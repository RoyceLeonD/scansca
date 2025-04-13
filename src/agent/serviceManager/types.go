package serviceManager

import (
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	dbmanager "github.com/royceleond/antska/src/agent/dbManager"
	"github.com/royceleond/antska/src/agent/scanner"
)

type ServiceType string

const (
	ServiceTypePostgreSQL ServiceType = "postgresql"
	// Add other service types as needed
)

type Service struct {
	ID        string
	Name      string
	Type      ServiceType
	Config    dbmanager.DBConfig
	CreatedAt time.Time
	UpdatedAt time.Time
}

type JobType string

const (
	JobTypeScan   JobType = "scan"
	JobTypeReport JobType = "report"
	// Add other job types as needed
)

type JobStatus string

const (
	JobStatusPending  JobStatus = "pending"
	JobStatusRunning  JobStatus = "running"
	JobStatusComplete JobStatus = "complete"
	JobStatusFailed   JobStatus = "failed"
)

type Job struct {
	ID        string
	ServiceID string
	Type      JobType
	Status    JobStatus
	Schedule  string // cron expression for scheduled jobs
	CreatedAt time.Time
	UpdatedAt time.Time
	LastRun   time.Time
}

type ServiceManager struct {
	services      map[string]*Service
	jobs          map[string]*Job
	dbManager     *dbmanager.DBManager
	scanner       *scanner.Scanner
	cronScheduler *cron.Cron
	mu            sync.Mutex
}
