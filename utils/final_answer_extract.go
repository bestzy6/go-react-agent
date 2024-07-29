package utils

import (
	"errors"
	"regexp"
)

const _FinalAnswerPattern = `\s*Thought:(.*?)\n+Answer:.*?(.*)`

var reFinalAnswer = regexp.MustCompile(_FinalAnswerPattern)

type FinalAnswer struct {
	Thought string
	Answer  string
}

func ExtractFinalAnswer(text string) (FinalAnswer, error) {
	matches := reFinalAnswer.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return FinalAnswer{}, errors.New("Could not extract tools use from input text")
	}

	match := matches[0]
	if len(match) != 3 {
		return FinalAnswer{}, errors.New("Could not extract tools use from input text")
	}

	return FinalAnswer{
		Thought: match[1],
		Answer:  match[2],
	}, nil
}
