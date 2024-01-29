package logutils

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
)

type LogFormatter struct {
	ReportCaller bool
	OutPutFile   *os.File
}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if f.ReportCaller {
		frame := entry.Caller
		// 带调用者信息的日志输出
		return []byte(fmt.Sprintf("[%s] | %s | %s:%d: %s\n",
			entry.Level, entry.Time.Format(time.RFC3339),
			path.Base(frame.File), frame.Line,
			entry.Message)), nil
	}

	// 不带调用者信息的日志输出
	return []byte(fmt.Sprintf("[%s] | %s: %s\n",
		entry.Level, entry.Time.Format(time.RFC3339),
		entry.Message)), nil
}

func (f *LogFormatter) NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(f)

	// 配置输出文件路径
	if f.OutPutFile != nil {
		// 设置 logrus 的输出为文件和标准输出
		mw := io.MultiWriter(os.Stdout, f.OutPutFile)
		logger.SetOutput(mw)
	}

	// 是否开启调用者信息
	logger.SetReportCaller(f.ReportCaller)

	return logger
}
