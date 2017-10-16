package pub

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	_ "golib/modules/logr"
	"path/filepath"
	"runtime"
	"unGateWay/Config"
)

/*基础资源用于产品结构加载*/
type BaseResource struct {
	logger *logrus.Entry //logger固定生成格式
}

//N 运行名字section、trans
func (t *BaseResource) SetLogger(logName string) {
	t.logger = Config.GetNameLog(logName).WithField("N", logName)
}

//报错文件目录&&行数
func (t *BaseResource) locate() *logrus.Entry {
	_, path, line, ok := runtime.Caller(2)
	if ok {
		_, file := filepath.Split(path)
		return t.logger.WithField("F", fmt.Sprintf("%s:%d", file, line))
	}
	return t.logger
}

func (t *BaseResource) GetLogger() *logrus.Entry {
	return t.logger
}

func (t *BaseResource) Info(msg ...interface{}) {
	t.locate().Infoln(msg...)
}

func (t *BaseResource) Debug(msg ...interface{}) {
	t.locate().Debugln(msg...)
}

func (t *BaseResource) Error(msg ...interface{}) {
	t.locate().Errorln(msg...)
}

func (t *BaseResource) Warn(msg ...interface{}) {
	t.locate().Warnln(msg...)
}

func (t *BaseResource) Panic(msg ...interface{}) {
	t.locate().Panicln(msg...)
}

func (t *BaseResource) Infof(format string, msg ...interface{}) {
	t.locate().Infof(format, msg...)
}

func (t *BaseResource) Debugf(format string, msg ...interface{}) {
	t.locate().Debugf(format, msg...)
}

func (t *BaseResource) Errorf(format string, msg ...interface{}) {
	t.locate().Errorf(format, msg...)
}

func (t *BaseResource) Warnf(format string, msg ...interface{}) {
	t.locate().Warnf(format, msg...)
}

func (t *BaseResource) Panicf(format string, msg ...interface{}) {
	t.locate().Panicf(format, msg...)
}
