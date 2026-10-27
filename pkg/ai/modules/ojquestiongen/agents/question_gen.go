package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"zk-code-arena-server/pkg/ai/infra/llm"
	"zk-code-arena-server/pkg/ai/modules/ojquestiongen/prompts"
	runtimecontext "zk-code-arena-server/pkg/ai/runtime/context"

	"github.com/cloudwego/eino/schema"
)

const questionGenAgentType = "oj_question_gen_agent"

// QuestionFocus 题目生成侧重点（保留给未来使用）
type QuestionFocus struct {
	ID    string `json:"id"`
	Focus string `json:"focus"`
	Type  string `json:"type"`
}

// GenerationParams 生成参数
type GenerationParams struct {
	KnowledgePoints        []string
	Difficulty             string
	Count                  int
	TestcaseCount          int
	StandardLanguage       string
	AdditionalRequirements string
}

// GeneratedProblem 生成中的题目（Agent 内部使用）
// json tag 必须与 prompts 中的输出契约（snake_case）一致。
// 否则 json.Unmarshal 只能做大小写不敏感匹配，input_format 匹配不上 InputFormat，
// 字段会被静默丢弃，导致题目详情页的输入输出说明为空。
type GeneratedProblem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	InputFormat  string `json:"input_format"`
	OutputFormat string `json:"output_format"`
	Hint         string `json:"hint"`
	Difficulty   string `json:"difficulty"`
}

// TestCase 测试用例（Agent 内部使用）
type TestCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	IsSample       bool   `json:"is_sample"`
}

// QuestionGenAgent 题目生成 Agent
type QuestionGenAgent struct {
	chatModel    llm.ModelFactory
	systemPrompt string
}

// NewQuestionGenAgent 创建题目生成 Agent
func NewQuestionGenAgent(
	chatModel llm.ModelFactory,
	systemPrompt string,
) *QuestionGenAgent {
	if systemPrompt == "" {
		systemPrompt = prompts.DefaultQuestionGenInstruction
	}
	return &QuestionGenAgent{
		chatModel:    chatModel,
		systemPrompt: systemPrompt,
	}
}

// ChatModel 返回 LLM 模型工厂
func (a *QuestionGenAgent) ChatModel() llm.ModelFactory {
	return a.chatModel
}

// Type 返回 Agent 类型标识
func (a *QuestionGenAgent) Type() string {
	return questionGenAgentType
}

// Process 主入口 - 封装 LLM 调用 + 解析
// Coordinator 调用此方法获取生成结果
func (a *QuestionGenAgent) Process(
	ctx context.Context,
	params *GenerationParams,
) ([]*GeneratedProblem, error) {
	// 1. 构建消息
	messages := a.buildMessages(params)

	// 2. 调用 LLM
	resp, err := a.chatModel.Call(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	// 3. 解析结果
	problems, err := a.parseProblems(resp.Content, params)
	if err != nil {
		return nil, fmt.Errorf("解析题目失败: %w", err)
	}

	return problems, nil
}

// buildMessages 构建 LLM 消息列表
func (a *QuestionGenAgent) buildMessages(params *GenerationParams) []*schema.Message {
	promptData := &prompts.GenerationParamsData{
		KnowledgePoints:        params.KnowledgePoints,
		Difficulty:             params.Difficulty,
		Count:                  params.Count,
		TestcaseCount:          params.TestcaseCount,
		StandardLanguage:       params.StandardLanguage,
		AdditionalRequirements: params.AdditionalRequirements,
	}

	return []*schema.Message{
		schema.SystemMessage(a.systemPrompt),
		schema.UserMessage(prompts.BuildQuestionGenUserPrompt(promptData)),
	}
}

// parseProblems 解析 LLM 返回的题目 JSON
func (a *QuestionGenAgent) parseProblems(content string, params *GenerationParams) ([]*GeneratedProblem, error) {
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("无法提取 JSON")
	}

	var problems []*GeneratedProblem
	if err := json.Unmarshal([]byte(jsonContent), &problems); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	if len(problems) == 0 {
		return nil, fmt.Errorf("没有生成任何题目")
	}

	// 补充 ID 和 difficulty
	for i := range problems {
		if problems[i].ID == "" {
			problems[i].ID = fmt.Sprintf("gen_%d", i+1)
		}
		if problems[i].Difficulty == "" {
			problems[i].Difficulty = params.Difficulty
		}
	}

	return problems, nil
}

// BuildMessages 构建 LLM 消息列表（保留给未来聊天 Agent 使用）
func (a *QuestionGenAgent) BuildMessages(snapshot *runtimecontext.Snapshot, question string) []*schema.Message {
	count := 2
	if snapshot != nil {
		count += len(snapshot.Messages)
	}
	messages := make([]*schema.Message, 0, count)

	messages = append(messages, schema.SystemMessage(a.systemPrompt))

	if snapshot != nil {
		for _, item := range snapshot.Messages {
			switch item.Role {
			case "assistant":
				messages = append(messages, schema.AssistantMessage(item.Content, nil))
			default:
				messages = append(messages, schema.UserMessage(item.Content))
			}
		}
	}

	messages = append(messages, schema.UserMessage(question))
	return messages
}

// extractJSON 从 markdown 或裸字符串中提取 JSON
func extractJSON(content string) string {
	start := strings.Index(content, "[")
	if start == -1 {
		start = strings.Index(content, "{")
	}
	if start == -1 {
		return ""
	}

	end := strings.LastIndex(content, "]")
	if end == -1 || end < start {
		end = strings.LastIndex(content, "}")
	}
	if end == -1 || end < start {
		return ""
	}

	return strings.TrimSpace(content[start : end+1])
}
#fix(ai): 修复出题字段解析与沙箱配置，新增演示模式
