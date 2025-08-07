package entity

import (
	"encoding/json"
	"time"
)

// EventLog represents a user event log entity
type EventLog struct {
	EventLogID  string 			//EventLog ID (UUID format)
	UserID      string          // User ID (UUID format)
	SessionID   string          // Session identifier (e.g., "550e8400-e29b-41d4-a716-446655440000")
	TraceID     string          // Trace ID (e.g., "550e8400-e29b-41d4-a716-446655440000")
	VideoDmmID  string          // DMM video ID (e.g., "abc123")
	ActressIDs  []string        // Actress IDs (e.g., ["123", "456"])
	DirectorIDs []string        // Director IDs (e.g., ["123", "456"])
	GenreIDs    []string        // Genre IDs (e.g., ["123", "456"])
	MakerIDs    []string        // Maker IDs (e.g., ["123", "456"])
	SeriesIDs   []string        // Series IDs (e.g., ["123", "456"])
	EventType   string          // Event type (e.g., "start", "pause", "skip", "complete", "like", "share")
	EventTime   time.Time       // Event timestamp (e.g., "2025-05-26T10:00:00Z" converted to time.Time)
	Props       json.RawMessage // Additional properties (e.g., { "position": 0 } stored as JSONB)
} 