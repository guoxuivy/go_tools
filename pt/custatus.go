// 客户属性自动更新封装
// 需要公司开启自动更新并配置客户过期周期时间
package pt

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type CuStatusClass struct {
}

func NewCu() *CuStatusClass {
	obj := new(CuStatusClass)
	return obj
}

/**
 *判断客户状态
 **/
func (obj *CuStatusClass) getState(orders int, practs int, state int) int {
	if orders > 0 || practs > 0 {
		return 1 //最近有订单OR有实操 为有效客户
	}
	if state == 3 {
		return 3 //死档客户
	}
	if state == -1 {
		return -1 //无效客户
	}
	//默认返回为久党客户
	return 2
}

/*
 * 获取公司关于有效客户的配置天数 默认30天
 */
func (obj *CuStatusClass) getConfig(str string) int {
	var num int
	n := strings.Index(str, "member_config")
	if n == -1 {
		num = 30
	} else {
		start := n + 20
		end := n + 22
		num2 := string([]byte(str)[start:end])
		num, _ = strconv.Atoi(num2)
	}
	return num
}

/**
 * 更新一个公司的客户状态 (PT) 考虑新建数据库连接 提高效率
 * @param  {[type]} db      *sql.DB       [description]
 * @param  {[type]} c       chan          int           [description]
 * @param  {[type]} comp_id int           [公司ID]
 * @param  {[type]} num int           	  [有效期天数]
 * @return {[type]}         			  [description]
 */
func (obj *CuStatusClass) updateOneComp(c chan string, comp_id int, num int) {
	db := Mydb()
	defer db.Close()
	end := time.Now().Unix()
	start := end - 3600*24*int64(num) //前推num天
	sql := "SELECT a.id,a.name,a.status,(SELECT COUNT(*) FROM `order_sale` WHERE `cu_id` = a.id AND `pay_time` > ? AND `pay_time` < ? AND `type` IN (1,2)) AS orders, (SELECT COUNT(*) FROM `practice_order` WHERE `cu_id` = a.id AND `pay_time` > ? AND `pay_time` < ?) AS practs FROM `customer_info` AS a LEFT JOIN `config_membership` AS m ON a.membership_id = m.id WHERE m.`is_member` = 1 AND a.`company_id` = ?"
	rows, err := db.Query(sql, start, end, start, end, comp_id)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	var id int
	var orders int
	var practs int
	var name string
	var status int
	for rows.Next() {
		err := rows.Scan(&id, &name, &status, &orders, &practs)
		if err != nil {
			log.Fatal(err)
		}

		new_status := obj.getState(orders, practs, status)
		if status != new_status {
			stmt, err := db.Prepare("UPDATE `customer_info` SET `status`=? WHERE `id`=?")
			defer stmt.Close()
			if err != nil {
				log.Println(err)
				return
			}
			stmt.Exec(new_status, id)
			//返回管道信息写入
			c <- fmt.Sprintf("%d", comp_id) + ":" + fmt.Sprintf("%d", id) + " " + name + " " + fmt.Sprintf("%d", status) + "->" + fmt.Sprintf("%d", new_status)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	close(c)
}

/**
 * 多公司并发处理 (PT)
 * @param  {[type]} db *sql.DB       [description]
 * @return {[type]}    [description]
 */
func (obj *CuStatusClass) Run() {
	db := Mydb()
	defer db.Close()
	sql := "SELECT id , auto_cu_status, config FROM `company_info` WHERE `status` = 1"
	rows, err := db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	chs := make([]chan string, 0) //开多个管道接受消息
	var id int
	var auto int
	var config string
	for rows.Next() {
		err := rows.Scan(&id, &auto, &config)
		if err != nil {
			log.Fatal(err)
		}
		if auto == 1 {
			num := obj.getConfig(config) //客户有效期设置
			c := make(chan string)
			chs = append(chs, c)
			go obj.updateOneComp(c, id, num)
		}

	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	for _, ch := range chs { //多管道写法
		for {
			x, ok := <-ch
			if ok == false {
				break
			}
			writeResult("cu_status", x)
			fmt.Println(x) //消息回收处理 可扩展写入文件日志
		}
	}

}

var Cu *CuStatusClass

func init() {
	Cu = NewCu()
}
