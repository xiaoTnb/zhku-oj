/*
@Author: urmsone urmsone@163.com
@Date: 2025/1/24 17:38
@Name: config.go
@Description: 配置管理模块
*/

package conf

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	AppMode      string        // 服务器启动模式，默认debug模式
	AppEnv       string        // 服务器部署环境，默认为开发环境模式
	Port         string        // 服务启动端口
	GracefulTime time.Duration // 服务优雅重启的时间，单位为秒，默认5s
	Config       *configYaml
	v            *viper.Viper
)

type configYaml struct {
	App struct {
		Mode         string        `yaml:"Mode"`
		Host         string        `yaml:"Host"`
		Port         string        `yaml:"Port"`
		Env          string        `yaml:"Env"`
		GracefulTime time.Duration `yaml:"GracefulTime"`
	} `yaml:"App"`
	Mongo struct {
		Uri    string `yaml:"Uri"`
		DbName string `yaml:"DbName"`
	} `yaml:"Mongo"`
	Server struct {
		Domain string `yaml:"Domain"`
	} `yaml:"Server"`
	Judge   JudgeConfig `yaml:"Judge"`
	Sandbox struct {
		Url     string        `yaml:"Url"`
		Method  string        `yaml:"Method"`
		Timeout time.Duration `yaml:"Timeout"`
	} `yaml:"Sandbox"`
	Log struct {
		Level  string `yaml:"Level"`
		Format string `yaml:"Format"`
		Output string `yaml:"Output"`
		File   struct {
			Path       string `yaml:"Path"`
			MaxSize    int    `yaml:"MaxSize"`
			MaxBackups int    `yaml:"MaxBackups"`
			MaxAge     int    `yaml:"MaxAge"`
			Compress   bool   `yaml:"Compress"`
		} `yaml:"File"`
	} `yaml:"Log"`
	JWT struct {
		Secret     string        `yaml:"Secret"`
		ExpireTime time.Duration `yaml:"ExpireTime"`
	} `yaml:"JWT"`
	Redis struct {
		Host     string `yaml:"Host"`
		Port     int    `yaml:"Port"`
		Password string `yaml:"Password"`
		DB       int    `yaml:"DB"`
		PoolSize int    `yaml:"PoolSize"`
	} `yaml:"Redis"`
	RateLimit struct {
		Enabled bool   `yaml:"Enabled"` // 限流开关
		Type    string `yaml:"Type"`    // 限流类型: memory 或 redis
		CodeRun struct {
			Limit  int `yaml:"Limit"`  // 时间窗口内最大请求数
			Window int `yaml:"Window"` // 时间窗口（秒）
		} `yaml:"CodeRun"`
		Submit struct {
			Limit  int `yaml:"Limit"`
			Window int `yaml:"Window"`
		} `yaml:"Submit"`
	} `yaml:"RateLimit"`
	AI struct {
		// DemoMode 演示模式：仅在未配置 ApiKey 时生效，使用内置样例题目代替大模型输出，
		// 用于无网络/无 Key 环境下演示完整生成链路（沙箱验证阶段仍是真实执行）。
		DemoMode bool `yaml:"DemoMode"`
		Chat     struct {
			ApiKey  string `yaml:"ApiKey"`
			BaseUrl string `yaml:"BaseUrl"`
			Model   string `yaml:"Model"`
		} `yaml:"Chat"`
		Embedding struct {
			ApiKey  string `yaml:"ApiKey"`
			BaseUrl string `yaml:"BaseUrl"`
			Model   string `yaml:"Model"`
		} `yaml:"Embedding"`
	} `yaml:"AI"`
	Scalar struct {
		Title              string `yaml:"Title"`
		Description        string `yaml:"Description"`
		SwaggerURL         string `yaml:"SwaggerUrl"`
		Theme              string `yaml:"Theme"`
		Layout             string `yaml:"Layout"`
		ShowSidebar        bool   `yaml:"ShowSidebar"`
		HideDownloadButton bool   `yaml:"HideDownloadButton"`
		CustomCSS          string `yaml:"CustomCss"`
	} `yaml:"Scalar"`
}

// JudgeConfig 判题配置
type JudgeConfig struct {
	SandboxURL string                     `yaml:"SandboxURL"` // 沙箱服务地址
	Workers    int                        `yaml:"Workers"`    // 消费者数量
	QueueSize  int                        `yaml:"QueueSize"`  // 队列容量
	Languages  map[string]*LanguageConfig `yaml:"Languages"`  // 语言配置
}

// LanguageConfig 语言配置
type LanguageConfig struct {
	Name    string       `yaml:"name"`              // 语言名称
	Compile *StageConfig `yaml:"compile,omitempty"` // 编译配置（可选，解释型语言无需编译）
	Run     *StageConfig `yaml:"run"`               // 运行配置
}

// StageConfig 阶段配置（编译或运行）
type StageConfig struct {
	Args           []string `yaml:"args"`                      // 执行命令参数
	Env            []string `yaml:"env"`                       // 环境变量
	TimeLimit      int64    `yaml:"time_limit"`                // 时间限制（纳秒）
	MemoryLimit    int64    `yaml:"memory_limit"`              // 内存限制（字节）
	ProcLimit      int      `yaml:"proc_limit"`                // 进程数限制
	SourceFile     string   `yaml:"source_file,omitempty"`     // 源文件名
	ExecutableFile string   `yaml:"executable_file,omitempty"` // 可执行文件名
}

func Init() {
	// 尝试加载项目根目录 .env，不存在时继续使用系统环境变量/配置文件。
	_ = godotenv.Load()

	// viper mapstructure 配置读取
	v = viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("../../../conf")
	v.AddConfigPath("../../conf")
	v.AddConfigPath("../conf")
	v.AddConfigPath("./conf")

	// 设置环境变量替换器
	replacer := strings.NewReplacer(".", "_")
	v.SetEnvKeyReplacer(replacer)
	v.AutomaticEnv()

	// 读取配置文件
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("配置文件读取失败: %s", err))
	}

	// 解析配置到结构体
	c := &configYaml{}
	err = v.Unmarshal(&c)
	if err != nil {
		panic(fmt.Errorf("配置文件解析失败: %s", err))
	}

	// 手动解析 Judge.Languages (Viper 对嵌套 map 的支持不好,且会将 key 转为小写)
	if languagesMap := v.GetStringMap("judge.languages"); languagesMap != nil {
		c.Judge.Languages = make(map[string]*LanguageConfig)
		for lang := range languagesMap {
			// Viper 会将所有 key 转为小写,所以我们需要使用小写的 key
			langKey := strings.ToLower(lang)
			langConfig := &LanguageConfig{
				Name: v.GetString(fmt.Sprintf("judge.languages.%s.name", langKey)),
			}

			// 解析 compile 配置
			compileKey := fmt.Sprintf("judge.languages.%s.compile", langKey)
			if v.IsSet(compileKey) {
				langConfig.Compile = &StageConfig{
					Args:           v.GetStringSlice(fmt.Sprintf("%s.args", compileKey)),
					Env:            v.GetStringSlice(fmt.Sprintf("%s.env", compileKey)),
					TimeLimit:      v.GetInt64(fmt.Sprintf("%s.time_limit", compileKey)),
					MemoryLimit:    v.GetInt64(fmt.Sprintf("%s.memory_limit", compileKey)),
					ProcLimit:      v.GetInt(fmt.Sprintf("%s.proc_limit", compileKey)),
					SourceFile:     v.GetString(fmt.Sprintf("%s.source_file", compileKey)),
					ExecutableFile: v.GetString(fmt.Sprintf("%s.executable_file", compileKey)),
				}
			}

			// 解析 run 配置
			runKey := fmt.Sprintf("judge.languages.%s.run", langKey)
			if v.IsSet(runKey) {
				langConfig.Run = &StageConfig{
					Args:           v.GetStringSlice(fmt.Sprintf("%s.args", runKey)),
					Env:            v.GetStringSlice(fmt.Sprintf("%s.env", runKey)),
					TimeLimit:      v.GetInt64(fmt.Sprintf("%s.time_limit", runKey)),
					MemoryLimit:    v.GetInt64(fmt.Sprintf("%s.memory_limit", runKey)),
					ProcLimit:      v.GetInt(fmt.Sprintf("%s.proc_limit", runKey)),
					SourceFile:     v.GetString(fmt.Sprintf("%s.source_file", runKey)),
					ExecutableFile: v.GetString(fmt.Sprintf("%s.executable_file", runKey)),
				}
			}

			c.Judge.Languages[lang] = langConfig
		}
	}

	applyEnvOverrides(c)

	Config = c

	// 开发环境打印配置信息
	if Config.App.Env == "development" {
		fmt.Printf("配置加载成功: %+v\n", c)
	}
}

func ConfigUtils() *viper.Viper {
	return v
}

// applyEnvOverrides 从环境变量覆盖配置（.env 文件会被 godotenv.Load() 加载）
func applyEnvOverrides(c *configYaml) {
	if c == nil {
		return
	}

	// MongoDB
	if v, ok := lookupNonEmptyEnv("MONGO_URI"); ok {
		c.Mongo.Uri = v
	}
	if v, ok := lookupNonEmptyEnv("MONGO_DB_NAME"); ok {
		c.Mongo.DbName = v
	}

	// Redis
	if v, ok := lookupNonEmptyEnv("REDIS_HOST"); ok {
		c.Redis.Host = v
	}
	if v, ok := lookupNonEmptyEnv("REDIS_PORT"); ok {
		if port, err := parseInt(v); err == nil {
			c.Redis.Port = port
		}
	}
	if v, ok := lookupNonEmptyEnv("REDIS_PASSWORD"); ok {
		c.Redis.Password = v
	}
	if v, ok := lookupNonEmptyEnv("REDIS_DB"); ok {
		if db, err := parseInt(v); err == nil {
			c.Redis.DB = db
		}
	}

	// Judge / Sandbox
	if v, ok := lookupNonEmptyEnv("JUDGE_SANDBOXURL"); ok {
		c.Judge.SandboxURL = v
	}
	if v, ok := lookupNonEmptyEnv("SANDBOX_URL"); ok {
		c.Sandbox.Url = v
	}

	// AI Chat
	if v, ok := lookupNonEmptyEnv("AI_API_KEY"); ok {
		c.AI.Chat.ApiKey = v
	}
	if v, ok := lookupNonEmptyEnv("AI_BASE_URL"); ok {
		c.AI.Chat.BaseUrl = v
	}
	if v, ok := lookupNonEmptyEnv("AI_MODEL"); ok {
		c.AI.Chat.Model = v
	}

	// AI Embedding
	if v, ok := lookupNonEmptyEnv("EMBEDDING_API_KEY"); ok {
		c.AI.Embedding.ApiKey = v
	}
	if v, ok := lookupNonEmptyEnv("EMBEDDING_BASE_URL"); ok {
		c.AI.Embedding.BaseUrl = v
	}
	if v, ok := lookupNonEmptyEnv("EMBEDDING_MODEL"); ok {
		c.AI.Embedding.Model = v
	}

	// JWT
	if v, ok := lookupNonEmptyEnv("JWT_SECRET"); ok {
		c.JWT.Secret = v
	}
}

func lookupNonEmptyEnv(key string) (string, bool) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	// Docker Compose 变量未展开时会把 ${KEY} 原样传入容器，这类值不能覆盖配置文件中的可用默认值。
	if strings.Contains(value, "${") || strings.Contains(value, "$${") {
		return "", false
	}
	return value, true
}

func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
//fix(ai): 修复出题字段解析与沙箱配置，新增演示模式
