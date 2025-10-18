package dictionary

import (
	"math/rand"
	"strings"

	"github.com/nospi/dictionary/internal/game"
)

// MockDictionary is a mock implementation for testing
type MockDictionary struct {
	words []game.Word
	used  map[string]bool // Track used words to avoid repetition
}

// NewMockDictionary creates a new mock dictionary service
func NewMockDictionary() *MockDictionary {
	return &MockDictionary{
		words: mockWords,
		used:  make(map[string]bool),
	}
}

// GetRandomWord returns a random word from the mock list
func (m *MockDictionary) GetRandomWord() (game.Word, error) {
	if len(m.words) == 0 {
		return game.Word{}, ErrWordNotFound
	}

	// Try to get an unused word
	attempts := 0
	maxAttempts := len(m.words) * 2

	for attempts < maxAttempts {
		idx := rand.Intn(len(m.words))
		word := m.words[idx]

		if !m.used[word.Text] {
			m.used[word.Text] = true
			return word, nil
		}

		attempts++
	}

	// If all words have been used, reset and return a random one
	m.used = make(map[string]bool)
	idx := rand.Intn(len(m.words))
	word := m.words[idx]
	m.used[word.Text] = true

	return word, nil
}

// GetWordDefinition returns a specific word's definition
func (m *MockDictionary) GetWordDefinition(word string) (game.Word, error) {
	word = strings.ToLower(strings.TrimSpace(word))

	for _, w := range m.words {
		if strings.ToLower(w.Text) == word {
			return w, nil
		}
	}

	return game.Word{}, ErrWordNotFound
}

// Reset clears the used words tracker
func (m *MockDictionary) Reset() {
	m.used = make(map[string]bool)
}
