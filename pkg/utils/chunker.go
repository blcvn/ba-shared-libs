package utils

import (
	"github.com/blcvn/backend/services/pkg/domain"
)

// DocChunker handles splitting ingestion blocks into smaller chunks for LLM processing.
type DocChunker struct {
	MaxTokens int // Rough estimate in characters for now (1 token ≈ 4 chars)
}

func NewDocChunker(maxTokens int) *DocChunker {
	return &DocChunker{MaxTokens: maxTokens}
}

// SplitIntoChunks divides blocks into groups based on Character count.
// It tries to keep blocks together and respects document structure by not splitting
// in the middle of a short block.
func (c *DocChunker) SplitIntoChunks(blocks []domain.Block) [][]domain.Block {
	var chunks [][]domain.Block
	var currentChunk []domain.Block
	currentSize := 0

	for _, block := range blocks {
		blockSize := len(block.Text)

		// If adding this block exceeds MaxTokens, start a new chunk
		// Unless the current chunk is empty (we must include at least one block)
		if currentSize+blockSize > c.MaxTokens && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = []domain.Block{}
			currentSize = 0
		}

		currentChunk = append(currentChunk, block)
		currentSize += blockSize
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
