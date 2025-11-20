package models

// ResourceUsage represents resource usage information for an agent execution.
type ResourceUsage struct {
	CPUPercent      float64 `json:"cpu_percent"`      // CPU usage percentage during execution
	MemoryMB        int64   `json:"memory_mb"`        // Memory usage in megabytes
	PeakMemoryMB    int64   `json:"peak_memory_mb"`   // Peak memory usage in megabytes
	DiskReadMB      int64   `json:"disk_read_mb"`     // Disk read operations in megabytes
	DiskWriteMB     int64   `json:"disk_write_mb"`    // Disk write operations in megabytes
	NetworkInMB     int64   `json:"network_in_mb"`    // Network input in megabytes
	NetworkOutMB    int64   `json:"network_out_mb"`   // Network output in megabytes
	StartTime       int64   `json:"start_time"`       // Timestamp when resource monitoring started (Unix timestamp)
	EndTime         int64   `json:"end_time"`         // Timestamp when resource monitoring ended (Unix timestamp)
	MeasurementUnit string  `json:"measurement_unit"` // Unit of measurement (e.g., "MB", "KB", "%")
}

// Validate validates the resource usage fields
func (ru *ResourceUsage) Validate() error {
	if ru.CPUPercent < 0 || ru.CPUPercent > 100 {
		return ValidationError("ResourceUsage CPUPercent must be between 0 and 100")
	}

	if ru.MemoryMB < 0 {
		return ValidationError("ResourceUsage MemoryMB cannot be negative")
	}

	if ru.PeakMemoryMB < 0 {
		return ValidationError("ResourceUsage PeakMemoryMB cannot be negative")
	}

	if ru.PeakMemoryMB > 0 && ru.PeakMemoryMB < ru.MemoryMB {
		return ValidationError("ResourceUsage PeakMemoryMB cannot be less than MemoryMB")
	}

	if ru.DiskReadMB < 0 {
		return ValidationError("ResourceUsage DiskReadMB cannot be negative")
	}

	if ru.DiskWriteMB < 0 {
		return ValidationError("ResourceUsage DiskWriteMB cannot be negative")
	}

	if ru.NetworkInMB < 0 {
		return ValidationError("ResourceUsage NetworkInMB cannot be negative")
	}

	if ru.NetworkOutMB < 0 {
		return ValidationError("ResourceUsage NetworkOutMB cannot be negative")
	}

	if ru.StartTime > ru.EndTime && ru.EndTime != 0 {
		return ValidationError("ResourceUsage StartTime cannot be after EndTime")
	}

	return nil
}

// CalculateTotalDiskIO returns the total disk I/O (read + write)
func (ru *ResourceUsage) CalculateTotalDiskIO() int64 {
	return ru.DiskReadMB + ru.DiskWriteMB
}

// CalculateTotalNetworkIO returns the total network I/O (in + out)
func (ru *ResourceUsage) CalculateTotalNetworkIO() int64 {
	return ru.NetworkInMB + ru.NetworkOutMB
}

// GetPeakMemoryPercentage returns the peak memory usage as a percentage of total system memory
// Note: This would typically need system information to calculate properly
func (ru *ResourceUsage) GetPeakMemoryPercentage(totalSystemMemoryMB int64) float64 {
	if totalSystemMemoryMB <= 0 {
		return 0
	}
	return float64(ru.PeakMemoryMB) / float64(totalSystemMemoryMB) * 100
}

// IsHighResourceUsage checks if the resource usage is above certain thresholds
func (ru *ResourceUsage) IsHighResourceUsage(cpuThreshold float64, memoryThreshold int64) bool {
	return ru.CPUPercent > cpuThreshold || ru.PeakMemoryMB > memoryThreshold
}