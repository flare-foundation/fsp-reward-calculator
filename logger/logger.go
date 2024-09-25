package logger

import (
	"errors"
	"fsp-rewards-calculator/config"
	"log"
	"os"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	sugar *zap.SugaredLogger
)

const (
	timeFormat = "[01-02|15:04:05.000]"
)

func init() {
	sugar = createSugaredLogger(DefaultLoggerConfig())
}

func Configure(config config.LoggerConfig) {
	sugar = createSugaredLogger(config)
}

func createSugaredLogger(config config.LoggerConfig) *zap.SugaredLogger {
	atom := zap.NewAtomicLevel()
	cores := make([]zapcore.Core, 0)
	if config.Console {
		cores = append(cores, createConsoleLoggerCore(atom))
	}
	if len(config.File) > 0 {
		cores = append(cores, createFileLoggerCore(config, atom))
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core,
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	defer func() {
		err := logger.Sync()

		if err != nil && (!errors.Is(err, syscall.EBADF) && !errors.Is(err, syscall.ENOTTY)) {
			log.Print("Failed to sync logger:", err)
		}
	}()

	sugar = logger.Sugar()

	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		sugar.Errorf("Wrong level %s", config.Level)
	}
	atom.SetLevel(level)
	return sugar
}

func createFileLoggerCore(config config.LoggerConfig, atom zap.AtomicLevel) zapcore.Core {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename: config.File,
		MaxSize:  config.MaxFileSize,
	})
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeLevel = fileLevelEncoder
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(timeFormat)
	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		w,
		atom,
	)
}

func createConsoleLoggerCore(atom zap.AtomicLevel) zapcore.Core {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeLevel = consoleColorLevelEncoder
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(timeFormat)
	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		atom,
	)
}

func consoleColorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	s, ok := levelToCapitalColorString[l]
	if !ok {
		s = unknownLevelColor.Wrap(l.CapitalString())
	}
	enc.AppendString(s)
}

func fileLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(l.CapitalString())
}

func DefaultLoggerConfig() config.LoggerConfig {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}
	return config.LoggerConfig{
		Level:   logLevel,
		Console: true,
	}
}

func Warn(msg string, args ...interface{}) {
	sugar.Warnf(msg, args...)
}

func Error(msg string, args ...interface{}) {
	sugar.Errorf(msg, args...)
}

func Info(msg string, args ...interface{}) {
	sugar.Infof(msg, args...)
}

func Debug(msg string, args ...interface{}) {
	sugar.Debugf(msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	sugar.Fatalf(msg, args...)
}
