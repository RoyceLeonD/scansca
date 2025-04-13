package main

import (
	"fmt"
	"log"
	"time"

	dbmanager "github.com/royceleond/antska/src/agent/dbManager"
	"github.com/royceleond/antska/src/agent/serviceManager"
)

func main() {
	dbManager := dbmanager.NewDBManager(dbmanager.DBConfig{})
	sm := serviceManager.NewServiceManager(dbManager)

	// Start the service manager
	sm.Start()
	defer sm.Stop()
	postgresConfig := dbmanager.DBConfig{
		Host:     "localhost",
		Port:     5433,
		User:     "aiuser",
		Password: "HfF9EJinHxMxGf9sAKgi8RqM&s",
		DBName:   "vehicalretail",
	}

	service, err := sm.RegisterService("My PostgreSQL DB", serviceManager.ServiceTypePostgreSQL, postgresConfig)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}
	fmt.Printf("Registered service: %s (%s)\n", service.Name, service.ID)

	// Register a scan job that runs every minute
	scanJob, err := sm.RegisterJob(service.ID, serviceManager.JobTypeScan, "* * * * *")
	if err != nil {
		log.Fatalf("Failed to register scan job: %v", err)
	}
	fmt.Printf("Registered scan job: %s\n", scanJob.ID)

	// Register a report job that runs every 5 minutes
	reportJob, err := sm.RegisterJob(service.ID, serviceManager.JobTypeReport, "*/5 * * * *")
	if err != nil {
		log.Fatalf("Failed to register report job: %v", err)
	}
	fmt.Printf("Registered report job: %s\n", reportJob.ID)

	// Keep the program running to allow scheduled jobs to execute
	fmt.Println("Service manager is running. Press Ctrl+C to stop.")
	for {
		time.Sleep(time.Minute)
	}
}
