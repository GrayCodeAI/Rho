package token

import (
	"strings"
)

// Compress applies a deterministic, budget-bounded compression pipeline to
// text. It is a local reimplementation of the former shrike pipeline's
// observable behavior: it removes redundant blank lines, collapses repeated
// lines, and finally enforces the token budget by truncating at a line
// boundary. The returned Stats report original/final token counts and the
// per-layer savings.
func Compress(text string, budget int) (string, Stats) {
	if text == "" {
		return "", Stats{}
	}
	original := EstimateTokens(text)
	stats := Stats{
		OriginalTokens: original,
		Layers:         map[string]LayerStat{},
	}

	// Layer 1: normalize whitespace (collapse 3+ blank lines, trim trailing
	// spaces). Structural, lossless for content.
	deduped := normalizeWhitespace(text)
	stats.Layers["whitespace"] = LayerStat{TokensSaved: max0(original - EstimateTokens(deduped))}

	// Layer 2: collapse runs of identical consecutive lines.
	collapsed := collapseRepeatedLines(deduped)
	stats.Layers["dedupe"] = LayerStat{TokensSaved: max0(EstimateTokens(deduped) - EstimateTokens(collapsed))}

	out := collapsed
	if budget > 0 {
		out = enforceBudget(collapsed, budget)
		stats.Layers["budget"] = LayerStat{TokensSaved: max0(EstimateTokens(collapsed) - EstimateTokens(out))}
	}

	stats.FinalTokens = EstimateTokens(out)
	stats.TokensSaved = stats.OriginalTokens - stats.FinalTokens
	if stats.OriginalTokens > 0 {
		stats.ReductionPercent = float64(stats.TokensSaved) / float64(stats.OriginalTokens) * 100
	}
	return out, stats
}

// CompressForContext compresses text to fit within a token budget, returning
// the compressed text and the final token count.
func CompressForContext(text string, budget int) (string, int) {
	compressed, stats := Compress(text, budget)
	return compressed, stats.FinalTokens
}

func normalizeWhitespace(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	blank := 0
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t\r")
		if strings.TrimSpace(trimmed) == "" {
			blank++
			if blank > 2 {
				continue
			}
		} else {
			blank = 0
		}
		out = append(out, trimmed)
	}
	return strings.Join(out, "\n")
}

func collapseRepeatedLines(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	i := 0
	for i < len(lines) {
		j := i + 1
		for j < len(lines) && lines[j] == lines[i] {
			j++
		}
		run := j - i
		if run >= 3 && strings.TrimSpace(lines[i]) != "" {
			out = append(out, lines[i])
			out = append(out, "[repeated line x"+itoa(run-2)+"]")
			out = append(out, lines[i])
		} else {
			out = append(out, lines[i:j]...)
		}
		i = j
	}
	return strings.Join(out, "\n")
}

func enforceBudget(text string, budget int) string {
	if EstimateTokens(text) <= budget {
		return text
	}
	lines := strings.Split(text, "\n")
	// Binary search the largest line prefix within budget.
	lo, hi := 0, len(lines)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if EstimateTokens(strings.Join(lines[:mid], "\n")) <= budget {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	if lo == 0 {
		return ""
	}
	return strings.Join(lines[:lo], "\n")
}

func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
