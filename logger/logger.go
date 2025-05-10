package logger

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

// Logger 封裝的日誌結構
type Logger struct {
	slog    *slog.Logger
	handler slog.Handler
}

// Config 日誌配置
type Config struct {
	Level      slog.Level `mapstructure:"level"`       // 日誌級別
	Format     string     `mapstructure:"format"`      // "json" 或 "text"
	AddSource  bool       `mapstructure:"add_source"`  // 是否添加源信息
	Output     *os.File   `mapstructure:"output"`      // 輸出目標，默認為 os.Stdout
	TimeFormat string     `mapstructure:"time_format"` // 自定義時間格式（可選）
}

// singleton 相關變量
var (
	instance *Logger
	once     sync.Once
	mutex    sync.Mutex
)

// Init 獲取單例 Logger
func Init(conf *Config) *Logger {
	once.Do(func() {
		var config Config
		if conf == nil {
			// 默認配置
			config = Config{
				Level:  slog.LevelInfo,
				Format: "text",
				Output: os.Stdout,
			}
		} else {
			config = *conf
		}

		if config.Output == nil {
			config.Output = os.Stdout
		}

		// 設置處理器選項
		opts := &slog.HandlerOptions{
			Level:     config.Level,
			AddSource: config.AddSource,
		}

		if config.TimeFormat != "" {
			opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{Key: "time", Value: slog.StringValue(a.Value.Time().Format(config.TimeFormat))}
				}
				return a
			}
		}

		// 根據格式選擇處理器
		var handler slog.Handler
		switch config.Format {
		case "json":
			handler = slog.NewJSONHandler(config.Output, opts)
		default: // 默認為 text
			handler = slog.NewTextHandler(config.Output, opts)
		}

		instance = &Logger{
			slog:    slog.New(handler),
			handler: handler,
		}
	})
	return instance
}

func GetInstance() *Logger {
	if instance == nil {
		panic("logger not initialized")
	}
	return instance
}

// With 添加上下文屬性
func (l *Logger) With(args ...any) *Logger {
	mutex.Lock()
	defer mutex.Unlock()
	return &Logger{
		slog:    l.slog.With(args...),
		handler: l.handler,
	}
}

func (l *Logger) WithGroup(name string) *Logger {
	mutex.Lock()
	defer mutex.Unlock()
	return &Logger{
		slog:    l.slog.WithGroup(name),
		handler: l.handler,
	}
}

// Debug 記錄 Debug 級別日誌
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// Info 記錄 Info 級別日誌
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// Warn 記錄 Warn 級別日誌
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

// Error 記錄 Error 級別日誌
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// Panic 記錄日誌並觸發 panic
func (l *Logger) Panic(msg string, args ...any) {
	l.slog.Error(msg, args...) // 先記錄為 Error 等級
	panic(msg)                 // 然後觸發 panic
}

// Fatal 記錄日誌並退出程序（可選）
func (l *Logger) Fatal(msg string, args ...any) {
	l.slog.Error(msg, args...)
	os.Exit(1) // 退出程序
}

// WithContext 從 context.Context 中提取上下文屬性
func (l *Logger) WithContext(ctx context.Context) *Logger {
	mutex.Lock()
	defer mutex.Unlock()

	// 從 context 中提取值（假設有一些鍵）
	var attrs []any
	if reqID, ok := ctx.Value("request_id").(string); ok {
		attrs = append(attrs, "request_id", reqID)
	}
	if userID, ok := ctx.Value("user_id").(int); ok {
		attrs = append(attrs, "user_id", userID)
	}

	// 如果沒有上下文屬性，返回原 Logger
	if len(attrs) == 0 {
		return l
	}
	return &Logger{
		slog:    l.slog.With(attrs...),
		handler: l.handler,
	}
}

// SetDefault 設置為全局默認 Logger
func (l *Logger) SetDefault() {
	mutex.Lock()
	defer mutex.Unlock()
	slog.SetDefault(l.slog)
}

// Handler 返回底層的 slog.Handler
func (l *Logger) Handler() slog.Handler {
	return l.handler
}
