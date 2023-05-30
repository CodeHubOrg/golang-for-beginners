package main

import (
	"fmt"
	"math/rand"
)

const (
	order      = 5
	iterations = 500
	startText  = "This beautiful morning will"
)

func main() {
	// Some completely random text to provide a basis for analysis
	inputTxt := `Today is a beautiful day.
Tomorrow might be even more beautiful, if the rain stops finally.
Nevermind the hailstorm, it's a wonderful day out nonetheless.
The will of the people is to plant willows and sleep on pillows.
I'm willing to row if marmaids don't follow the boat in the morning sunrise.
I'm cheerful and full of optimism that our ally will arrive in the midday sun.
How about the hermit though, it's always so unpredictable to torment my soul.
`

	// Process input text and store state transition probabilities
	ngrams := processText(inputTxt, order)
	// Generate new text based on statistical model built in previous step
	result := generateText(startText, order, iterations, ngrams)
	// Print out the result
	fmt.Println(result)
}

// processText finds every order-long character sequence in the given input text
// and returns a map of every such character sequence and their corresponding slice
// of every occurrence of immediately following characters.
func processText(input string, order int) map[string][]rune {
	ngrams := make(map[string][]rune)
	inputRunes := []rune(input)
	inputRunesLen := len(inputRunes)

	// Iterate from the start of the input text all the way to the end
	// respecting the length of the character sequences we're looking for (order)
	for i := 0; i < inputRunesLen-order; i++ {
		// Character sequence from current position (i -> i+order)
		gram := string(inputRunes[i : i+order])
		// If we haven't come across the current character sequence
		// create an entry for it in our map
		if ngrams[gram] == nil {
			ngrams[gram] = []rune{}
		}
		// Append the character immediately following our character sequence
		// to the list of every occurrence of following charaters
		ngrams[gram] = append(ngrams[gram], []rune(inputRunes)[i+order])
	}
	return ngrams
}

// generateText returns a procedurally generated text starting with startTxt
// and characters appended to it from the provided ngrams map (outputs of processText)
// which contains all the possible order-length character sequences
// and their possible follow-up characters with their respective probabilities.
// The number of characters generated and appended to startTxt is defined by iterations.
// If the algorithm encounters a charater sequence that has no follow-up character
// provided in the ngram map before the number of iterations is reached
// the function will return with the text generated up to that point.
func generateText(startTxt string, order int, iterations int, ngrams map[string][]rune) string {
	// Convert start text to runes
	startRunes := []rune(startTxt)
	startRunesLen := len(startRunes)

	// If start text is shorter than the pattern length based on which
	// we look for the probability of next characters, we bail as
	// we won't be able find any matching patterns and their possible next characters
	if startRunesLen < order {
		return startTxt
	}

	// The starting pattern is the last order-length character sequence in the start text
	currentGram := startRunes[startRunesLen-order:]
	// The final result will start with the start text (converted to characters)
	// to which we'll append any new characters to generate the full output
	result := startRunes

	// Look for next patters iterations-times
	// or if encountered a pattern that has zero possible follow-up characters
	for i := 0; i < iterations; i++ {
		// Declare storage possible follow-up characters here
		var possibilities []rune
		possibilities = *new([]rune)
		// Look up every possible follow-up charaters for the current pattern
		if p, ok := ngrams[string(currentGram)]; !ok {
			// None found, we can't continue
			break
		} else {
			// Store possible follow-up characters
			possibilities = p
		}
		// Pick one possible follow-up character randomly
		nextRune := possibilities[rand.Intn(len(possibilities))]
		// Append it to the result
		result = append(result, nextRune)
		// Change the current pattern
		resultLen := len(result)
		currentGram = result[resultLen-order : resultLen]
	}
	return string(result)
}
