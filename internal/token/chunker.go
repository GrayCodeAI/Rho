package token

import "strings"

// CodeChunk represents a chunk of source code for semantic search.
type CodeChunk struct {
	Content   string `json:"content"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Symbol    string `json:"symbol,omitempty"`
	Tokens    int    `json:"tokens"`
}

// ChunkOptions configures code chunking.
type ChunkOptions struct {
	MaxTokens     int
	MinTokens     int
	MinChunkSize  int
	Language      string
	Overlap       int
	KeepSeparator int
}

// DefaultChunkOptions returns sensible defaults for code chunking.
func DefaultChunkOptions() ChunkOptions {
	return ChunkOptions{MaxTokens: 1000, MinTokens: 100}
}

// ChunkCode splits source into line-bounded chunks that respect the token
// budget. It is a local, dependency-free implementation: chunks break on line
// boundaries and never exceed MaxTokens (default 1000).
func ChunkCode(source string, opts ChunkOptions) []CodeChunk {
	if strings.TrimSpace(source) == "" {
		return nil
	}
	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1000
	}
	lines := strings.Split(source, "\n")
	var chunks []CodeChunk
	start := 0
	cur := 0
	for i, line := range lines {
		lineTokens := EstimateTokensFast(line) + 1
		if cur > 0 && cur+lineTokens > maxTokens {
			chunks = append(chunks, makeChunk(lines, start, i))
			start = i
			cur = 0
		}
		cur += lineTokens
	}
	if start < len(lines) {
		chunks = append(chunks, makeChunk(lines, start, len(lines)))
	}
	return chunks
}

func makeChunk(lines []string, start, end int) CodeChunk {
	content := strings.Join(lines[start:end], "\n")
	return CodeChunk{
		Content:   content,
		StartLine: start + 1,
		EndLine:   end,
		Tokens:    EstimateTokensFast(content),
	}
}