package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchMetadataEntity(t *testing.T) {
	metadata := SearchMetadata{
		ResultCount:   20,
		TotalCount:    100,
		FirstPosition: 1,
	}

	assert.Equal(t, 20, metadata.ResultCount)
	assert.Equal(t, 100, metadata.TotalCount)
	assert.Equal(t, 1, metadata.FirstPosition)
} 