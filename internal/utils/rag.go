package utils

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ChunkTextRunes(text string, chunkSize int, overlap int, maxChunks int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if chunkSize <= 0 {
		chunkSize = 1200
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 4
	}
	if maxChunks <= 0 {
		maxChunks = 200
	}

	r := []rune(text)
	step := chunkSize - overlap
	chunks := make([]string, 0, maxChunks)
	for start := 0; start < len(r) && len(chunks) < maxChunks; start += step {
		end := start + chunkSize
		if end > len(r) {
			end = len(r)
		}
		chunk := strings.TrimSpace(string(r[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(r) {
			break
		}
	}
	return chunks
}

func CosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dot, na, nb float64
	for i := 0; i < n; i++ {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

type ScoredIndex struct {
	Index int
	Score float64
}

func TopK(scores []ScoredIndex, k int) []ScoredIndex {
	if k <= 0 {
		k = 5
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	if len(scores) > k {
		return scores[:k]
	}
	return scores
}

var tokenRe = regexp.MustCompile(`[A-Za-z0-9]+`)

func KeywordScore(query string, text string) float64 {
	qTokens := tokenRe.FindAllString(strings.ToLower(query), -1)
	if len(qTokens) == 0 {
		return 0
	}
	content := strings.ToLower(text)
	score := 0.0
	for _, t := range qTokens {
		if len(t) < 3 {
			continue
		}
		score += float64(strings.Count(content, t))
	}
	return score
}

// ExtractTextFromPDF extracts text content from a PDF file
func ExtractTextFromPDF(filepath string) (string, error) {
	f, r, err := pdf.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var textBuilder strings.Builder
	totalPages := r.NumPage()

	// Limit to first 50 pages to avoid excessive processing
	maxPages := totalPages
	if maxPages > 50 {
		maxPages = 50
	}

	for pageNum := 1; pageNum <= maxPages; pageNum++ {
		p := r.Page(pageNum)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			// Log but continue with other pages
			continue
		}

		textBuilder.WriteString(text)
		textBuilder.WriteString("\n\n")
	}

	return strings.TrimSpace(textBuilder.String()), nil
}
