package services

import (
	"sync"
	"time"

	"github.com/algonius/algonius-supervisor/pkg/types"
	"go.uber.org/zap"
)

// MetricsCollector collects and provides system metrics
type MetricsCollector struct {
	logger *zap.Logger
	mutex  sync.RWMutex
	
	// Execution metrics
	executionCount       int64
	failedExecutionCount int64
	totalExecutionTime   time.Duration
	executionHistory     []ExecutionMetric
	
	// Agent metrics
	agentMetrics map[string]*AgentMetric
	
	// Resource usage
	resourceUsage *types.ResourceUsage
	
	// Scheduler metrics
	scheduledTaskCount int64
	successfulTaskCount int64
	failedTaskCount    int64
	
	// A2A protocol metrics
	a2aRequestCount int64
	a2aErrorCount   int64
}

// ExecutionMetric represents metrics for a single execution
type ExecutionMetric struct {
	ID              string
	AgentID         string
	ExecutionTime   time.Duration
	Status          types.ExecutionStatus
	StartTime       time.Time
	IsScheduled     bool
}

// AgentMetric represents metrics for a specific agent
type AgentMetric struct {
	ID                  string
	TotalExecutions     int64
	SuccessfulExecutions int64
	FailedExecutions    int64
	AvgExecutionTime    time.Duration
	LastExecutionTime   time.Time
	ActiveExecutions    int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	
	return &MetricsCollector{
		logger:         logger,
		agentMetrics:   make(map[string]*AgentMetric),
		executionHistory: make([]ExecutionMetric, 0),
	}
}

// RecordExecution records metrics for an execution
func (mc *MetricsCollector) RecordExecution(agentID string, executionTime time.Duration, status types.ExecutionStatus, isScheduled bool) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	// Update overall execution metrics
	if status == types.SuccessStatus {
		mc.executionCount++
	} else {
		mc.failedExecutionCount++
	}
	mc.totalExecutionTime += executionTime
	
	// Add to execution history (keep last 100 entries)
	executionMetric := ExecutionMetric{
		ID:            generateMetricID(), // This would be a function to generate IDs
		AgentID:       agentID,
		ExecutionTime: executionTime,
		Status:        status,
		StartTime:     time.Now(),
		IsScheduled:   isScheduled,
	}
	
	mc.executionHistory = append(mc.executionHistory, executionMetric)
	if len(mc.executionHistory) > 100 {
		mc.executionHistory = mc.executionHistory[1:] // Remove oldest
	}
	
	// Update agent-specific metrics
	agentMetric, exists := mc.agentMetrics[agentID]
	if !exists {
		agentMetric = &AgentMetric{
			ID: agentID,
		}
		mc.agentMetrics[agentID] = agentMetric
	}
	
	agentMetric.TotalExecutions++
	if status == types.SuccessStatus {
		agentMetric.SuccessfulExecutions++
	} else {
		agentMetric.FailedExecutions++
	}
	
	// Update average execution time
	totalTime := agentMetric.AvgExecutionTime*time.Duration(agentMetric.TotalExecutions-1) + executionTime
	agentMetric.AvgExecutionTime = totalTime / time.Duration(agentMetric.TotalExecutions)
	agentMetric.LastExecutionTime = time.Now()
}

// RecordScheduledTask records metrics for a scheduled task
func (mc *MetricsCollector) RecordScheduledTask(status types.ExecutionStatus) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.scheduledTaskCount++
	if status == types.SuccessStatus {
		mc.successfulTaskCount++
	} else {
		mc.failedTaskCount++
	}
}

// RecordA2ARequest records metrics for an A2A request
func (mc *MetricsCollector) RecordA2ARequest(isError bool) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.a2aRequestCount++
	if isError {
		mc.a2aErrorCount++
	}
}

// GetExecutionMetrics returns overall execution metrics
func (mc *MetricsCollector) GetExecutionMetrics() map[string]interface{} {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	totalExecutions := mc.executionCount + mc.failedExecutionCount
	successRate := 0.0
	if totalExecutions > 0 {
		successRate = float64(mc.executionCount) / float64(totalExecutions) * 100
	}
	
	avgExecutionTime := time.Duration(0)
	if mc.executionCount > 0 {
		avgExecutionTime = time.Duration(int64(mc.totalExecutionTime) / mc.executionCount)
	}
	
	return map[string]interface{}{
		"total_executions":        totalExecutions,
		"successful_executions":   mc.executionCount,
		"failed_executions":       mc.failedExecutionCount,
		"success_rate_percentage": successRate,
		"average_execution_time":  avgExecutionTime,
		"execution_history":       mc.executionHistory,
	}
}

// GetAgentMetrics returns metrics for a specific agent
func (mc *MetricsCollector) GetAgentMetrics(agentID string) (*AgentMetric, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	agentMetric, exists := mc.agentMetrics[agentID]
	if !exists {
		return nil, false
	}
	
	// Make a copy to prevent race conditions
	metricCopy := *agentMetric
	return &metricCopy, true
}

// GetAllAgentMetrics returns metrics for all agents
func (mc *MetricsCollector) GetAllAgentMetrics() map[string]*AgentMetric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	// Make a copy to prevent race conditions
	result := make(map[string]*AgentMetric)
	for id, metric := range mc.agentMetrics {
		metricCopy := *metric
		result[id] = &metricCopy
	}
	
	return result
}

// GetSchedulerMetrics returns scheduler-specific metrics
func (mc *MetricsCollector) GetSchedulerMetrics() map[string]interface{} {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	totalScheduled := mc.successfulTaskCount + mc.failedTaskCount
	successRate := 0.0
	if totalScheduled > 0 {
		successRate = float64(mc.successfulTaskCount) / float64(totalScheduled) * 100
	}
	
	return map[string]interface{}{
		"total_scheduled_tasks":     mc.scheduledTaskCount,
		"successful_task_executions": mc.successfulTaskCount,
		"failed_task_executions":    mc.failedTaskCount,
		"success_rate_percentage":   successRate,
	}
}

// GetA2AMetrics returns A2A protocol metrics
func (mc *MetricsCollector) GetA2AMetrics() map[string]interface{} {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	totalRequests := mc.a2aRequestCount
	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(mc.a2aErrorCount) / float64(totalRequests) * 100
	}
	
	return map[string]interface{}{
		"total_a2a_requests":      totalRequests,
		"a2a_errors":              mc.a2aErrorCount,
		"error_rate_percentage":   errorRate,
	}
}

// GetOverallMetrics returns all system metrics
func (mc *MetricsCollector) GetOverallMetrics() map[string]interface{} {
	return map[string]interface{}{
		"execution_metrics": mc.GetExecutionMetrics(),
		"agent_metrics":     mc.GetAllAgentMetrics(),
		"scheduler_metrics": mc.GetSchedulerMetrics(),
		"a2a_metrics":       mc.GetA2AMetrics(),
	}
}

// generateMetricID generates a unique ID for a metric entry
func generateMetricID() string {
	// In a real implementation, this would generate a proper unique ID
	return "metric-" + time.Now().Format("20060102-150405")
}