package base

import (
	"fmt"
	"github.com/livekit/protocol/logger"
	"go.uber.org/zap/zapcore"
	"time"
)

const (
	logTmFmtWithMS = "2006-01-02 15:04:05.000"
)

var (
	_levelToColor = map[zapcore.Level]_TextColor{
		zapcore.DebugLevel:  Magenta,
		zapcore.InfoLevel:   Blue,
		zapcore.WarnLevel:   Yellow,
		zapcore.ErrorLevel:  Red,
		zapcore.DPanicLevel: Red,
		zapcore.PanicLevel:  Red,
		zapcore.FatalLevel:  Red,
	}
	_unknownLevelColor = Red

	_levelToLowercaseColorString = make(map[zapcore.Level]string, len(_levelToColor))
	_levelToCapitalColorString   = make(map[zapcore.Level]string, len(_levelToColor))
)

func init() {
	for level, color := range _levelToColor {
		_levelToLowercaseColorString[level] = color.Add(level.String())
		_levelToCapitalColorString[level] = color.Add(level.CapitalString()[0:4])
	}

	logger.InitFromConfig(&logger.Config{
		Level:             "info",
		EncoderConfig:     &RecommendedEncoderConfig,
		DisableCaller:     false,
		DisableStacktrace: true,
	}, "default")
}

func InitDefaultLogger() {
	// trigger to init()
}

func InitSimpleLogger(name, level string) {
	logger.InitFromConfig(&logger.Config{
		Level:             level,
		EncoderConfig:     &RecommendedEncoderConfig,
		DisableCaller:     false,
		DisableStacktrace: true,
	}, name)
}

func InitLogger(name string, cfg *logger.Config) {
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.EncoderConfig == nil {
		cfg.EncoderConfig = &RecommendedEncoderConfig
	}
	logger.InitFromConfig(cfg, name)
}

var (
	// 自定义时间输出格式
	customTimeEncoder = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + t.Format(logTmFmtWithMS) + "]")
	}

	// 自定义日志级别显示
	customLevelEncoder = func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		s, ok := _levelToCapitalColorString[level]
		if !ok {
			s = _unknownLevelColor.Add(level.CapitalString()[0:4])
		}

		enc.AppendString("[" + s + "]")
	}

	// 调用路径/行号输出项
	customCallerEncoder = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + caller.TrimmedPath() + "]")
	}

	RecommendedEncoderConfig = zapcore.EncoderConfig{
		CallerKey:        "caller",
		LevelKey:         "level",
		MessageKey:       "msg",
		TimeKey:          "ts",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeTime:       customTimeEncoder,
		EncodeLevel:      customLevelEncoder,
		EncodeCaller:     customCallerEncoder,
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeName:       zapcore.FullNameEncoder,
		ConsoleSeparator: " ",
	}
)

// _TextColor represents a text color.
type _TextColor uint8

// Foreground colors.
const (
	Black _TextColor = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// Add adds the coloring to the given string.
func (c _TextColor) Add(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(c), s)
}
