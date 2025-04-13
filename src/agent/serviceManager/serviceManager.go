package serviceManager

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	dbmanager "github.com/royceleond/antska/src/agent/dbManager"
	"github.com/royceleond/antska/src/agent/scanner"
)

func NewServiceManager(dbm *dbmanager.DBManager) *ServiceManager {
	return &ServiceManager{
		services:      make(map[string]*Service),
		jobs:          make(map[string]*Job),
		dbManager:     dbm,
		scanner:       scanner.NewScanner(dbm),
		cronScheduler: cron.New(),
	}
}

func (sm *ServiceManager) RegisterService(name string, serviceType ServiceType, config dbmanager.DBConfig) (*Service, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate service type
	if serviceType != ServiceTypePostgreSQL {
		return nil, fmt.Errorf("unsupported service type: %s", serviceType)
	}

	// Create new service
	service := &Service{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      serviceType,
		Config:    config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store service
	sm.services[service.ID] = service

	return service, nil
}

func (sm *ServiceManager) RegisterJob(serviceID string, jobType JobType, schedule string) (*Job, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if service exists
	if _, exists := sm.services[serviceID]; !exists {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	// Validate job type
	if jobType != JobTypeScan && jobType != JobTypeReport {
		return nil, fmt.Errorf("unsupported job type: %s", jobType)
	}

	// Create new job
	job := &Job{
		ID:        uuid.New().String(),
		ServiceID: serviceID,
		Type:      jobType,
		Status:    JobStatusPending,
		Schedule:  schedule,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store job
	sm.jobs[job.ID] = job

	// Schedule job if a schedule is provided
	if schedule != "" {
		_, err := sm.cronScheduler.AddFunc(schedule, func() {
			sm.runJob(job)
		})
		if err != nil {
			return nil, fmt.Errorf("failed to schedule job: %v", err)
		}
	}

	return job, nil
}

func (sm *ServiceManager) runJob(job *Job) {
	sm.mu.Lock()
	job.Status = JobStatusRunning
	job.LastRun = time.Now()
	sm.mu.Unlock()

	var err error
	switch job.Type {
	case JobTypeScan:
		err = sm.runScanJob(job)
	case JobTypeReport:
		err = sm.runReportJob(job)
	default:
		err = fmt.Errorf("unsupported job type: %s", job.Type)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if err != nil {
		job.Status = JobStatusFailed
	} else {
		job.Status = JobStatusComplete
	}
	job.UpdatedAt = time.Now()
}

func (sm *ServiceManager) runScanJob(job *Job) error {
	service, exists := sm.services[job.ServiceID]
	if !exists {
		return fmt.Errorf("service not found: %s", job.ServiceID)
	}

	// Connect to the database
	err := sm.dbManager.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer sm.dbManager.Close()

	// Perform the scan
	dbInfo, err := sm.scanner.ScanDatabase()
	if err != nil {
		return fmt.Errorf("failed to scan database: %v", err)
	}

	// TODO: Store or process the scan results
	fmt.Printf("Scan completed for service %s (%s). Found %d schemas and %d tables.\n",
		service.Name, service.ID, dbInfo.TotalSchemas, dbInfo.TotalTables)

	return nil
}

func (sm *ServiceManager) runReportJob(job *Job) error {
	// TODO: Implement report generation logic
	fmt.Printf("Report generation not yet implemented for job %s\n", job.ID)
	return nil
}

func (sm *ServiceManager) Start() {
	sm.cronScheduler.Start()
}

func (sm *ServiceManager) Stop() {
	sm.cronScheduler.Stop()
}

func (sm *ServiceManager) GetServices() []*Service {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	services := make([]*Service, 0, len(sm.services))
	for _, service := range sm.services {
		services = append(services, service)
	}
	return services
}
