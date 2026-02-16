package crawler

import "sync"

// InMemoryRegistry implements Registry using sync.Map
type InMemoryRegistry struct {
	visited sync.Map
}

// NewInMemoryRegistry creates a new in-memory registry
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{}
}

// Visit marks a URL as visited. Returns true if it was visited for the first time.
func (r *InMemoryRegistry) Visit(u string) bool {
	_, loaded := r.visited.LoadOrStore(u, true)
	return !loaded
}

// IsVisited checks if a URL has been visited.
func (r *InMemoryRegistry) IsVisited(u string) bool {
	_, ok := r.visited.Load(u)
	return ok
}
