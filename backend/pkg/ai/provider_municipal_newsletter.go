package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kozaktomas/diacritics"
)

const municipalNewsletterURL = "https://static.kozak.in/pub-ai/zpravodaj.json"
const municipalNewsletterCacheTTL = 1 * time.Hour

type MunicipalNewsletterIssue struct {
	ID       string                       `json:"id"`
	Title    string                       `json:"title"`
	Date     string                       `json:"date"`
	Articles []MunicipalNewsletterArticle `json:"articles"`
}

type MunicipalNewsletterArticle struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Content string `json:"content"`
}

type MunicipalNewsletterSearchResult struct {
	ArticleID      string `json:"article_id"`
	Title          string `json:"title"`
	IssueTitle     string `json:"issue_title"`
	IssueDate      string `json:"issue_date"`
	RelevanceScore int    `json:"relevance_score"`
}

type municipalNewsletterCache struct {
	mu        sync.RWMutex
	data      []MunicipalNewsletterIssue
	fetchedAt time.Time
}

var mnCache = &municipalNewsletterCache{}

func fetchMunicipalNewsletterData() ([]MunicipalNewsletterIssue, error) {
	mnCache.mu.RLock()
	if mnCache.data != nil && time.Since(mnCache.fetchedAt) < municipalNewsletterCacheTTL {
		defer mnCache.mu.RUnlock()
		return mnCache.data, nil
	}
	mnCache.mu.RUnlock()

	mnCache.mu.Lock()
	defer mnCache.mu.Unlock()

	// Double-check after acquiring write lock
	if mnCache.data != nil && time.Since(mnCache.fetchedAt) < municipalNewsletterCacheTTL {
		return mnCache.data, nil
	}

	issues, err := downloadMunicipalNewsletterData()
	if err != nil {
		// If download fails and we have cached data, return the old data
		if mnCache.data != nil {
			return mnCache.data, nil
		}
		return nil, err
	}

	mnCache.data = issues
	mnCache.fetchedAt = time.Now()

	return issues, nil
}

func downloadMunicipalNewsletterData() ([]MunicipalNewsletterIssue, error) {
	resp, err := http.DefaultClient.Get(municipalNewsletterURL)
	if err != nil {
		return nil, fmt.Errorf("could not fetch municipal newsletter data: %w", err)
	}
	defer resp.Body.Close() //nolint: errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read municipal newsletter response: %w", err)
	}

	var issues []MunicipalNewsletterIssue
	if err := json.Unmarshal(body, &issues); err != nil {
		return nil, fmt.Errorf("could not unmarshal municipal newsletter data: %w", err)
	}

	return issues, nil
}

type scoredSearchResult struct {
	result MunicipalNewsletterSearchResult
	score  int
}

// ProvideMunicipalNewsletterSearch searches articles by query string
func ProvideMunicipalNewsletterSearch(query string) (string, error) {
	issues, err := fetchMunicipalNewsletterData()
	if err != nil {
		return "", err
	}

	queryNormalized := normalizeText(query)
	queryWords := strings.Fields(queryNormalized)

	var scored []scoredSearchResult

	for _, issue := range issues {
		for _, article := range issue.Articles {
			titleNormalized := normalizeText(article.Title)
			contentNormalized := normalizeText(article.Content)

			score := calculateRelevanceScore(titleNormalized, contentNormalized, queryNormalized, queryWords, issue.Date)
			if score > 0 {
				scored = append(scored, scoredSearchResult{
					result: MunicipalNewsletterSearchResult{
						ArticleID:      article.ID,
						Title:          article.Title,
						IssueTitle:     issue.Title,
						IssueDate:      issue.Date,
						RelevanceScore: score,
					},
					score: score,
				})
			}
		}
	}

	if len(scored) == 0 {
		return "No articles found matching the query.", nil
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Limit to 10 most relevant results
	if len(scored) > 10 {
		scored = scored[:10]
	}

	results := make([]MunicipalNewsletterSearchResult, len(scored))
	for i, s := range scored {
		results[i] = s.result
	}

	output, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("could not marshal search results: %w", err)
	}

	return string(output), nil
}

// Czech stop words to ignore in search
var czechStopWords = map[string]bool{
	"a": true, "i": true, "o": true, "u": true, "v": true, "z": true, "k": true, "s": true,
	"je": true, "se": true, "na": true, "ze": true, "do": true, "to": true, "za": true,
	"po": true, "od": true, "pro": true, "pri": true, "tak": true, "jak": true, "ale": true,
	"jsou": true, "byl": true, "byla": true, "bylo": true, "byli": true, "bude": true,
	"jsme": true, "jste": true, "jako": true, "jeho": true, "jej": true, "jeji": true,
	"jejich": true, "tento": true, "tato": true, "toto": true, "tyto": true, "ktery": true,
	"ktera": true, "ktere": true, "nebo": true, "ani": true, "kdy": true, "kde": true,
	"proc": true, "jiz": true, "jen": true, "aby": true, "vsak": true, "tedy": true,
	"tim": true, "tom": true, "vse": true, "mui": true, "me": true, "mi": true, "ma": true,
	"the": true, "and": true, "for": true, "are": true, "but": true, "not": true,
}

// calculateRelevanceScore calculates how relevant an article is to the query
func calculateRelevanceScore(titleNormalized, contentNormalized, queryNormalized string, queryWords []string, issueDate string) int {
	score := 0

	// Exact phrase match in title (highest priority)
	if strings.Contains(titleNormalized, queryNormalized) {
		score += 100
	}

	// Exact phrase match in content
	if strings.Contains(contentNormalized, queryNormalized) {
		score += 20
	}

	// Filter out stop words and short words
	significantWords := filterSignificantWords(queryWords)

	// Individual word matches
	matchedWords := 0
	for _, word := range significantWords {
		titleMatch := strings.Contains(titleNormalized, word)
		contentMatch := strings.Contains(contentNormalized, word)

		if titleMatch {
			score += 15
			matchedWords++
		}
		if contentMatch {
			score += 3
			if !titleMatch {
				matchedWords++
			}
		}
	}

	// Bonus for matching multiple words (coverage bonus)
	if len(significantWords) > 1 && matchedWords > 1 {
		coverageRatio := float64(matchedWords) / float64(len(significantWords))
		score += int(coverageRatio * 25)
	}

	// Recency bonus - newer articles are more relevant
	score += calculateRecencyBonus(issueDate)

	return score
}

// filterSignificantWords removes stop words and short words from the query
func filterSignificantWords(words []string) []string {
	var result []string
	for _, word := range words {
		if len(word) >= 3 && !czechStopWords[word] {
			result = append(result, word)
		}
	}
	return result
}

// calculateRecencyBonus returns a bonus score based on how recent the article is
func calculateRecencyBonus(issueDate string) int {
	date, err := time.Parse("2006-01-02", issueDate)
	if err != nil {
		return 0
	}

	daysSince := int(time.Since(date).Hours() / 24)

	switch {
	case daysSince <= 30:
		return 50 // Last month
	case daysSince <= 90:
		return 30 // Last 3 months
	case daysSince <= 180:
		return 15 // Last 6 months
	case daysSince <= 365:
		return 5 // Last year
	default:
		return 0
	}
}

// normalizeText removes diacritics and converts to lowercase for search comparison
func normalizeText(s string) string {
	normalized, err := diacritics.Remove(s)
	if err != nil {
		// fallback to original string if diacritics removal fails
		normalized = s
	}
	return strings.ToLower(normalized)
}

// ProvideMunicipalNewsletterArticle returns full article content by ID
func ProvideMunicipalNewsletterArticle(articleID string) (string, error) {
	issues, err := fetchMunicipalNewsletterData()
	if err != nil {
		return "", err
	}

	for _, issue := range issues {
		for _, article := range issue.Articles {
			if article.ID == articleID {
				type FullArticle struct {
					Title      string `json:"title"`
					Author     string `json:"author"`
					Content    string `json:"content"`
					IssueTitle string `json:"issue_title"`
					IssueDate  string `json:"issue_date"`
				}

				result := FullArticle{
					Title:      article.Title,
					Author:     article.Author,
					Content:    article.Content,
					IssueTitle: issue.Title,
					IssueDate:  issue.Date,
				}

				output, err := json.Marshal(result)
				if err != nil {
					return "", fmt.Errorf("could not marshal article: %w", err)
				}

				return string(output), nil
			}
		}
	}

	return "Article not found.", nil
}
