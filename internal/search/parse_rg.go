package search

import (
	"bufio"
	"encoding/json"
	"io"
)

type rgEvent struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines"`
		LineNumber int `json:"line_number"`
		Submatches []struct {
			Match struct {
				Text string `json:"text"`
			} `json:"match"`
			Start int `json:"start"`
			End   int `json:"end"`
		} `json:"submatches"`
	} `json:"data"`
}

func ParseRGOutput(reader io.Reader, contextChars int, limit int) ([]SearchResult, error) {
	results, _, err := ParseRGOutputStream(reader, contextChars, limit)
	return results, err
}

func ParseRGOutputStream(reader io.Reader, contextChars int, limit int) ([]SearchResult, bool, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	results := []SearchResult{}
	limitReached := false
	for scanner.Scan() {
		line := scanner.Bytes()
		event := rgEvent{}
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}
		if event.Type != "match" {
			continue
		}
		if len(event.Data.Submatches) == 0 {
			continue
		}
		sub := event.Data.Submatches[0]
		snippet := trimContext(event.Data.Lines.Text, sub.Start, sub.End, contextChars)
		results = append(results, SearchResult{
			Path:      event.Data.Path.Text,
			Line:      event.Data.LineNumber,
			Column:    sub.Start + 1,
			Match:     sub.Match.Text,
			Snippet:   snippet,
			MatchType: "text",
		})
		if limit > 0 && len(results) >= limit {
			limitReached = true
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, limitReached, err
	}
	return results, limitReached, nil
}

func trimContext(line string, start, end, contextChars int) string {
	runes := []rune(line)
	if contextChars <= 0 {
		contextChars = 80
	}
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	left := start - contextChars
	if left < 0 {
		left = 0
	}
	right := end + contextChars
	if right > len(runes) {
		right = len(runes)
	}
	return string(runes[left:right])
}
