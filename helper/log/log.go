package log

/*
 * 日志相关
 */
import (
	"log"
	"os"
)

var spiderLog *log.Logger

func InitLog(path string) {
	f, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE | os.O_APPEND , 0755)
	if err != nil {
		log.Fatal(err)
	}
	spiderLog = log.New(f, "", log.LstdFlags)
}

func Infof(format string, v ...interface{}) {
	spiderLog.Printf("[INFO] "+format, v...)
}

func Infoln(v ...interface{}) {
	v1 := []interface{}{"[INFO]"}
	v1 = append(v1, v...)
	spiderLog.Println(v1...)
}

func Info(v ...interface{}) {
	v1 := []interface{}{"[INFO]"}
	v1 = append(v1, v...)
	spiderLog.Print(v1...)
}

func Warnf(format string, v ...interface{}) {
	spiderLog.Printf("[WARN] "+format, v...)
}

func Warnln(v ...interface{}) {
	v1 := []interface{}{"[WARN]"}
	v1 = append(v1, v...)
	spiderLog.Println(v1...)
}

func Warn(v ...interface{}) {
	v1 := []interface{}{"[WARN]"}
	v1 = append(v1, v...)
	spiderLog.Print(v1...)
}

func Fatalf(format string, v ...interface{}) {
	spiderLog.Fatalf("[WARN] "+format, v...)
}

func Fatalln(v ...interface{}) {
	v1 := []interface{}{"[FATAL]"}
	v1 = append(v1, v...)
	spiderLog.Fatalln(v1...)
}

func Fatal(v ...interface{}) {
	v1 := []interface{}{"[FATAL]"}
	v1 = append(v1, v...)
	spiderLog.Fatal(v1...)
}
