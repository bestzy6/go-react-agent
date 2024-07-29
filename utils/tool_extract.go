package utils

import (
	"errors"
	"regexp"
)

const _ToolUsePattern = `^\s*Thought: (.*?)\n+Action: ([a-zA-Z0-9_]+).*?\n+Action Input: .*?(\{.*\})`

var re = regexp.MustCompile(_ToolUsePattern)

type ToolUse struct {
	Thought     string
	Action      string
	ActionInput string
}

func ExtractToolUse(text string) (ToolUse, error) {
	matches := re.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return ToolUse{}, errors.New("Could not extract tools use from input text")
	}

	// Only take the first matched text
	match := matches[0]
	if len(match) != 4 {
		return ToolUse{}, errors.New("Could not extract tools use from input text")
	}

	thought := match[1]
	action := match[2]
	actionInputStr := match[3]
	return ToolUse{Thought: thought, Action: action, ActionInput: actionInputStr}, nil
}
