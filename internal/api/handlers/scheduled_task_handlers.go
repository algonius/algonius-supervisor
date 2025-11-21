package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/algonius/algonius-supervisor/internal/services"
	"github.com/algonius/algonius-supervisor/internal/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ScheduledTaskHandlers handles scheduled task API requests
type ScheduledTaskHandlers struct {
	schedulerService services.ISchedulerService
	logger           *zap.Logger
}

// NewScheduledTaskHandlers creates a new instance of ScheduledTaskHandlers
func NewScheduledTaskHandlers(schedulerService services.ISchedulerService, logger *zap.Logger) *ScheduledTaskHandlers {
	return &ScheduledTaskHandlers{
		schedulerService: schedulerService,
		logger:           logger,
	}
}

// RegisterScheduledTaskRoutes registers all scheduled task-related routes
func (sth *ScheduledTaskHandlers) RegisterScheduledTaskRoutes(router *gin.Engine) {
	// Register scheduled task routes
	taskGroup := router.Group("/tasks")
	
	// Apply authentication middleware if required
	// taskGroup.Use(authenticationMiddleware) // This would be added based on security requirements
	
	taskGroup.GET("", sth.ListTasks)
	taskGroup.GET("/:taskId", sth.GetTask)
	taskGroup.POST("", sth.CreateTask)
	taskGroup.PUT("/:taskId", sth.UpdateTask)
	taskGroup.DELETE("/:taskId", sth.DeleteTask)
	taskGroup.POST("/:taskId/execute", sth.ExecuteTask)
	taskGroup.POST("/:taskId/pause", sth.PauseTask)
	taskGroup.POST("/:taskId/resume", sth.ResumeTask)
}

// ListTasks returns a list of all scheduled tasks
func (sth *ScheduledTaskHandlers) ListTasks(c *gin.Context) {
	sth.logger.Info("handling list tasks request")

	tasks, err := sth.schedulerService.ListScheduledTasks()
	if err != nil {
		sth.logger.Error("failed to list tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list tasks",
		})
		return
	}

	// Convert tasks to response format
	taskList := make([]gin.H, len(tasks))
	for i, task := range tasks {
		taskList[i] = gin.H{
			"id":             task.ID,
			"name":           task.Name,
			"agent_id":       task.AgentID,
			"cron_expression": task.CronExpression,
			"enabled":        task.Enabled,
			"active":         task.Active,
			"created_at":     task.CreatedAt,
			"updated_at":     task.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": taskList,
		"total": len(taskList),
	})
}

// GetTask returns details for a specific scheduled task
func (sth *ScheduledTaskHandlers) GetTask(c *gin.Context) {
	taskID := c.Param("taskId")

	sth.logger.Info("handling get task request",
		zap.String("task_id", taskID))

	task, err := sth.schedulerService.GetTask(taskID)
	if err != nil {
		sth.logger.Error("task not found",
			zap.String("task_id", taskID),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	response := gin.H{
		"id":               task.ID,
		"name":             task.Name,
		"agent_id":         task.AgentID,
		"cron_expression":  task.CronExpression,
		"enabled":          task.Enabled,
		"active":           task.Active,
		"input_parameters": task.InputParameters,
		"created_at":       task.CreatedAt,
		"updated_at":       task.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// CreateTask creates a new scheduled task
func (sth *ScheduledTaskHandlers) CreateTask(c *gin.Context) {
	var requestData struct {
		Name            string                 `json:"name"`
		AgentID         string                 `json:"agent_id"`
		CronExpression  string                 `json:"cron_expression"`
		Enabled         bool                   `json:"enabled"`
		InputParameters map[string]interface{} `json:"input_parameters"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		sth.logger.Error("failed to parse create task request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Create the scheduled task model
	task := &models.ScheduledTask{
		ID:              generateTaskID(), // This would be a function to generate unique IDs
		Name:            requestData.Name,
		AgentID:         requestData.AgentID,
		CronExpression:  requestData.CronExpression,
		Enabled:         requestData.Enabled,
		Active:          requestData.Enabled, // Active by default if enabled
		InputParameters: requestData.InputParameters,
	}

	// Schedule the task
	err := sth.schedulerService.ScheduleTask(task)
	if err != nil {
		sth.logger.Error("failed to schedule task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to schedule task",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task scheduled successfully",
		"task_id": task.ID,
	})
}

// UpdateTask updates an existing scheduled task
func (sth *ScheduledTaskHandlers) UpdateTask(c *gin.Context) {
	taskID := c.Param("taskId")

	var requestData struct {
		Name            string                 `json:"name"`
		AgentID         string                 `json:"agent_id"`
		CronExpression  string                 `json:"cron_expression"`
		Enabled         bool                   `json:"enabled"`
		InputParameters map[string]interface{} `json:"input_parameters"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		sth.logger.Error("failed to parse update task request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Check if the task exists
	existingTask, err := sth.schedulerService.GetTask(taskID)
	if err != nil {
		sth.logger.Error("task not found for update",
			zap.String("task_id", taskID),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	// Update the task properties
	existingTask.Name = requestData.Name
	existingTask.AgentID = requestData.AgentID
	existingTask.CronExpression = requestData.CronExpression
	existingTask.Enabled = requestData.Enabled
	existingTask.InputParameters = requestData.InputParameters

	// Update the task in the scheduler
	err = sth.schedulerService.UpdateTask(existingTask)
	if err != nil {
		sth.logger.Error("failed to update task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task updated successfully",
		"task_id": existingTask.ID,
	})
}

// DeleteTask deletes a scheduled task
func (sth *ScheduledTaskHandlers) DeleteTask(c *gin.Context) {
	taskID := c.Param("taskId")

	sth.logger.Info("handling delete task request",
		zap.String("task_id", taskID))

	err := sth.schedulerService.UnscheduleTask(taskID)
	if err != nil {
		sth.logger.Error("failed to delete task",
			zap.String("task_id", taskID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
		"task_id": taskID,
	})
}

// ExecuteTask immediately executes a scheduled task
func (sth *ScheduledTaskHandlers) ExecuteTask(c *gin.Context) {
	taskID := c.Param("taskId")

	sth.logger.Info("handling execute task request",
		zap.String("task_id", taskID))

	result, err := sth.schedulerService.ExecuteTask(taskID)
	if err != nil {
		sth.logger.Error("failed to execute task",
			zap.String("task_id", taskID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to execute task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task executed successfully",
		"task_id": taskID,
		"result": gin.H{
			"execution_id": result.ID,
			"status":       string(result.Status),
			"output":       result.Output,
			"execution_time_ms": result.ExecutionTime,
		},
	})
}

// PauseTask pauses a scheduled task
func (sth *ScheduledTaskHandlers) PauseTask(c *gin.Context) {
	taskID := c.Param("taskId")

	sth.logger.Info("handling pause task request",
		zap.String("task_id", taskID))

	err := sth.schedulerService.PauseTask(taskID)
	if err != nil {
		sth.logger.Error("failed to pause task",
			zap.String("task_id", taskID),
			zap.Error(err))
		if err.Error() == "task with ID "+taskID+" not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to pause task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task paused successfully",
		"task_id": taskID,
	})
}

// ResumeTask resumes a paused task
func (sth *ScheduledTaskHandlers) ResumeTask(c *gin.Context) {
	taskID := c.Param("taskId")

	sth.logger.Info("handling resume task request",
		zap.String("task_id", taskID))

	err := sth.schedulerService.ResumeTask(taskID)
	if err != nil {
		sth.logger.Error("failed to resume task",
			zap.String("task_id", taskID),
			zap.Error(err))
		if err.Error() == "task with ID "+taskID+" not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to resume task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task resumed successfully",
		"task_id": taskID,
	})
}

// Helper function to generate task IDs (in a real implementation, this would be more sophisticated)
func generateTaskID() string {
	// In a real implementation, this could use UUID generation
	return "task-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}