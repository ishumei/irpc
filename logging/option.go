package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type coreConfig struct {
	enc zapcore.Encoder
	ws  zapcore.WriteSyncer
	lvl zap.AtomicLevel
}

type traceConfig struct {
	recordStackTraceInSpan bool
	errorSpanLevel         zapcore.Level
}

type config struct {
	coreConfig  coreConfig
	zapOpts     []zap.Option
	traceConfig *traceConfig
}

// defaultCoreConfig default zapcore config: json encoder, atomic level, stdout write syncer
func defaultCoreConfig() *coreConfig {
	// default log encoder
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	// default log level
	lvl := zap.NewAtomicLevelAt(zap.InfoLevel)
	// default write syncer stdout
	ws := zapcore.AddSync(os.Stdout)

	return &coreConfig{
		enc: enc,
		ws:  ws,
		lvl: lvl,
	}
}

// defaultConfig default config
func defaultConfig() *config {
	coreConfig := defaultCoreConfig()
	return &config{
		coreConfig: *coreConfig,
		traceConfig: &traceConfig{
			recordStackTraceInSpan: true,
			errorSpanLevel:         zapcore.ErrorLevel,
		},
		zapOpts: []zap.Option{},
	}
}

// WithCoreEnc zapcore encoder
func WithCoreEnc(enc zapcore.Encoder) Option {
	return option(func(cfg *config) {
		cfg.coreConfig.enc = enc
	})
}

// WithCoreWs zapcore write syncer
func WithCoreWs(ws zapcore.WriteSyncer) Option {
	return option(func(cfg *config) {
		cfg.coreConfig.ws = ws
	})
}

// WithCoreLevel zapcore log level
func WithCoreLevel(lvl zap.AtomicLevel) Option {
	return option(func(cfg *config) {
		cfg.coreConfig.lvl = lvl
	})
}

// WithZapOptions add origin zap option
func WithZapOptions(opts ...zap.Option) Option {
	return option(func(cfg *config) {
		cfg.zapOpts = append(cfg.zapOpts, opts...)
	})
}

// WithTraceErrorSpanLevel trace error span level option
func WithTraceErrorSpanLevel(level zapcore.Level) Option {
	return option(func(cfg *config) {
		cfg.traceConfig.errorSpanLevel = level
	})
}

// WithRecordStackTraceInSpan record stack track option
func WithRecordStackTraceInSpan(recordStackTraceInSpan bool) Option {
	return option(func(cfg *config) {
		cfg.traceConfig.recordStackTraceInSpan = recordStackTraceInSpan
	})
}
