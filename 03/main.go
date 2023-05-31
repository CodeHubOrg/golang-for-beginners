package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	defaultStartText  = "This beautiful morning will"
	defaultOrder      = 5
	defaultIterations = 500
	defaultPort       = 8080
	orderHelpTxt      = `[o]rder:
The number of past states based on which future states are determined.
That is, when looking for character patters, based on how long a sequence
of characters to look for probability of the next character appearing in the sample text.`
	iterationsHelpTxt = `[i]terations:
When generating text, how many times to append a newly generated state to the current one.
That is, how many extra characters to add to the starting text.
`
	portHelpTxt = `[p]ort:
The port the HTTP server will listen on for requests.
`
)

var (
	startText  = defaultStartText
	order      = defaultOrder
	iterations = defaultIterations
	port       = defaultPort
)

func main() {
	// Read command line arguments into pointers
	orderPtr := flag.Int("o", defaultOrder, orderHelpTxt)          // default -o=5
	iterPtr := flag.Int("i", defaultIterations, iterationsHelpTxt) // default -i=500
	portPtr := flag.Int("p", defaultPort, iterationsHelpTxt)       // default -i=8080
	flag.Parse()

	// Dereference pointers to read values
	order = *orderPtr
	iterations = *iterPtr
	port = *portPtr

	// Set up webserver handlers for specific URL patterns
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/markov/", markovHandler)

	// Start webserver
	addr := fmt.Sprintf(":%v", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// homeHandler prints "Welcome!" in the browser if called on the root directory
// Returns
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	fmt.Fprint(w, "Welcome!")
}

func markovHandler(w http.ResponseWriter, r *http.Request) {
	// If URL doesn't match return not found
	if len(r.URL.Path) < 8 || r.URL.Path[:8] != "/markov/" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	// If author name is zero length return not found
	author := r.URL.Path[8:]
	if len(author) == 0 {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	// Read o query parameter, if sent, for order
	// If not sent, order is already set to default value
	params := r.URL.Query()
	if params.Has("o") {
		o, err := strconv.Atoi(params.Get("o"))
		if err == nil {
			order = o
		}
	}

	// Read i query parameter for iterations
	// If not sent, iterations is already set to default value
	if params.Has("i") {
		i, err := strconv.Atoi(params.Get("i"))
		if err == nil {
			iterations = i
		}
	}

	// Treat additional command line arguments, if any,
	// as starting text for the text we're about to generate
	if params.Has("s") {
		startText = params.Get("s")
	}

	// Load input text based on which state transitions will be determined
	inputTxt, err := loadText(author)
	if err != nil {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	// Process input text and store state transition probabilities
	ngrams := processText(inputTxt, order)
	// Generate new text based on statistical model built in previous step
	result := generateText(startText, order, iterations, ngrams)
	// Print out the result to the browser
	fmt.Fprint(w, result)
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "Error 404\nPage not found")
	}
}

// loadText returns a sample book from the given author as a rune slice that,
// in turn, can be used to provide a statistical basis for newly generated text
func loadText(author string) (string, error) {
	a := strings.ToLower(author)
	var bookTitleTxt string
	switch a {
	case "carroll":
		bookTitleTxt = "alice_in_wonderland"
	case "forster":
		bookTitleTxt = "a_room_with_a_view"
	case "shakespeare":
		bookTitleTxt = "the_tragedie_of_macbeth"
	default:
		return "", errors.New("author not found")
	}

	f := fmt.Sprintf("data/%v.txt", bookTitleTxt)
	b, err := os.ReadFile(f)
	if err != nil {
		fmt.Print(err)
	}
	return string(b), nil
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
