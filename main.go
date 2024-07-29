package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"go-react-agent/agent"
	"log"
	"os"
)

type Sum struct {
}

func (a *Sum) Name() string {
	return "Sum"
}

func (a *Sum) Description() string {
	return "Calculate the sum of two numbers"
}

func (a *Sum) ArgsDescription() string {
	return `Accept parameters n1, n2. e,g.  {"n1":0.0,"n2":2.0}`
}

func (a *Sum) Exec(ctx context.Context, input string) (string, error) {
	var m map[string]float64
	err := json.Unmarshal([]byte(input), &m)
	if err != nil {
		return "", err
	}
	v1, ok := m["n1"]
	if !ok {
		return "", errors.New("")
	}

	v2, ok := m["n2"]
	if !ok {
		return "", errors.New("")
	}

	sum := v1 + v2
	return fmt.Sprintf("%f", sum), nil
}

func main() {
	authToken := os.Getenv("OPENAI_API_KEY")
	config := openai.DefaultConfig(authToken)
	config.BaseURL = os.Getenv("OPENAI_API_BASE")
	llm := openai.NewClientWithConfig(config)

	reactAgent := agent.NewReActAgent(llm, &Sum{})
	call, err := reactAgent.Call(context.Background(), "1.23+3.21=?")
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(call)
}
