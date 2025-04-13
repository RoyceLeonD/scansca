package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/royceleond/antska/internal/aml/registry"
	"github.com/royceleond/antska/internal/aml/scheduler"
)

// CreateJobRequest represents the request body for creating a job
type CreateJobRequest struct {
	Name         string `json:"name" binding:"required"`
	DatabaseName string `json:"database_name" binding:"required"`
	Type         string `json:"type" binding:"required"`
	Schedule     string `json:"schedule" binding:"required"`
	Query        string `json:"query,omitempty"`
}

// JobResponse represents a job in the response format
type JobResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	DatabaseName string    `json:"database_name"`
	Type         string    `json:"type"`
	Schedule     string    `json:"schedule"`
	Status       string    `json:"status"`
	LastRun      string    `json:"last_run,omitempty"`
	NextRun      string    `json:"next_run,omitempty"`
	LastError    string    `json:"last_error,omitempty"`
}

// jobToResponse converts a scheduler.Job to a JobResponse
func jobToResponse(job *scheduler.Job) JobResponse {
	response := JobResponse{
		ID:           job.ID,
		Name:         job.Name,
		DatabaseName: job.DatabaseName,
		Type:         "generic", // Default
		Schedule:     job.Schedule,
		Status:       string(job.Status),
	}

	// Add timestamps if available
	if job.LastRun != nil {
		response.LastRun = job.LastRun.Format("2006-01-02T15:04:05Z07:00")
	}
	if job.NextRun != nil {
		response.NextRun = job.NextRun.Format("2006-01-02T15:04:05Z07:00")
	}

	// Add error message if available
	if job.LastError != nil {
		response.LastError = job.LastError.Error()
	}

	return response
}

// CreateJobHandler handles creating a new scheduled job
func CreateJobHandler(registry *registry.Registry, scheduler *scheduler.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request CreateJobRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify that the database exists
		_, err := registry.GetConnector(request.DatabaseName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Database not found: %s", err.Error()),
			})
			return
		}

		// Create the job function based on the job type
		var jobFunc scheduler.JobFunction

		if request.Type == "query" {
			// Query job type requires a query parameter
			if request.Query == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Query parameter is required for query job type",
				})
				return
			}

			// Create a query job
			query := request.Query // Capture for closure
			jobFunc = func(ctx context.Context) error {
				connector, err := registry.GetConnector(request.DatabaseName)
				if err != nil {
					return fmt.Errorf("failed to get connector: %w", err)
				}

				_, err = connector.ExecuteQuery(ctx, query)
				if err != nil {
					return fmt.Errorf("failed to execute query: %w", err)
				}

				return nil
			}
		} else {
			// Generic job for now (can be expanded later)
			jobFunc = func(ctx context.Context) error {
				log.Info().
					Str("job", request.Name).
					Str("database", request.DatabaseName).
					Msg("Generic job executed")
				return nil
			}
		}

		// Schedule the job
		jobID, err := scheduler.AddJob(request.Name, request.Schedule, request.DatabaseName, jobFunc)
		if err != nil {
			log.Error().Err(err).
				Str("name", request.Name).
				Str("schedule", request.Schedule).
				Msg("Failed to schedule job")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the job to include in the response
		job, _ := scheduler.GetJob(jobID)

		log.Info().
			Str("id", jobID).
			Str("name", request.Name).
			Str("schedule", request.Schedule).
			Msg("Job scheduled successfully")

		c.JSON(http.StatusCreated, gin.H{
			"message": "Job scheduled successfully",
			"job":     jobToResponse(job),
		})
	}
}

// ListJobsHandler handles listing all scheduled jobs
func ListJobsHandler(scheduler *scheduler.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobs := scheduler.ListJobs()

		// Convert to response format
		jobResponses := make([]JobResponse, len(jobs))
		for i, job := range jobs {
			jobResponses[i] = jobToResponse(job)
		}

		c.JSON(http.StatusOK, gin.H{
			"jobs": jobResponses,
		})
	}
}

// GetJobHandler handles retrieving a specific job
func GetJobHandler(scheduler *scheduler.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")
		if jobID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
			return
		}

		job, err := scheduler.GetJob(jobID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"job": jobToResponse(job),
		})
	}
}

// DeleteJobHandler handles removing a scheduled job
func DeleteJobHandler(scheduler *scheduler.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")
		if jobID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
			return
		}

		err := scheduler.RemoveJob(jobID)
		if err != nil {
			log.Error().Err(err).
				Str("id", jobID).
				Msg("Failed to delete job")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Info().
			Str("id", jobID).
			Msg("Job deleted successfully")

		c.JSON(http.StatusOK, gin.H{
			"message": "Job deleted successfully",
		})
	}
}

// RunJobHandler handles running a job immediately
func RunJobHandler(scheduler *scheduler.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")
		if jobID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
			return
		}

		err := scheduler.RunJobNow(jobID)
		if err != nil {
			log.Error().Err(err).
				Str("id", jobID).
				Msg("Failed to run job")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Info().
			Str("id", jobID).
			Msg("Job execution triggered")

		c.JSON(http.StatusOK, gin.H{
			"message": "Job execution triggered",
		})
	}
}