package main

import (
	"bufio"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	. "pt"
)

func main() {
	args := os.Args
	if len(args) == 2 {
		switch args[1] {
		case "1":
			Cu.Run()
		case "2":
			//平台负债表数据静态化
			Fz.Run()
		case "0":
			os.Exit(0)
		default:
		}
	}
	if len(args) == 1 {
		for {
			fmt.Println("操作目录: ")
			fmt.Println("1、平台有效客户更新（202:13306-platform）。 ")
			fmt.Println("2、平台负债数据静态化（202:13306-platform）。")
			fmt.Println("0、退出。 ")
			inputReader := bufio.NewReader(os.Stdin)
			command, _, _ := inputReader.ReadLine()
			code := string(command)
			switch code {
			case "1":
				Cu.Run()
			case "2":
				//平台负债表数据静态化
				Fz.Run()
			case "0":
				os.Exit(0)
			default:
				fmt.Println("default")
			}
			fmt.Println("-------处理完成-------")
		}
	}

}
