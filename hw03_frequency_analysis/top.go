package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

type WordFrequency struct {
	Word      string
	Frequency int
}

func Top10(text string) []string {
	// Place your code here.
	if !isCorrect(text) {
		return []string{}
	}

	words := strings.Fields(strings.ToLower(text))

	wordCount := make(map[string]int)

	for _, word := range words {
		if word != "-" {
			cleanWord := strings.Trim(word, "!?,.;")
			wordCount[cleanWord]++
		}
	}

	wordFrequencies := make([]WordFrequency, 0, len(wordCount))
	for word, count := range wordCount {
		wordFrequencies = append(wordFrequencies, WordFrequency{word, count})
	}

	sort.Slice(wordFrequencies, func(i, j int) bool {
		if wordFrequencies[i].Frequency == wordFrequencies[j].Frequency {
			return wordFrequencies[i].Word < wordFrequencies[j].Word
		}
		return wordFrequencies[i].Frequency > wordFrequencies[j].Frequency
	})

	top10 := make([]string, 0, 10)

	for _, wordFrequency := range wordFrequencies {
		top10 = append(top10, wordFrequency.Word)
	}

	return top10[:10]
}

func isCorrect(text string) (result bool) {
	if text == "" || strings.Trim(text, " \t\r\n!?,.;") == "" {
		return false
	}
	return true
}
