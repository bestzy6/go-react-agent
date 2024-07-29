package agent

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"go-react-agent/tools"
	"go-react-agent/utils"
	"strings"
	"text/template"
)

const (
	_ModelName   = "gpt-4-turbo"
	_Temperature = 0.8

	_MaxIteration = 8
)

//go:embed prompt/react_agent.md
var _ReActSystemPrompt string

type ReActAgent struct {
	llm          *openai.Client
	tools        map[string]tools.Tool
	systemPrompt string
}

func NewReActAgent(llm *openai.Client, functionTool ...tools.Tool) *ReActAgent {
	tpl := template.Must(
		template.New("system").Option("missingkey=default").Parse(_ReActSystemPrompt),
	)

	var (
		builder  strings.Builder
		toolMap  = make(map[string]tools.Tool, len(functionTool))
		toolName = make([]string, 0, len(functionTool))
	)

	// 构建toolDesc
	for i := range functionTool {
		t := functionTool[i]
		builder.WriteString("- ToolName: " + t.Name() + "\n")
		builder.WriteString("    - Description: " + t.Description() + "\n")
		builder.WriteString("    - Args Description: " + t.ArgsDescription() + "\n")

		toolMap[t.Name()] = t
		toolName = append(toolName, t.Name())
	}
	toolDesc := builder.String()

	// 重置builder
	builder.Reset()
	if err := tpl.Execute(&builder, map[string]string{
		"Background": "", // custom
		"ToolDesc":   toolDesc,
		"ToolNames":  strings.Join(toolName, " , "),
	}); err != nil {
		panic(err)
	}

	agent := &ReActAgent{
		llm:          llm,
		tools:        toolMap,
		systemPrompt: builder.String(),
	}

	fmt.Println(agent.systemPrompt)

	return agent
}

func (r *ReActAgent) Call(ctx context.Context, query string) (string, error) {
	msg := r.appendMsg(openai.ChatMessageRoleUser, "Question: "+query+"\n", r.initMsg())

	response, err := r.llm.CreateChatCompletion(ctx, msg)
	if err != nil {
		return "", err
	}

	// 迭代N次，直到出现结果
	for i := 0; i < _MaxIteration; i++ {
		output := response.Choices[0].Message.Content

		switch {
		case strings.Contains(output, "Action:"): // Function Call
			use, err := utils.ExtractToolUse(output)
			if err != nil {
				// 添加错误信息
				msg = r.appendMsg(openai.ChatMessageRoleUser, "Error: 请按照正确的格式回答\n", msg)
				break
			}

			assistantPrompt := fmt.Sprintf("Thought: %s\nAction: %s\nAction Input: %s", use.Thought, use.Action, use.ActionInput)
			fmt.Println(assistantPrompt)
			msg = r.appendMsg(openai.ChatMessageRoleAssistant, assistantPrompt, msg)

			// 找到工具
			tool, ok := r.tools[strings.TrimSpace(use.Action)]
			if !ok {
				msg = r.appendMsg(openai.ChatMessageRoleUser, "Error: 不要选择不存在的Tool，请仔细思考\n", msg)
				break
			}

			result, xError := tool.Exec(ctx, use.ActionInput)
			if xError != nil {
				// 告诉LLM发生了错误，重来一遍
				msg = r.appendMsg(openai.ChatMessageRoleUser, "Error: 发生了错误，请仔细思考！错误细节："+xError.Error()+"\n", msg)
				break
			}

			// 成功执行工具
			msg = r.appendMsg(openai.ChatMessageRoleUser, "Observation: "+result+"\n", msg)
			fmt.Println("Observation: " + result + "\n")

		case strings.Contains(output, "Answer:"): // Final Answer
			answer, err := utils.ExtractFinalAnswer(output)
			if err != nil {
				// 添加错误信息
				msg = r.appendMsg(openai.ChatMessageRoleUser, "Error: 请按照正确的格式回答\n", msg)
				break
			}

			fmt.Println(fmt.Sprintf("Thought: %s\nAnswer: %s\n", answer.Thought, answer.Answer))
			return answer.Answer, nil

		default: // 没有解析出Function Call 或 Final Answer
			msg = r.appendMsg(openai.ChatMessageRoleUser, "Error: 请按照正确的格式回答\n", msg)
		}

		// 发送新的消息
		response, err = r.llm.CreateChatCompletion(ctx, msg)
		if err != nil {
			return "", err
		}
	}
	return "", errors.New("exceeding the maximum iteration")
}

func (r *ReActAgent) initMsg() openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: _ModelName,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: r.systemPrompt,
			},
		},
		Temperature: _Temperature,
	}
}

func (r *ReActAgent) appendMsg(role, content string, history openai.ChatCompletionRequest) openai.ChatCompletionRequest {
	msgs := append(history.Messages, openai.ChatCompletionMessage{
		Role:    role,
		Content: content,
	})
	return openai.ChatCompletionRequest{
		Model:       _ModelName,
		Messages:    msgs,
		Temperature: _Temperature,
	}
}
