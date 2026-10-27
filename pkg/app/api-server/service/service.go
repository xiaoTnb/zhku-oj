/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: service.go
@Description: 服务层入口
*/

package service

import (
	"context"
	"fmt"

	"zk-code-arena-server/conf"
	aiassistants "zk-code-arena-server/pkg/ai/assistants"
	aiwebsearch "zk-code-arena-server/pkg/ai/assistants/websearch"
	"zk-code-arena-server/pkg/ai/infra/llm"
	toolsearch "zk-code-arena-server/pkg/ai/infra/search"
	aimodules "zk-code-arena-server/pkg/ai/modules"
	ojquestiongen "zk-code-arena-server/pkg/ai/modules/ojquestiongen"
	runtimecontext "zk-code-arena-server/pkg/ai/runtime/context"
	aigateway "zk-code-arena-server/pkg/ai/runtime/gateway"
	runtimesession "zk-code-arena-server/pkg/ai/runtime/session"
	runtimetask "zk-code-arena-server/pkg/ai/runtime/task"
	"zk-code-arena-server/pkg/app/api-server/repository"
	"zk-code-arena-server/pkg/common/queue"
	"zk-code-arena-server/pkg/common/sandbox"
	"zk-code-arena-server/pkg/utils"
)

// Service 服务层结构
type Service struct {
	UserService                    *UserService
	ProblemService                 *ProblemService
	SubmitService                  *SubmitService
	JudgeService                   *JudgeService
	TestCaseService                *TestCaseService
	CourseService                  *CourseService
	ChoiceQuestionService          *ChoiceQuestionService
	BlankQuestionService           *BlankQuestionService
	CourseContentService           *CourseContentService
	ContentSubmissionService       *ContentSubmissionService
	StatisticsService              *StatisticsService
	ClazzService                   *ClazzService
	DailyProblemService            *DailyProblemService
	TeacherCopilot                 *TeacherCopilotService
	AIGeneratedProblemDraftService *AIGeneratedProblemDraftService
	AssistantGateway               aigateway.AssistantGateway
	ModuleGateway                  aigateway.ModuleGateway
	RSSService                     *RSSService
}

// NewService 创建服务实例
func NewService() *Service {
	// 1. 创建 Repository 容器
	repos := repository.NewRepositories()

	// 2. 创建基础服务（使用 Repository）
	clazzService := NewClazzService(repos.ClazzRepository)
	userService := NewUserService(clazzService)
	submitService := NewSubmitService(repos.SubmitRepository)
	testCaseService := NewTestCaseService(repos.TestCaseRepository)
	courseService := NewCourseService()
	choiceQuestionService := NewChoiceQuestionService(repos.ChoiceQuestionRepository, repos.ContentRelationRepository, courseService)
	blankQuestionService := NewBlankQuestionService(repos.BlankQuestionRepository, repos.ContentRelationRepository, courseService)
	courseContentService := NewCourseContentService(
		repos.ProblemRepository,
		repos.ChoiceQuestionRepository,
		repos.BlankQuestionRepository,
		repos.ContentRelationRepository,
		courseService,
	)
	contentSubmissionService := NewContentSubmissionService(
		repos.ContentSubmissionRepository,
		repos.ChoiceQuestionRepository,
		repos.BlankQuestionRepository,
		repos.ContentRelationRepository,
		courseService,
	)

	// 3. 创建沙箱客户端
	// 将 conf.LanguageConfig 转换为 sandbox.LanguageConfig
	sandboxLanguages := make(map[string]*sandbox.LanguageConfig)
	for lang, confLang := range conf.Config.Judge.Languages {
		sandboxLang := &sandbox.LanguageConfig{
			Name: confLang.Name,
		}
		if confLang.Compile != nil {
			// 调试：打印配置
			fmt.Printf("=== 加载语言配置: %s ===\nSourceFile: '%s'\nExecutableFile: '%s'\n",
				lang, confLang.Compile.SourceFile, confLang.Compile.ExecutableFile)

			sandboxLang.Compile = &sandbox.StageConfig{
				Args:           confLang.Compile.Args,
				Env:            confLang.Compile.Env,
				TimeLimit:      confLang.Compile.TimeLimit,
				MemoryLimit:    confLang.Compile.MemoryLimit,
				ProcLimit:      confLang.Compile.ProcLimit,
				SourceFile:     confLang.Compile.SourceFile,
				ExecutableFile: confLang.Compile.ExecutableFile,
			}
		}
		if confLang.Run != nil {
			sandboxLang.Run = &sandbox.StageConfig{
				Args:           confLang.Run.Args,
				Env:            confLang.Run.Env,
				TimeLimit:      confLang.Run.TimeLimit,
				MemoryLimit:    confLang.Run.MemoryLimit,
				ProcLimit:      confLang.Run.ProcLimit,
				SourceFile:     confLang.Run.SourceFile,
				ExecutableFile: confLang.Run.ExecutableFile,
			}
		}
		sandboxLanguages[lang] = sandboxLang
	}

	sandboxClient := sandbox.NewGoJudgeClient(
		conf.Config.Judge.SandboxURL,
		sandboxLanguages,
	)

	// 4. 创建消息队列
	queueSize := conf.Config.Judge.QueueSize
	if queueSize <= 0 {
		queueSize = 100
	}
	messageQueue := queue.NewChannelQueue(queueSize, submitService)

	// 5. 创建需要依赖注入的服务
	// ProblemService 需要课程内容关系仓储回填题目挂载范围。
	svc := &Service{}
	problemService := NewProblemService(
		sandboxClient,
		testCaseService,
		repos.ProblemRepository,
		repos.ContentRelationRepository,
		clazzService,
		svc,
	)

	// ClazzService 班级服务
	//clazzService := NewClazzService()

	// JudgeService 需要所有基础服务的引用
	judgeService := NewJudgeService(
		sandboxClient,
		submitService,
		testCaseService,
		problemService,
		messageQueue,
	)

	// StatisticsService 需要其他服务的引用
	statisticsService := NewStatisticsService(submitService, problemService, userService, repos.ProblemRepository)

	// DailyProblemService 需要 ProblemRepository
	dailyProblemService := NewDailyProblemService(repos.ProblemRepository)
	aiGeneratedProblemDraftService := NewAIGeneratedProblemDraftService(
		repos.AIGeneratedProblemDraftRepository,
		problemService,
		testCaseService,
		courseService,
	)

	aiRuntimeComponents := newAIRuntimeComponents()
	teacherCopilotService := NewTeacherCopilotService(
		userService,
		clazzService,
		problemService,
		aiRuntimeComponents.SessionService,
		aiRuntimeComponents.TaskManager,
		aiRuntimeComponents.ContextManager,
		aiRuntimeComponents.ChatModel,
	)

	rssService := NewRSSService(repos.RSSRepository, aiRuntimeComponents.ChatModel)

	return &Service{
		UserService:                    userService,
		ProblemService:                 problemService,
		SubmitService:                  submitService,
		JudgeService:                   judgeService,
		TestCaseService:                testCaseService,
		CourseService:                  courseService,
		ChoiceQuestionService:          choiceQuestionService,
		BlankQuestionService:           blankQuestionService,
		CourseContentService:           courseContentService,
		ContentSubmissionService:       contentSubmissionService,
		StatisticsService:              statisticsService,
		ClazzService:                   clazzService,
		DailyProblemService:            dailyProblemService,
		TeacherCopilot:                 teacherCopilotService,
		AIGeneratedProblemDraftService: aiGeneratedProblemDraftService,
		AssistantGateway:               aiRuntimeComponents.AssistantGateway,
		ModuleGateway:                  aiRuntimeComponents.ModuleGateway,
		RSSService:                     rssService,
	}
}

// Close 关闭服务，依次释放 MongoDB 与 Redis 连接。
func (s *Service) Close(ctx context.Context) error {
	if err := utils.CloseRedis(); err != nil {
		utils.Logger.Warnf("关闭 Redis 连接失败: %v", err)
	}
	return utils.CloseMongoDB(ctx)
}

type aiRuntimeComponents struct {
	AssistantRegistry *aiassistants.Registry
	AssistantGateway  aigateway.AssistantGateway
	ModuleRegistry    *aimodules.Registry
	ModuleGateway     aigateway.ModuleGateway
	SessionService    runtimesession.SessionService
	TaskManager       runtimetask.TaskManager
	ContextManager    runtimecontext.ContextManager
	ChatModel         llm.ModelFactory
	SandboxClient     sandbox.Client
}

func newAIRuntimeComponents() *aiRuntimeComponents {
	assistantRegistry := aiassistants.NewRegistry()
	moduleRegistry := aimodules.NewRegistry()

	var sessionService runtimesession.SessionService
	if utils.MongoDB != nil {
		sessionService = runtimesession.NewMongoSessionService(utils.MongoDB)
	}
	if sessionService == nil {
		sessionService = runtimesession.NewMemorySessionService()
	}
	taskManager := runtimetask.NewMemoryTaskManager()
	contextManager := runtimecontext.NewManager(sessionService, 10)
	// 未配置 ApiKey 时，若开启了 Ai.DemoMode 则切换到演示模式工厂，
	// 保证「生成题目 → 标准答案 → 测试用例 → 沙箱验证」整条链路可以演示。
	// 一旦配置了真实 ApiKey，这里始终优先使用配置工厂，演示模式自动让位。
	var modelFactory llm.ModelFactory = llm.NewConfigModelFactory()
	if !modelFactory.IsChatConfigured() && conf.Config.AI.DemoMode {
		utils.Logger.Warnf("AI 未配置 ApiKey 且 Ai.DemoMode=true，已启用演示模式：" +
			"题目/标准答案/测试用例三阶段返回内置样例，沙箱验证阶段仍为真实执行")
		modelFactory = llm.NewDemoModelFactory()
	}
	chatConfig := modelFactory.ChatConfig()
	chatConfigured := modelFactory.IsChatConfigured()
	// AI 配置日志只打印是否已设置，避免在 Docker 日志中泄露 API Key。
	utils.Logger.Infof("AI Chat配置检查: configured=%v, api_key_set=%v, base_url_set=%v, model_set=%v, base_url=%s, model=%s",
		chatConfigured,
		chatConfig.APIKey != "",
		chatConfig.BaseURL != "",
		chatConfig.Model != "",
		chatConfig.BaseURL,
		chatConfig.Model,
	)

	// 创建 sandbox client
	sandboxLanguages := make(map[string]*sandbox.LanguageConfig)
	for lang, confLang := range conf.Config.Judge.Languages {
		sandboxLang := &sandbox.LanguageConfig{
			Name: confLang.Name,
		}
		// MemoryLimit / ProcLimit / SourceFile / ExecutableFile 必须一并透传：
		// 若 procLimit 为 0，go-judge 会把 cgroup 的 pids.max 设为 0，
		// 导致任何 fork/clone 直接返回 EAGAIN（resource temporarily unavailable）；
		// SourceFile 为空则 copyIn 写入的文件名为空，编译源文件根本不会落盘。
		if confLang.Compile != nil {
			sandboxLang.Compile = &sandbox.StageConfig{
				Args:           confLang.Compile.Args,
				Env:            confLang.Compile.Env,
				TimeLimit:      confLang.Compile.TimeLimit,
				MemoryLimit:    confLang.Compile.MemoryLimit,
				ProcLimit:      confLang.Compile.ProcLimit,
				SourceFile:     confLang.Compile.SourceFile,
				ExecutableFile: confLang.Compile.ExecutableFile,
			}
		}
		if confLang.Run != nil {
			sandboxLang.Run = &sandbox.StageConfig{
				Args:           confLang.Run.Args,
				Env:            confLang.Run.Env,
				TimeLimit:      confLang.Run.TimeLimit,
				MemoryLimit:    confLang.Run.MemoryLimit,
				ProcLimit:      confLang.Run.ProcLimit,
				SourceFile:     confLang.Run.SourceFile,
				ExecutableFile: confLang.Run.ExecutableFile,
			}
		}
		sandboxLanguages[lang] = sandboxLang
	}
	sandboxClient := sandbox.NewGoJudgeClient(
		conf.Config.Judge.SandboxURL,
		sandboxLanguages,
	)
	utils.Logger.Infof("AI OJQuestionGen沙箱配置检查: sandbox_client_ready=%v, sandbox_url_set=%v, sandbox_url=%s, language_count=%d",
		sandboxClient != nil,
		conf.Config.Judge.SandboxURL != "",
		conf.Config.Judge.SandboxURL,
		len(sandboxLanguages),
	)

	// 注册 WebSearch Assistant
	tavilyConfig := toolsearch.ConfigFromEnv()
	toolProvider := toolsearch.NewProvider(
		toolsearch.NewHTTPClient(tavilyConfig, nil),
		tavilyConfig,
	)

	if chatConfigured && toolProvider.IsConfigured() {
		assistantRegistry.Register(aiwebsearch.NewAssistant(
			sessionService,
			taskManager,
			contextManager,
			toolProvider,
			modelFactory,
			aiwebsearch.DefaultInstruction,
		))
		utils.Logger.Infof("AI WebSearch Assistant注册成功: assistant_type=%s", aiwebsearch.CanonicalAssistantType)
	} else {
		assistantRegistry.Register(aiwebsearch.NewUnconfigured())
		utils.Logger.Warnf("AI WebSearch Assistant未完整配置: chat_configured=%v, search_configured=%v, assistant_type=%s",
			chatConfigured,
			toolProvider.IsConfigured(),
			aiwebsearch.CanonicalAssistantType,
		)
	}

	// 注册 OJQuestionGen Module
	if chatConfigured && sandboxClient != nil {
		coordinator := ojquestiongen.NewCoordinator(
			sessionService,
			taskManager,
			contextManager,
			modelFactory,
			sandboxClient,
		)
		moduleRegistry.Register(coordinator)
		utils.Logger.Infof("AI OJQuestionGen Module注册成功: module_type=%s", ojquestiongen.CanonicalModuleType)
	} else {
		utils.Logger.Warnf("AI OJQuestionGen Module未注册: module_type=%s, chat_configured=%v, sandbox_client_ready=%v",
			ojquestiongen.CanonicalModuleType,
			chatConfigured,
			sandboxClient != nil,
		)
	}

	return &aiRuntimeComponents{
		AssistantRegistry: assistantRegistry,
		AssistantGateway:  aigateway.NewAssistantGateway(assistantRegistry, taskManager),
		ModuleRegistry:    moduleRegistry,
		ModuleGateway:     aigateway.NewModuleGateway(moduleRegistry, taskManager),
		SessionService:    sessionService,
		TaskManager:       taskManager,
		ContextManager:    contextManager,
		ChatModel:         modelFactory,
		SandboxClient:     sandboxClient,
	}
}
//fix(ai):修复出题字段解析与沙箱配置，新增演示模式
