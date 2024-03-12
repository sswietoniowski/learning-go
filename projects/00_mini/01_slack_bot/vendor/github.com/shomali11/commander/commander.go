package commander

import (
	"regexp"
	"strings"

	"github.com/shomali11/proper"
)

const (
	escapeCharacter          = "\\"
	ignoreCase               = "(?i)"
	wordParameterPattern     = "{\\S+}"
	sentenceParameterPattern = "<\\S+>"
	spacePattern             = "\\s+"
	wordInputPattern         = "(.+?)"
	sentenceInputPattern     = "(.+)"
	preCommandPattern        = "(\\s|^)"
	postCommandPattern       = "(\\s|$)"
)

const (
	notParameter      = "NOT_PARAMETER"
	wordParameter     = "WORD_PARAMETER"
	sentenceParameter = "SENTENCE_PARAMETER"
)

var (
	regexCharacters = []string{"\\", "(", ")", "{", "}", "[", "]", "?", ".", "+", "|", "^", "$"}
)

// NewCommand creates a new Command object from the format passed in
func NewCommand(format string) *Command {
	tokens := tokenize(format)
	expressions := generate(tokens)
	return &Command{tokens: tokens, expressions: expressions}
}

// Token represents the Token object
type Token struct {
	Word string
	Type string
}

func (t Token) IsParameter() bool {
	return t.Type != notParameter
}

// Command represents the Command object
type Command struct {
	tokens      []*Token
	expressions []*regexp.Regexp
}

// Match takes in the command and the text received, attempts to find the pattern and extract the parameters
func (c *Command) Match(text string) (*proper.Properties, bool) {
	if len(c.expressions) == 0 {
		return nil, false
	}

	for _, expression := range c.expressions {
		matches := expression.FindStringSubmatch(text)
		if len(matches) == 0 {
			continue
		}

		values := matches[2 : len(matches)-1]

		valueIndex := 0
		parameters := make(map[string]string)
		for i := 0; i < len(c.tokens) && valueIndex < len(values); i++ {
			token := c.tokens[i]
			if !token.IsParameter() {
				continue
			}

			parameters[token.Word] = values[valueIndex]
			valueIndex++
		}
		return proper.NewProperties(parameters), true
	}
	return nil, false
}

// Tokenize returns Command info as tokens
func (c *Command) Tokenize() []*Token {
	return c.tokens
}

func escape(text string) string {
	for _, character := range regexCharacters {
		text = strings.Replace(text, character, escapeCharacter+character, -1)
	}
	return text
}

func tokenize(format string) []*Token {
	parameterRegex := regexp.MustCompile(sentenceParameterPattern)
	lazyParameterRegex := regexp.MustCompile(wordParameterPattern)
	words := strings.Fields(format)
	tokens := make([]*Token, len(words))
	for i, word := range words {
		switch {
		case lazyParameterRegex.MatchString(word):
			tokens[i] = &Token{Word: word[1 : len(word)-1], Type: wordParameter}
		case parameterRegex.MatchString(word):
			tokens[i] = &Token{Word: word[1 : len(word)-1], Type: sentenceParameter}
		default:
			tokens[i] = &Token{Word: word, Type: notParameter}
		}
	}
	return tokens
}

func generate(tokens []*Token) []*regexp.Regexp {
	regexps := []*regexp.Regexp{}
	if len(tokens) == 0 {
		return regexps
	}

	for index := len(tokens) - 1; index >= -1; index-- {
		regex := compile(create(tokens, index))
		if regex != nil {
			regexps = append(regexps, regex)
		}
	}

	return regexps
}

func create(tokens []*Token, boundary int) []*Token {
	newTokens := []*Token{}
	for i := 0; i < len(tokens); i++ {
		if !tokens[i].IsParameter() || i <= boundary {
			newTokens = append(newTokens, tokens[i])
		}
	}
	return newTokens
}

func compile(tokens []*Token) *regexp.Regexp {
	if len(tokens) == 0 {
		return nil
	}

	pattern := preCommandPattern + getInputPattern(tokens[0])
	for index := 1; index < len(tokens); index++ {
		currentToken := tokens[index]
		pattern += spacePattern + getInputPattern(currentToken)
	}
	pattern += postCommandPattern

	return regexp.MustCompile(ignoreCase + pattern)
}

func getInputPattern(token *Token) string {
	switch token.Type {
	case wordParameter:
		return wordInputPattern
	case sentenceParameter:
		return sentenceInputPattern
	default:
		return escape(token.Word)
	}
}
