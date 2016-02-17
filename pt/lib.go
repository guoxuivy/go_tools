package pt

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

const (
	DSN string = "root:mysqladmin56@tcp(192.168.0.202:13306)/howa_platform_prod?charset=utf8"
	//DSN string = "root:@tcp(127.0.0.1:3306)/db_name?charset=utf8"
)

/**
 * 数据库连接
 */
func Mydb() *sql.DB {
	db, err := sql.Open("mysql", DSN)
	if err != nil {
		log.Fatalf("Open database error: %s\n", err)
	}
	//defer db.Close() //不关闭连接
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

/**
 * 升级日志写入  文件追加参数奇葩 多 要3个
 * @param  {[type]} log string        [description]
 * @return {[type]}     [description]
 */
func writeResult(tag string, data string) {
	str_time := time.Now().Format("2006_01_02")
	filename := tag + "_" + str_time + ".log"
	fl, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Println(err)
	}
	defer fl.Close()
	fl.WriteString(data)
	fl.WriteString("\n")
}
