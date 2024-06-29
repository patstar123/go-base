// Copyright 2023 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	discardLogger        = logr.Discard()
	defaultLogger Logger = LogRLogger(discardLogger)
	pkgLogger     Logger = LogRLogger(discardLogger)
	pkgConfig     *zap.Config
)

// InitFromConfig initializes a Zap-based logger
func InitFromConfig(conf *Config, name string) {
	l, c, err := NewZapLogger(conf)
	if err == nil {
		SetLogger(l, c, name)
	}
}

// GetLogger returns the logger that was set with SetLogger with an extra depth of 1
func GetLogger() Logger {
	return defaultLogger
}

func GetPkgLogger() Logger {
	return pkgLogger
}

// SetLogger lets you use a custom logger. Pass in a logr.Logger with default depth
func SetLogger(l Logger, c *zap.Config, name string) {
	defaultLogger = l.WithCallDepth(1).WithName(name)
	// pkg wrapper needs to drop two levels of depth
	pkgLogger = l.WithCallDepth(2).WithName(name)
	pkgConfig = c
}

func SetLogLevel(level string) {
	if pkgConfig != nil {
		lvl := ParseZapLevel(level)
		pkgConfig.Level.SetLevel(lvl)
	}
	defaultLogger.SetLevel(level)
	pkgLogger.SetLevel(level)
}

func GetLogLevel() string {
	return pkgLogger.GetLevel()
}

func IsEnabledDebug() bool {
	return pkgLogger.IsEnabledDebug()
}

func Debug(args ...interface{}) {
	pkgLogger.Debug(args...)
}

func Info(args ...interface{}) {
	pkgLogger.Info(args...)
}

func Warn(args ...interface{}) {
	pkgLogger.Warn(args...)
}

func Error(args ...interface{}) {
	pkgLogger.Error(args...)
}

func Debugf(template string, args ...interface{}) {
	pkgLogger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	pkgLogger.Infof(template, args...)
}

func Warnf(err error, template string, args ...interface{}) {
	pkgLogger.Warnf(err, template, args...)
}

func Errorf(err error, template string, args ...interface{}) {
	pkgLogger.Errorf(err, template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	pkgLogger.Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	pkgLogger.Infow(msg, keysAndValues...)
}

func Warnw(msg string, err error, keysAndValues ...interface{}) {
	pkgLogger.Warnw(msg, err, keysAndValues...)
}

func Errorw(msg string, err error, keysAndValues ...interface{}) {
	pkgLogger.Errorw(msg, err, keysAndValues...)
}

func Debugln(args ...interface{}) {
	pkgLogger.Debugln(args...)
}

func Infoln(args ...interface{}) {
	pkgLogger.Infoln(args...)
}

func Warnln(args ...interface{}) {
	pkgLogger.Warnln(args...)
}

func Errorln(args ...interface{}) {
	pkgLogger.Errorln(args...)
}

func ParseZapLevel(level string) zapcore.Level {
	lvl := zapcore.InfoLevel
	if level != "" {
		_ = lvl.UnmarshalText([]byte(level))
	}
	return lvl
}

type Logger interface {
	IsEnabledDebug() bool
	WithValues(keysAndValues ...interface{}) Logger
	WithName(name string) Logger
	// WithComponent creates a new logger with name as "<name>.<component>", and uses a log level as specified
	WithComponent(component string) Logger
	WithCallDepth(depth int) Logger
	WithItemSampler() Logger
	// WithoutSampler returns the original logger without sampling
	WithoutSampler() Logger

	SetLevel(lvl string) error
	GetLevel() string

	// add by tuyj
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(err error, template string, args ...interface{})
	Errorf(err error, template string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, err error, keysAndValues ...interface{})
	Errorw(msg string, err error, keysAndValues ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
}

type sharedConfig struct {
	level           zap.AtomicLevel
	mu              sync.Mutex
	componentLevels map[string]zap.AtomicLevel
	config          *Config
}

func newSharedConfig(conf *Config) *sharedConfig {
	sc := &sharedConfig{
		level:           zap.NewAtomicLevelAt(ParseZapLevel(conf.Level)),
		config:          conf,
		componentLevels: make(map[string]zap.AtomicLevel),
	}
	conf.AddUpdateObserver(sc.onConfigUpdate)
	_ = sc.onConfigUpdate(conf)
	return sc
}

func (c *sharedConfig) onConfigUpdate(conf *Config) error {
	// update log levels
	c.level.SetLevel(ParseZapLevel(conf.Level))

	// we have to update alla existing component levels
	c.mu.Lock()
	c.config = conf
	for component, atomicLevel := range c.componentLevels {
		effectiveLevel := c.level.Level()
		parts := strings.Split(component, ".")
	confSearch:
		for len(parts) > 0 {
			search := strings.Join(parts, ".")
			if compLevel, ok := conf.ComponentLevels[search]; ok {
				effectiveLevel = ParseZapLevel(compLevel)
				break confSearch
			}
			parts = parts[:len(parts)-1]
		}
		atomicLevel.SetLevel(effectiveLevel)
	}
	c.mu.Unlock()
	return nil
}

// ensure we have an atomic level in the map representing the full component path
// this makes it possible to update the log level after the fact
func (c *sharedConfig) setEffectiveLevel(component string) zap.AtomicLevel {
	c.mu.Lock()
	defer c.mu.Unlock()
	if compLevel, ok := c.componentLevels[component]; ok {
		return compLevel
	}

	// search up the hierarchy to find the first level that is set
	atomicLevel := zap.NewAtomicLevelAt(c.level.Level())
	c.componentLevels[component] = atomicLevel
	parts := strings.Split(component, ".")
	for len(parts) > 0 {
		search := strings.Join(parts, ".")
		if compLevel, ok := c.config.ComponentLevels[search]; ok {
			atomicLevel.SetLevel(ParseZapLevel(compLevel))
			return atomicLevel
		}
		parts = parts[:len(parts)-1]
	}
	return atomicLevel
}

type ZapLogger struct {
	zap *zap.SugaredLogger
	// store original logger without sampling to avoid multiple samplers
	unsampled *zap.SugaredLogger
	component string
	// use a nested field as pointer so that all loggers share the same sharedConfig
	sharedConfig   *sharedConfig
	level          zap.AtomicLevel
	SampleDuration time.Duration
	SampleInitial  int
	SampleInterval int
}

func NewZapLogger(conf *Config) (*ZapLogger, *zap.Config, error) {
	sc := newSharedConfig(conf)
	zl := &ZapLogger{
		sharedConfig:   sc,
		level:          sc.level,
		SampleDuration: time.Duration(conf.ItemSampleSeconds) * time.Second,
		SampleInitial:  conf.ItemSampleInitial,
		SampleInterval: conf.ItemSampleInterval,
	}
	zapConfig := zap.Config{
		// set to the lowest level since we are doing our own filtering in `isEnabled`
		Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:       false,
		Encoding:          "console",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     conf.DisableCaller,
		DisableStacktrace: conf.DisableStacktrace,
	}
	if conf.JSON {
		zapConfig.Encoding = "json"
		zapConfig.EncoderConfig = zap.NewProductionEncoderConfig()
	}
	if conf.EncoderConfig != nil {
		zapConfig.EncoderConfig = *conf.EncoderConfig
	}
	l, err := zapConfig.Build()
	if err != nil {
		return nil, nil, err
	}
	zl.unsampled = l.Sugar()

	if conf.Sample {
		// use a sampling logger for the main logger
		samplingConf := &zap.SamplingConfig{
			Initial:    conf.SampleInitial,
			Thereafter: conf.SampleInterval,
		}
		// sane defaults
		if samplingConf.Initial == 0 {
			samplingConf.Initial = 20
		}
		if samplingConf.Thereafter == 0 {
			samplingConf.Thereafter = 100
		}
		zl.zap = l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(
				core,
				time.Second,
				samplingConf.Initial,
				samplingConf.Thereafter,
			)
		})).Sugar()
	} else {
		zl.zap = zl.unsampled
	}
	return zl, &zapConfig, nil
}

func (l *ZapLogger) WithFieldSampler(config FieldSamplerConfig) *ZapLogger {
	dup := *l
	dup.zap = l.zap.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return NewFieldSampler(core, config)
	}))
	return &dup
}

func (l *ZapLogger) ToZap() *zap.SugaredLogger {
	return l.zap
}

func (l *ZapLogger) IsEnabledDebug() bool {
	return l.isEnabled(zapcore.DebugLevel)
}

func (l *ZapLogger) Debug(args ...interface{}) {
	if !l.isEnabled(zapcore.DebugLevel) {
		return
	}
	l.zap.Debugln(args...)
}

func (l *ZapLogger) Info(args ...interface{}) {
	if !l.isEnabled(zapcore.InfoLevel) {
		return
	}
	l.zap.Debugln(args...)
}

func (l *ZapLogger) Warn(args ...interface{}) {
	if !l.isEnabled(zapcore.WarnLevel) {
		return
	}
	l.zap.Debugln(args...)
}

func (l *ZapLogger) Error(args ...interface{}) {
	if !l.isEnabled(zapcore.ErrorLevel) {
		return
	}
	l.zap.Debugln(args...)
}

func (l *ZapLogger) Debugf(template string, args ...interface{}) {
	if !l.isEnabled(zapcore.DebugLevel) {
		return
	}
	l.zap.Debugf(template, args...)
}

func (l *ZapLogger) Infof(template string, args ...interface{}) {
	if !l.isEnabled(zapcore.InfoLevel) {
		return
	}
	l.zap.Infof(template, args...)
}

func (l *ZapLogger) Warnf(err error, template string, args ...interface{}) {
	if !l.isEnabled(zapcore.WarnLevel) {
		return
	}
	if err != nil {
		template += " (error: %v)"
		args = append(args, err)
	}
	l.zap.Warnf(template, args...)
}

func (l *ZapLogger) Errorf(err error, template string, args ...interface{}) {
	if !l.isEnabled(zapcore.ErrorLevel) {
		return
	}
	if err != nil {
		template += " (error: %v)"
		args = append(args, err)
	}
	l.zap.Errorf(template, args...)
}

func (l *ZapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	if !l.isEnabled(zapcore.DebugLevel) {
		return
	}
	l.zap.Debugw(msg, keysAndValues...)
}

func (l *ZapLogger) Infow(msg string, keysAndValues ...interface{}) {
	if !l.isEnabled(zapcore.InfoLevel) {
		return
	}
	l.zap.Infow(msg, keysAndValues...)
}

func (l *ZapLogger) Warnw(msg string, err error, keysAndValues ...interface{}) {
	if !l.isEnabled(zapcore.WarnLevel) {
		return
	}
	if err != nil {
		keysAndValues = append(keysAndValues, "error", err)
	}
	l.zap.Warnw(msg, keysAndValues...)
}

func (l *ZapLogger) Errorw(msg string, err error, keysAndValues ...interface{}) {
	if !l.isEnabled(zapcore.ErrorLevel) {
		return
	}
	if err != nil {
		keysAndValues = append(keysAndValues, "error", err)
	}
	l.zap.Errorw(msg, keysAndValues...)
}

func (l *ZapLogger) Debugln(args ...interface{}) {
	if !l.isEnabled(zapcore.DebugLevel) {
		return
	}
	l.zap.Debugln(args...)
}

func (l *ZapLogger) Infoln(args ...interface{}) {
	if !l.isEnabled(zapcore.InfoLevel) {
		return
	}
	l.zap.Infoln(args...)
}

func (l *ZapLogger) Warnln(args ...interface{}) {
	if !l.isEnabled(zapcore.WarnLevel) {
		return
	}
	l.zap.Warnln(args...)
}

func (l *ZapLogger) Errorln(args ...interface{}) {
	if !l.isEnabled(zapcore.ErrorLevel) {
		return
	}
	l.zap.Errorln(args...)
}

func (l *ZapLogger) WithValues(keysAndValues ...interface{}) Logger {
	dup := *l
	dup.zap = l.zap.With(keysAndValues...)
	// mirror unsampled logger too
	if l.unsampled == l.zap {
		dup.unsampled = dup.zap
	} else {
		dup.unsampled = l.unsampled.With(keysAndValues...)
	}
	return &dup
}

func (l *ZapLogger) WithName(name string) Logger {
	dup := *l
	dup.zap = l.zap.Named(name)
	if l.unsampled == l.zap {
		dup.unsampled = dup.zap
	} else {
		dup.unsampled = l.unsampled.Named(name)
	}
	return &dup
}

func (l *ZapLogger) WithComponent(component string) Logger {
	// zap automatically appends .<name> to the logger name
	dup := l.WithName(component).(*ZapLogger)
	if dup.component == "" {
		dup.component = component
	} else {
		dup.component = dup.component + "." + component
	}
	dup.level = dup.sharedConfig.setEffectiveLevel(dup.component)
	return dup
}

func (l *ZapLogger) WithCallDepth(depth int) Logger {
	dup := *l
	dup.zap = l.zap.WithOptions(zap.AddCallerSkip(depth))
	if l.unsampled == l.zap {
		dup.unsampled = dup.zap
	} else {
		dup.unsampled = l.unsampled.WithOptions(zap.AddCallerSkip(depth))
	}
	return &dup
}

func (l *ZapLogger) WithItemSampler() Logger {
	if l.SampleDuration == 0 {
		return l
	}
	dup := *l
	dup.zap = l.unsampled.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewSamplerWithOptions(
			core,
			l.SampleDuration,
			l.SampleInitial,
			l.SampleInterval,
		)
	}))
	return &dup
}

func (l *ZapLogger) WithoutSampler() Logger {
	if l.unsampled == l.zap {
		return l
	}
	dup := *l
	dup.zap = l.unsampled
	return &dup
}

func (l *ZapLogger) SetLevel(level string) error {
	lvl := ParseZapLevel(level)
	l.level = zap.NewAtomicLevelAt(lvl)
	return nil
}

func (l *ZapLogger) GetLevel() string {
	return l.level.String()
}

func (l *ZapLogger) isEnabled(level zapcore.Level) bool {
	return level >= l.level.Level()
}

type LogRLogger logr.Logger

func (l LogRLogger) toLogr() logr.Logger {
	if logr.Logger(l).GetSink() == nil {
		return discardLogger
	}
	return logr.Logger(l)
}

func (l LogRLogger) IsEnabledDebug() bool {
	return true
}

func (l LogRLogger) Debug(args ...interface{}) {
	l.toLogr().V(1).Info("", args...)
}

func (l LogRLogger) Info(args ...interface{}) {
	l.toLogr().Info("", args...)
}

func (l LogRLogger) Warn(args ...interface{}) {
	l.toLogr().Info("", args...)
}

func (l LogRLogger) Error(args ...interface{}) {
	l.toLogr().Error(nil, "", args)
}

func (l LogRLogger) Debugf(template string, args ...interface{}) {
	l.toLogr().V(1).Info(fmt.Sprintf(template, args...))
}

func (l LogRLogger) Infof(template string, args ...interface{}) {
	l.toLogr().Info(fmt.Sprintf(template, args...))
}

func (l LogRLogger) Warnf(err error, template string, args ...interface{}) {
	l.toLogr().Info(fmt.Sprintf(template, args...), "error", err)
}

func (l LogRLogger) Errorf(err error, template string, args ...interface{}) {
	l.toLogr().Error(err, fmt.Sprintf(template, args...))
}

func (l LogRLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.toLogr().V(1).Info(msg, keysAndValues...)
}

func (l LogRLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.toLogr().Info(msg, keysAndValues...)
}

func (l LogRLogger) Warnw(msg string, err error, keysAndValues ...interface{}) {
	if err != nil {
		keysAndValues = append(keysAndValues, "error", err)
	}
	l.toLogr().Info(msg, keysAndValues...)
}

func (l LogRLogger) Errorw(msg string, err error, keysAndValues ...interface{}) {
	l.toLogr().Error(err, msg, keysAndValues...)
}

func (l LogRLogger) Debugln(args ...interface{}) {
	l.toLogr().V(1).Info("", args...)
}

func (l LogRLogger) Infoln(args ...interface{}) {
	l.toLogr().Info("", args...)
}

func (l LogRLogger) Warnln(args ...interface{}) {
	l.toLogr().Info("", args...)
}

func (l LogRLogger) Errorln(args ...interface{}) {
	l.toLogr().Error(nil, "", args)
}

func (l LogRLogger) WithValues(keysAndValues ...interface{}) Logger {
	return LogRLogger(l.toLogr().WithValues(keysAndValues...))
}

func (l LogRLogger) WithName(name string) Logger {
	return LogRLogger(l.toLogr().WithName(name))
}

func (l LogRLogger) WithComponent(component string) Logger {
	return LogRLogger(l.toLogr().WithName(component))
}

func (l LogRLogger) WithCallDepth(depth int) Logger {
	return LogRLogger(l.toLogr().WithCallDepth(depth))
}

func (l LogRLogger) WithItemSampler() Logger {
	// logr does not support sampling
	return l
}

func (l LogRLogger) WithoutSampler() Logger {
	return l
}

func (l LogRLogger) SetLevel(level string) error {
	return errors.New("not supported")
}

func (l LogRLogger) GetLevel() string {
	return "info"
}
