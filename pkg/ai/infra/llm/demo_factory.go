package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// 三个 Agent 的 system prompt 标识串，取自 prompts.Default*Instruction。
// 用子串而非全量比较，避免提示词微调后演示模式直接失效。
const (
	questionGenMarker  = "你是 OJ 编程题生成 Agent"
	standardCodeMarker = "你是 OJ 标准程序生成 Agent"
	testcaseGenMarker  = "你是 OJ 测试用例生成 Agent"
)

// DemoModelFactory 演示模式模型工厂。
//
// 启用条件：conf.Config.AI.DemoMode == true 且未配置任何 ApiKey。
// 一旦配置了真实 ApiKey，装配处会优先使用 ConfigModelFactory，演示模式自动让位。
//
// 职责边界（重要）：
//   - 第 1~3 阶段（生成题目 / 标准答案 / 测试用例）返回内置样例内容，
//     即本工厂扮演的是"大模型"这一环，产出是预设的，不是实时推理结果。
//   - 第 4 阶段（沙箱验证）不经过本工厂，由 Verifier 真实编译并运行标准答案，
//     逐用例比对输出，因此"验证通过"是真实结论，可以用来佐证整条链路可用。
type DemoModelFactory struct {
	// Latency 模拟单次模型调用耗时，让前端四个阶段的推进节奏接近真实体验。
	Latency time.Duration
}

// NewDemoModelFactory 创建演示模式模型工厂
func NewDemoModelFactory() *DemoModelFactory {
	return &DemoModelFactory{Latency: 900 * time.Millisecond}
}

// IsChatConfigured 演示模式始终视为已配置，以保证 OJQuestionGen 模块正常注册
func (f *DemoModelFactory) IsChatConfigured() bool {
	return true
}

// ChatConfig 返回演示模式配置（不含任何密钥）
func (f *DemoModelFactory) ChatConfig() ChatConfig {
	return ChatConfig{
		BaseURL: "demo://builtin-samples",
		APIKey:  "",
		Model:   "demo-sample",
	}
}

// NewToolCallingChatModel 演示模式不提供工具调用能力
func (f *DemoModelFactory) NewToolCallingChatModel(ctx context.Context) (einomodel.ToolCallingChatModel, error) {
	return nil, fmt.Errorf("演示模式不支持工具调用，请配置真实的大模型 ApiKey")
}

// Stream 将演示样例以单条消息的形式包装为流（OJQuestionGen 走 Call，不经过此方法）
func (f *DemoModelFactory) Stream(ctx context.Context, messages []*schema.Message, tools []*schema.ToolInfo) (*schema.StreamReader[*schema.Message], error) {
	msg, err := f.Call(ctx, messages)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

// Call 根据 system prompt 判定当前属于哪个阶段，返回对应的样例内容
func (f *DemoModelFactory) Call(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	sys, user := splitDemoMessages(messages)

	content, err := demoRespond(sys, user)
	if err != nil {
		return nil, err
	}

	// 模拟真实模型调用的耗时；同时响应 ctx 取消，避免请求断开后空等
	select {
	case <-time.After(f.Latency):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return schema.AssistantMessage(content, nil), nil
}

// splitDemoMessages 拆出 system 与 user 文本
func splitDemoMessages(messages []*schema.Message) (sys, user string) {
	parts := make([]string, 0, len(messages))
	for _, m := range messages {
		if m == nil {
			continue
		}
		switch m.Role {
		case schema.System:
			if sys == "" {
				sys = m.Content
			}
		case schema.User:
			parts = append(parts, m.Content)
		}
	}
	return sys, strings.Join(parts, "\n")
}

// demoRespond 按阶段分派
func demoRespond(sys, user string) (string, error) {
	switch {
	case strings.Contains(sys, questionGenMarker):
		return demoQuestionJSON(user)
	case strings.Contains(sys, standardCodeMarker):
		return demoStandardCode(user)
	case strings.Contains(sys, testcaseGenMarker):
		return demoTestcaseJSON(user)
	default:
		return "", fmt.Errorf("演示模式无法识别本次模型调用：system prompt 未匹配任何已知 Agent")
	}
}

// demoProblem 样例题面结构（json tag 需与 prompts 输出契约一致）
type demoProblem struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	InputFormat  string `json:"input_format"`
	OutputFormat string `json:"output_format"`
	Hint         string `json:"hint"`
	Difficulty   string `json:"difficulty"`
}

// demoCase 样例测试用例结构
type demoCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	IsSample       bool   `json:"is_sample"`
}

// demoScenario 一道内置样例题：题面 + 标准答案 + 测试用例，三者必须自洽
type demoScenario struct {
	keywords []string
	problem  demoProblem
	code     map[string]string // 语言 -> 标准答案源码
	cases    []demoCase
}

// demoScenarios 内置样例题库。按知识点关键词命中，未命中则用第 1 套。
// 每套的测试用例期望值均由标准答案真实推导，确保第 4 阶段沙箱验证可通过。
var demoScenarios = []demoScenario{
	{
		keywords: []string{"数组", "循环", "哈希", "字符串", "枚举", "模拟"},
		problem: demoProblem{
			Title:       "校园图书角借书统计",
			Description: "仲恺农业工程学院某栋教学楼的图书角记录了连续 n 天每天借出的图书数量。现在给定这 n 天的记录，请你计算这 n 天的总借出量，以及日均借出量（向下取整）。",
			InputFormat: "第一行输入一个整数 n（1 ≤ n ≤ 1000），表示记录的天数。第二行输入 n 个整数，表示每天借出的图书数量。",
			// 输出格式：总和 + 空格 + 整除得到的日均量
			OutputFormat: "输出一行，包含两个整数：总借出量和日均借出量（向下取整），中间用一个空格分隔。",
			Hint:         "遍历数组累加求和得到总借出量，再用总借出量整除 n 得到日均借出量。题目保证 n ≥ 1，无需处理除零情况。",
			Difficulty:   "easy",
		},
		code: map[string]string{
			"java": `import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int n = sc.nextInt();
        long sum = 0;
        for (int i = 0; i < n; i++) {
            sum += sc.nextInt();
        }
        System.out.println(sum + " " + (sum / n));
    }
}
`,
			"python": `import sys


def main():
    data = sys.stdin.read().split()
    if not data:
        return
    n = int(data[0])
    nums = [int(x) for x in data[1:1 + n]]
    total = sum(nums)
    print(total, total // n)


if __name__ == "__main__":
    main()
`,
			"cpp": `#include <iostream>
using namespace std;

int main() {
    int n;
    if (!(cin >> n)) return 0;
    long long sum = 0;
    for (int i = 0; i < n; i++) {
        long long x;
        cin >> x;
        sum += x;
    }
    cout << sum << " " << sum / n << endl;
    return 0;
}
`,
		},
		cases: []demoCase{
			{Input: "5\n1 2 3 4 5\n", ExpectedOutput: "15 3", IsSample: true},
			{Input: "1\n0\n", ExpectedOutput: "0 0", IsSample: false},
			{Input: "3\n10 20 25\n", ExpectedOutput: "55 18", IsSample: false},
		},
	},
	{
		keywords: []string{"动态规划", "dp", "递推", "背包"},
		problem: demoProblem{
			Title:        "教学楼台阶走法统计",
			Description:  "学校教学楼有一段共 n 级台阶的楼梯，同学每次可以选择向上走 1 级或 2 级台阶。请你计算一共有多少种不同的走法可以走完这段楼梯。",
			InputFormat:  "第一行输入一个整数 n（1 ≤ n ≤ 30），表示台阶的总级数。",
			OutputFormat: "输出一个整数，表示走完这段楼梯的不同走法总数。",
			Hint:         "设 f(n) 为走完 n 级台阶的走法数，最后一步要么从第 n-1 级上 1 级，要么从第 n-2 级上 2 级，因此 f(n) = f(n-1) + f(n-2)，边界为 f(1)=1、f(2)=2，用递推即可。",
			Difficulty:   "medium",
		},
		code: map[string]string{
			"java": `import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int n = sc.nextInt();
        if (n <= 2) {
            System.out.println(n);
            return;
        }
        long prev = 1;
        long cur = 2;
        for (int i = 3; i <= n; i++) {
            long next = prev + cur;
            prev = cur;
            cur = next;
        }
        System.out.println(cur);
    }
}
`,
			"python": `import sys


def main():
    data = sys.stdin.read().split()
    if not data:
        return
    n = int(data[0])
    if n <= 2:
        print(n)
        return
    prev, cur = 1, 2
    for _ in range(3, n + 1):
        prev, cur = cur, prev + cur
    print(cur)


if __name__ == "__main__":
    main()
`,
			"cpp": `#include <iostream>
using namespace std;

int main() {
    int n;
    if (!(cin >> n)) return 0;
    if (n <= 2) {
        cout << n << endl;
        return 0;
    }
    long long prev = 1, cur = 2;
    for (int i = 3; i <= n; i++) {
        long long next = prev + cur;
        prev = cur;
        cur = next;
    }
    cout << cur << endl;
    return 0;
}
`,
		},
		cases: []demoCase{
			{Input: "5\n", ExpectedOutput: "8", IsSample: true},
			{Input: "1\n", ExpectedOutput: "1", IsSample: false},
			{Input: "10\n", ExpectedOutput: "89", IsSample: false},
		},
	},
	{
		keywords: []string{"排序", "二分", "查找", "贪心"},
		problem: demoProblem{
			Title:        "期末成绩升序整理",
			Description:  "期末成绩系统录入了 n 名同学的程序设计课程成绩。为了方便统计，请将这些成绩按从低到高的顺序整理后输出。",
			InputFormat:  "第一行输入一个整数 n（1 ≤ n ≤ 1000），表示成绩条数。第二行输入 n 个整数，表示各同学的成绩。",
			OutputFormat: "输出一行，包含 n 个整数，为升序排列后的成绩，中间用一个空格分隔。",
			Hint:         "使用语言内置的排序函数即可，时间复杂度 O(n log n) 完全满足要求，注意输出时数字之间用单个空格分隔且行末无多余空格。",
			Difficulty:   "easy",
		},
		code: map[string]string{
			"java": `import java.util.Arrays;
import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int n = sc.nextInt();
        int[] scores = new int[n];
        for (int i = 0; i < n; i++) {
            scores[i] = sc.nextInt();
        }
        Arrays.sort(scores);

        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < n; i++) {
            if (i > 0) {
                sb.append(' ');
            }
            sb.append(scores[i]);
        }
        System.out.println(sb);
    }
}
`,
			"python": `import sys


def main():
    data = sys.stdin.read().split()
    if not data:
        return
    n = int(data[0])
    nums = [int(x) for x in data[1:1 + n]]
    nums.sort()
    print(" ".join(str(x) for x in nums))


if __name__ == "__main__":
    main()
`,
			"cpp": `#include <algorithm>
#include <iostream>
#include <vector>
using namespace std;

int main() {
    int n;
    if (!(cin >> n)) return 0;
    vector<int> a(n);
    for (int i = 0; i < n; i++) cin >> a[i];
    sort(a.begin(), a.end());
    for (int i = 0; i < n; i++) {
        if (i > 0) cout << ' ';
        cout << a[i];
    }
    cout << endl;
    return 0;
}
`,
		},
		cases: []demoCase{
			{Input: "5\n88 70 95 61 73\n", ExpectedOutput: "61 70 73 88 95", IsSample: true},
			{Input: "1\n50\n", ExpectedOutput: "50", IsSample: false},
			{Input: "3\n100 100 1\n", ExpectedOutput: "1 100 100", IsSample: false},
		},
	},
}

// demoPickScenarioIndex 按请求的知识点挑选最贴合的内置样例下标，未命中则回落到第 1 套
func demoPickScenarioIndex(points string) int {
	for i := range demoScenarios {
		for _, kw := range demoScenarios[i].keywords {
			if strings.Contains(points, kw) {
				return i
			}
		}
	}
	return 0
}

// demoQuestionJSON 第 1 阶段：返回题面 JSON 数组
func demoQuestionJSON(user string) (string, error) {
	count := demoIntField(user, "question_count")
	if count <= 0 {
		count = 1
	}
	if count > len(demoScenarios) {
		count = len(demoScenarios)
	}

	difficulty := strings.TrimSpace(demoStringField(user, "difficulty"))
	if difficulty == "" {
		difficulty = "easy"
	}

	// 从最贴合知识点诉求的样例开始取，避免「要动态规划却返回一道求和题」
	start := demoPickScenarioIndex(demoStringField(user, "knowledge_points"))

	items := make([]demoProblem, 0, count)
	for i := 0; i < count; i++ {
		p := demoScenarios[(start+i)%len(demoScenarios)].problem
		p.Difficulty = difficulty
		items = append(items, p)
	}

	encoded, err := json.Marshal(items)
	if err != nil {
		return "", fmt.Errorf("序列化样例题目失败: %w", err)
	}
	return string(encoded), nil
}

// demoStandardCode 第 2 阶段：返回标准答案源码（纯文本，符合输出契约）
func demoStandardCode(user string) (string, error) {
	scenario, ok := demoMatchScenario(user)
	if !ok {
		return "", fmt.Errorf("演示模式：请求中未匹配到内置样例题目")
	}

	language := strings.ToLower(strings.TrimSpace(demoStringField(user, "target_language")))
	if code, exists := scenario.code[language]; exists {
		return code, nil
	}
	// 未预置的语言回落到 Java 实现，保证演示链路不会中断
	return scenario.code["java"], nil
}

// demoTestcaseJSON 第 3 阶段：返回测试用例 JSON 数组
func demoTestcaseJSON(user string) (string, error) {
	scenario, ok := demoMatchScenario(user)
	if !ok {
		return "", fmt.Errorf("演示模式：请求中未匹配到内置样例题目")
	}

	encoded, err := json.Marshal(scenario.cases)
	if err != nil {
		return "", fmt.Errorf("序列化样例测试用例失败: %w", err)
	}
	return string(encoded), nil
}

// demoMatchScenario 用题目标题反查内置样例（第 2、3 阶段的 user prompt 都带 <title>）
func demoMatchScenario(user string) (*demoScenario, bool) {
	title := strings.TrimSpace(demoStringField(user, "title"))
	if title == "" {
		return &demoScenarios[0], true
	}
	for i := range demoScenarios {
		if demoScenarios[i].problem.Title == title {
			return &demoScenarios[i], true
		}
		for _, kw := range demoScenarios[i].keywords {
			if strings.Contains(title, kw) {
				return &demoScenarios[i], true
			}
		}
	}
	// 标题无法识别时给默认样例，保证链路不中断
	return &demoScenarios[0], true
}

// demoStringField 取出 <tag>\n...\n</tag> 区块中的文本内容。
// 提示词由 prompts.writeTaggedBlock 生成，格式固定为 "<tag>\n内容\n</tag>\n\n"。
// 这里用普通字符串查找而非正则：Go 的 RE2 不支持反向引用。
func demoStringField(user, tag string) string {
	open := "<" + tag + ">\n"
	closed := "\n</" + tag + ">"

	start := strings.Index(user, open)
	if start == -1 {
		return ""
	}
	start += len(open)

	rest := user[start:]
	end := strings.Index(rest, closed)
	if end == -1 {
		return ""
	}
	return rest[:end]
}

// demoIntField 取出指定标签的整数内容
func demoIntField(user, tag string) int {
	raw := strings.TrimSpace(demoStringField(user, tag))
	if raw == "" || raw == "(无)" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return value
}
