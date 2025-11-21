package services

import (
	"sync"
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"
)

// ExecutionResultCache provides caching for execution results
type ExecutionResultCache struct {
	cache map[string]*models.ExecutionResult
	ttl   time.Duration // Time-to-live for cached results
	mutex sync.RWMutex
	
	// Optional: cleanup ticker for removing expired entries
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// NewExecutionResultCache creates a new execution result cache
func NewExecutionResultCache(ttl time.Duration) *ExecutionResultCache {
	cache := &ExecutionResultCache{
		cache:       make(map[string]*models.ExecutionResult),
		ttl:         ttl,
		cleanupTicker: time.NewTicker(5 * time.Minute), // Clean up every 5 minutes
		stopCleanup:   make(chan bool),
	}

	// Start the cleanup goroutine
	go cache.startCleanup()
	
	return cache
}

// Put adds an execution result to the cache
func (c *ExecutionResultCache) Put(result *models.ExecutionResult) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache[result.ID] = result
}

// Get retrieves an execution result from the cache
func (c *ExecutionResultCache) Get(id string) (*models.ExecutionResult, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	result, exists := c.cache[id]
	if !exists {
		return nil, false
	}
	
	// Check if the result has expired (based on execution time + TTL)
	if time.Since(result.EndTime) > c.ttl {
		// Remove expired entry
		delete(c.cache, id)
		return nil, false
	}
	
	return result, true
}

// Remove removes an execution result from the cache
func (c *ExecutionResultCache) Remove(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.cache, id)
}

// Clear removes all entries from the cache
func (c *ExecutionResultCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache = make(map[string]*models.ExecutionResult)
}

// Size returns the number of entries in the cache
func (c *ExecutionResultCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return len(c.cache)
}

// startCleanup starts the cleanup goroutine
func (c *ExecutionResultCache) startCleanup() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.cleanupExpired()
		case <-c.stopCleanup:
			c.cleanupTicker.Stop()
			return
		}
	}
}

// cleanupExpired removes expired entries from the cache
func (c *ExecutionResultCache) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	now := time.Now()
	for id, result := range c.cache {
		if now.Sub(result.EndTime) > c.ttl {
			delete(c.cache, id)
		}
	}
}

// Close stops the cleanup goroutine
func (c *ExecutionResultCache) Close() {
	close(c.stopCleanup)
}

// MetricsCache provides caching for metrics data
type MetricsCache struct {
	cache map[string]interface{}
	ttl   time.Duration
	mutex sync.RWMutex
}

// NewMetricsCache creates a new metrics cache
func NewMetricsCache(ttl time.Duration) *MetricsCache {
	return &MetricsCache{
		cache: make(map[string]interface{}),
		ttl:   ttl,
	}
}

// Put adds metrics data to the cache
func (c *MetricsCache) Put(key string, data interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache[key] = data
}

// Get retrieves metrics data from the cache
func (c *MetricsCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	data, exists := c.cache[key]
	if !exists {
		return nil, false
	}
	
	return data, true
}

// Remove removes metrics data from the cache
func (c *MetricsCache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.cache, key)
}

// AgentConnectionPool manages connections to frequently accessed agents
type AgentConnectionPool struct {
	connections map[string]interface{} // In a real implementation, this would hold actual connections
	maxSize     int
	mutex       sync.RWMutex
}

// NewAgentConnectionPool creates a new agent connection pool
func NewAgentConnectionPool(maxSize int) *AgentConnectionPool {
	return &AgentConnectionPool{
		connections: make(map[string]interface{}),
		maxSize:     maxSize,
	}
}

// GetConnection attempts to get a pooled connection for the agent
func (p *AgentConnectionPool) GetConnection(agentID string) (interface{}, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	conn, exists := p.connections[agentID]
	return conn, exists
}

// PutConnection attempts to store a connection in the pool
func (p *AgentConnectionPool) PutConnection(agentID string, conn interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if len(p.connections) >= p.maxSize {
		// Remove a random connection to make space
		for id := range p.connections {
			delete(p.connections, id)
			break
		}
	}
	
	p.connections[agentID] = conn
}

// RemoveConnection removes a connection from the pool
func (p *AgentConnectionPool) RemoveConnection(agentID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	delete(p.connections, agentID)
}