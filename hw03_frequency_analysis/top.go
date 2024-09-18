package hw03frequencyanalysis

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type WordFrequency struct {
	Word      string
	Frequency int
}

func Top10(text string) []string {
	// Place your code here.
	words, errorText := processWords(text)
	if errorText != nil {
		return []string{}
	}

	wordCount := make(map[string]int)

	for _, word := range words {
		if word != "-" {
			wordCount[word]++
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

	top10 := make([]string, 0)

	for _, wordFrequency := range wordFrequencies {
		if len(top10) == 10 {
			break
		}
		top10 = append(top10, wordFrequency.Word)
	}

	return top10
}

func processWords(text string) ([]string, error) {
	lowerText := strings.ToLower(text)
	re := regexp.MustCompile(`[\p{L}-]+`)
	words := re.FindAllString(lowerText, -1)
	if len(words) == 0 {
		return nil, fmt.Errorf("incorrect text")
	}
	return words, nil
}
