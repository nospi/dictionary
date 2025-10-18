package dictionary

import (
	"errors"

	"github.com/nospi/dictionary/internal/game"
)

// Service defines the interface for dictionary operations
type Service interface {
	GetRandomWord() (game.Word, error)
	GetWordDefinition(word string) (game.Word, error)
}

var ErrWordNotFound = errors.New("word not found")

// Mock word list for testing
var mockWords = []game.Word{
	{
		Text:       "scurryfunge",
		Definition: "A hasty tidying of the house between the time you see a neighbor and the time she knocks on the door",
		Source:     "mock",
	},
	{
		Text:       "borborygmus",
		Definition: "A rumbling or gurgling noise made by the movement of fluid and gas in the intestines",
		Source:     "mock",
	},
	{
		Text:       "nudiustertian",
		Definition: "Relating to the day before yesterday",
		Source:     "mock",
	},
	{
		Text:       "kakorrhaphiophobia",
		Definition: "An abnormal fear of failure or defeat",
		Source:     "mock",
	},
	{
		Text:       "floccinaucinihilipilification",
		Definition: "The action or habit of estimating something as worthless",
		Source:     "mock",
	},
	{
		Text:       "snickersnee",
		Definition: "A large knife designed for use as a weapon",
		Source:     "mock",
	},
	{
		Text:       "fartlek",
		Definition: "A training technique, used especially among runners, consisting of bursts of intense effort loosely alternating with less strenuous activity",
		Source:     "mock",
	},
	{
		Text:       "gobemouche",
		Definition: "A gullible or credulous person",
		Source:     "mock",
	},
	{
		Text:       "limerence",
		Definition: "The state of being infatuated or obsessed with another person",
		Source:     "mock",
	},
	{
		Text:       "nudiustertian",
		Definition: "Of or relating to the day before yesterday",
		Source:     "mock",
	},
	{
		Text:       "petrichor",
		Definition: "A pleasant smell that frequently accompanies the first rain after a long period of warm, dry weather",
		Source:     "mock",
	},
	{
		Text:       "taradiddle",
		Definition: "A petty lie or pretentious nonsense",
		Source:     "mock",
	},
	{
		Text:       "widdershins",
		Definition: "In a direction contrary to the sun's course, considered as unlucky; anticlockwise",
		Source:     "mock",
	},
	{
		Text:       "smellfungus",
		Definition: "An excessively faultfinding person",
		Source:     "mock",
	},
	{
		Text:       "rawgabbit",
		Definition: "A person who speaks confidently on a subject they know little about",
		Source:     "mock",
	},
	{
		Text:       "mumpsimus",
		Definition: "A person who obstinately adheres to old ways in spite of clear evidence they are wrong",
		Source:     "mock",
	},
	{
		Text:       "callipygian",
		Definition: "Having well-shaped buttocks",
		Source:     "mock",
	},
	{
		Text:       "fudgel",
		Definition: "Pretending to work when you're not actually doing anything",
		Source:     "mock",
	},
	{
		Text:       "erinaceous",
		Definition: "Of, pertaining to, or resembling a hedgehog",
		Source:     "mock",
	},
	{
		Text:       "lollygag",
		Definition: "To spend time aimlessly; to dawdle",
		Source:     "mock",
	},
	{
		Text:       "bumfuzzle",
		Definition: "To confuse or fluster",
		Source:     "mock",
	},
	{
		Text:       "sialoquent",
		Definition: "Spraying saliva while speaking",
		Source:     "mock",
	},
	{
		Text:       "ultracrepidarian",
		Definition: "A person who gives opinions on matters beyond their knowledge",
		Source:     "mock",
	},
	{
		Text:       "pneumonoultramicroscopicsilicovolcanoconiosis",
		Definition: "A lung disease caused by inhaling very fine ash and sand dust",
		Source:     "mock",
	},
	{
		Text:       "bumbershoot",
		Definition: "An umbrella",
		Source:     "mock",
	},
	{
		Text:       "gardyloo",
		Definition: "A warning cry given before throwing waste water out of a window onto the street below",
		Source:     "mock",
	},
	{
		Text:       "collywobbles",
		Definition: "A feeling of fear, apprehension, or nervousness; butterflies in the stomach",
		Source:     "mock",
	},
	{
		Text:       "biblioklept",
		Definition: "A person who steals books",
		Source:     "mock",
	},
	{
		Text:       "snollygoster",
		Definition: "A shrewd, unprincipled person, especially a politician",
		Source:     "mock",
	},
	{
		Text:       "absquatulate",
		Definition: "To leave abruptly or hastily",
		Source:     "mock",
	},
}
