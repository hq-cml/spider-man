package log

/*
 * 日志相关
 */
import (
	"log"
)

func Infof(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

func Infoln(v ...interface{}) {
	v1 := []interface{}{"[INFO]"}
	v1 = append(v1, v...)
	log.Println(v1...)
}

func Info(format string, v ...interface{}) {
	v1 := []interface{}{"[INFO]"}
	v1 = append(v1, v...)
	log.Print(v1...)
}

func Warnf(format string, v ...interface{}) {
	log.Printf("[WARN] "+format, v...)
}

func Warnln(v ...interface{}) {
	v1 := []interface{}{"[WARN]"}
	v1 = append(v1, v...)
	log.Println(v1...)
}

func Warn(format string, v ...interface{}) {
	v1 := []interface{}{"[WARN]"}
	v1 = append(v1, v...)
	log.Print(v1...)
}

func Fatalf(format string, v ...interface{}) {
	log.Fatalf("[WARN] "+format, v...)
}

func Fatalln(format string, v ...interface{}) {
	v1 := []interface{}{"[FATAL]"}
	v1 = append(v1, v...)
	log.Fatalln(v1...)
}

func Fatal(format string, v ...interface{}) {
	v1 := []interface{}{"[FATAL]"}
	v1 = append(v1, v...)
	log.Fatal(v1...)
}
