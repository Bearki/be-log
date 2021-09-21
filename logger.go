/**
 *@Title belog核心代码
 *@Desc belog日志的主要实现都在这里了，欢迎大家指出需要改进的地方
 *@Author Bearki
 *@DateTime 2021/09/21 19:16
 */

package belog

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Logger 日志接口
type Logger interface {
	Trace(format string, val ...interface{}) // 通知级别的日志
	Debug(format string, val ...interface{}) // 调试级别的日志
	Info(format string, val ...interface{})  // 普通级别的日志
	Warn(format string, val ...interface{})  // 警告级别的日志
	Error(format string, val ...interface{}) // 错误级别的日志
	Fatal(format string, val ...interface{}) // 致命级别的日志
}

// 日志引擎类型
type belogEngine uint8

// 日志级别类型
type belogLevel uint8

// 日志级别字符类型
type belogLevelChar byte

// 日志输出引擎方法类型
type printFuncEngine func(logStr string)

// beLog 记录器对象
type beLog struct {
	logPath    string                          // 日志文件保存路径
	isSplitDay bool                            // 是否开启按日分割
	isFileLine bool                            // 是否开启文件行号记录
	maxSize    uint16                          // 单文件最大容量（单位：byte）
	saveDay    uint16                          // 日志保存天数
	skip       uint                            // 需要向上捕获的函数栈层数（该值会自动加2，以便于实例化用户可直接使用）【0-runtime.Caller函数的执行位置（在belog包内），1-Belog各级别方法实现位置（在belog包内），2-belog实例调用各级别日志函数位置，依次类推】
	engine     map[belogEngine]printFuncEngine // 输出引擎方法类型映射
	level      map[belogLevel]belogLevelChar   // 需要记录的日志级别字符映射
}

// 记录引擎定义
var (
	EngineConsole belogEngine = 1 // 控制台引擎
	EngineFile    belogEngine = 2 // 文件引擎
)

// 日志引擎映射
var engineMap = map[belogEngine]printFuncEngine{
	EngineConsole: printConsoleLog, // 控制台输出方法映射
	EngineFile:    printFileLog,    // 文件输出方法映射
}

// 日志保存级别定义
var (
	LevelTrace belogLevel = 1 // 通知级别
	LevelDebug belogLevel = 2 // 调试级别
	LevelInfo  belogLevel = 3 // 普通级别
	LevelWarn  belogLevel = 4 // 警告级别
	LevelError belogLevel = 5 // 错误级别
	LevelFatal belogLevel = 6 // 致命级别
)

// 日志级别映射
var levelMap = map[belogLevel]belogLevelChar{
	1: 'T',
	2: 'D',
	3: 'I',
	4: 'W',
	5: 'E',
	6: 'F',
}

// New 初始化一个日志记录器实例
// @params engine      belogEngine    必选的日志引擎
// @params otherEngine ...belogEngine 其他日志引擎
// @params logPath string 日志文件保存地址(可选)
// @return         *beLog 日志记录器实例指针
// @return         error  初始化时发生的错误信息
func New() *beLog {
	// if len(logPath) > 0 {
	// 	// 判断是文件路径还是文件夹路径
	// 	filePath, err := os.Stat(logPath[0])
	// 	if err != nil {
	// 		fmt.Printf("beLog: %s\n", err.Error())
	// 		return nil
	// 	}
	// 	if filePath.IsDir() {
	// 		fmt.Printf("beLog: %s is dir, `logPath` it should be a file\n", logPath[0])
	// 		return nil
	// 	}
	// }

	// 初始化日志记录器对象
	belog := new(beLog)
	// 赋值默认值
	belog.SetEngine(EngineConsole).
		SetLevel(LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal).
		SetMaxSize(100).
		SetSaveDay(7).
		SetSkip(0)
	return belog
}

// SetEngine 设置日志记录引擎
// @params val ...belogEngine 任意数量的日志记录引擎
func (belog *beLog) SetEngine(val ...belogEngine) *beLog {
	// 判断引擎是否为空
	if belog.engine == nil {
		// 初始化一下
		belog.engine = make(map[belogEngine]printFuncEngine)
	}
	// 遍历输入的引擎
	for _, item := range val {
		belog.engine[item] = engineMap[item]
	}
	return belog
}

// SetLevel 设置日志记录保存级别
// @params val ...belogLevel 任意数量的日志记录级别
func (belog *beLog) SetLevel(val ...belogLevel) *beLog {
	// 判断级别映射是否为空
	if belog.level == nil {
		// 初始化一下
		belog.level = make(map[belogLevel]belogLevelChar)
	}
	// 遍历输入的级别
	for _, item := range val {
		belog.level[item] = levelMap[item]
	}
	return belog
}

// OpenSplitDay 开启日志文件按日分割
func (belog *beLog) OpenSplitDay() *beLog {
	belog.isSplitDay = true
	return belog
}

// OpenFileLine 开启文件行号记录
func (belog *beLog) OpenFileLine() *beLog {
	belog.isFileLine = true
	return belog
}

// SetMaxSize 配置单文件储存容量
// @params maxSize uint16 单文件最大容量（单位：MB）
func (belog *beLog) SetMaxSize(maxSize uint16) *beLog {
	byteSize := 1024 * 1024
	belog.maxSize = uint16(byteSize) * maxSize
	return belog
}

// SetSaveDay 配置日志保存天数
// @params saveDay uint16 保存天数
func (belog *beLog) SetSaveDay(saveDay uint16) *beLog {
	belog.saveDay = saveDay
	return belog
}

// SetSkip 配置需要向上捕获的函数栈层数
func (belog *beLog) SetSkip(skip uint) *beLog {
	belog.skip = 2 + skip
	return belog
}

// print 日志集中打印地，日志的真实记录地
func (belog *beLog) print(logstr string, levelChar belogLevelChar) {
	// 统一当前时间
	currTime := time.Now()
	// 是否需要打印文件行数
	if belog.isFileLine {
		_, file, line, _ := runtime.Caller(int(belog.skip))
		// 格式化
		logstr = fmt.Sprintf(
			"%s.%03d [%s] [%s:%d] %s\n",
			currTime.Format("2006/01/02 15:04:05"),
			currTime.UnixMilli()%currTime.Unix(),
			string(levelChar),
			filepath.Base(file),
			line,
			logstr,
		)
	} else {
		// 格式化
		logstr = fmt.Sprintf(
			"%s.%03d [%s] %s\n",
			currTime.Format("2006/01/02 15:04:05"),
			currTime.UnixMilli()%currTime.Unix(),
			string(levelChar),
			logstr,
		)
	}
	// 异步等待组
	var wg sync.WaitGroup
	// 遍历引擎，执行输出
	for _, printFunc := range belog.engine {
		wg.Add(1)
		go func(ouput printFuncEngine) {
			defer wg.Done()
			ouput(logstr)
		}(printFunc)
	}
	wg.Wait()
}

// Trace 通知级别的日志
// @params format string         序列化格式
// @params val    ...interface{} 待序列化内容
func (belog *beLog) Trace(format string, val ...interface{}) {
	// 判断当前级别日志是否需要记录
	if _, ok := belog.level[LevelTrace]; !ok { // 当前级别日志不需要记录
		return
	}
	// 执行日志记录
	logStr := fmt.Sprintf(format, val...)
	belog.print(logStr, belog.level[LevelTrace])
}

// Debug 调试级别的日志
// @params format string         序列化格式
// @params val    ...interface{} 待序列化内容
func (belog *beLog) Debug(format string, val ...interface{}) {
	// 判断当前级别日志是否需要记录
	if _, ok := belog.level[LevelDebug]; !ok { // 当前级别日志不需要记录
		return
	}
	// 执行日志记录
	logStr := fmt.Sprintf(format, val...)
	belog.print(logStr, belog.level[LevelDebug])
}

// Info 普通级别的日志
// @params format string         序列化格式
// @params val    ...interface{} 待序列化内容
func (belog *beLog) Info(format string, val ...interface{}) {
	// 判断当前级别日志是否需要记录
	if _, ok := belog.level[LevelInfo]; !ok { // 当前级别日志不需要记录
		return
	}
	// 执行日志记录
	logStr := fmt.Sprintf(format, val...)
	belog.print(logStr, belog.level[LevelInfo])
}

// Warn 警告级别的日志
// @params format string         序列化格式
// @params val    ...interface{} 待序列化内容
func (belog *beLog) Warn(format string, val ...interface{}) {
	// 判断当前级别日志是否需要记录
	if _, ok := belog.level[LevelWarn]; !ok { // 当前级别日志不需要记录
		return
	}
	// 执行日志记录
	logStr := fmt.Sprintf(format, val...)
	belog.print(logStr, belog.level[LevelWarn])
}

// Error 错误级别的日志
// @params format string         序列化格式
// @params val    ...interface{} 待序列化内容
func (belog *beLog) Error(format string, val ...interface{}) {
	// 判断当前级别日志是否需要记录
	if _, ok := belog.level[LevelError]; !ok { // 当前级别日志不需要记录
		return
	}
	// 执行日志记录
	logStr := fmt.Sprintf(format, val...)
	belog.print(logStr, belog.level[LevelError])
}

// Fatal 致命级别的日志
// @params format string         序列化格式
// @params val    ...interface{} 待序列化内容
func (belog *beLog) Fatal(format string, val ...interface{}) {
	// 判断当前级别日志是否需要记录
	if _, ok := belog.level[LevelFatal]; !ok { // 当前级别日志不需要记录
		return
	}
	// 执行日志记录
	logStr := fmt.Sprintf(format, val...)
	belog.print(logStr, belog.level[LevelFatal])
}
