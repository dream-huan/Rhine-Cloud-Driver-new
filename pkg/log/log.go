package log

import (
	"Rhine-Cloud-Driver/pkg/conf"
	"errors"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// Logger used to log log item
var Logger *zap.Logger

// ErrLevel level not valid
var ErrLevel = errors.New("invalid log level")

/**
 * 获取日志
 * filePath 日志文件路径
 * level 日志级别
 * maxSize 每个日志文件保存的最大尺寸 单位：M
 * maxBackups 日志文件最多保存多少个备份
 * maxAge 文件最多保存多少天
 * compress 是否压缩
 * serviceName 服务名
 */

// NewLogger create  Logger
func NewLogger(filePath string, level int, maxSize int, maxBackups int, maxAge int, compress, logConsole bool, serviceName string) (*zap.Logger, error) {
	if level < (int)(zap.DebugLevel) || level > (int)(zap.FatalLevel) {
		return nil, ErrLevel
	}
	l := (zapcore.Level)(level)
	core := newCore(filePath, l, maxSize, maxBackups, maxAge, compress, logConsole)
	return zap.New(core, zap.AddCaller(), zap.Development(), zap.Fields(zap.String("serviceName", serviceName))), nil
}

/**
 * zapcore构造
 */
func newCore(filePath string, level zapcore.Level, maxSize int, maxBackups int, maxAge int, compress, logConsole bool) zapcore.Core {
	//日志文件路径配置2
	hook := lumberjack.Logger{
		Filename:   filePath,   // 日志文件路径
		MaxSize:    maxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: maxBackups, // 日志文件最多保存多少个备份
		MaxAge:     maxAge,     // 文件最多保存多少天
		Compress:   compress,   // 是否压缩
	}
	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)
	//公用编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,    // 大写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	if logConsole {
		return zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),                                           // 编码器配置
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
			atomicLevel, // 日志级别
		)
	}
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),               // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(&hook)), // 打印到文件
		atomicLevel, // 日志级别
	)

}

func InitLog(cf *conf.LogConfig) error {
	var err error
	lg, err := NewLogger(cf.LogPath, cf.LogLevel, cf.MaxSize, cf.MaxBackup,
		cf.MaxAge, cf.Compress, cf.LogConsole, cf.ServiceName)
	if err != nil {
		return err
	}
	Logger = lg
	return nil
}
