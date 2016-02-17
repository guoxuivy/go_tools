// 负债数据静态化封装
package pt

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

//门店负债数据
type FzData struct {
	mdid  int
	xjye  float32
	zsye  float32
	lcye  float32
	month string
}

type FzDataMap map[int]FzData

type FzClass struct {
	list FzDataMap
}

//list 使用前需要被初始化  所以直接在new时初始化 以下2种初始化方式都可
func NewFz() *FzClass {
	fz := &FzClass{list: make(FzDataMap)}
	//fz := new(FzClass)
	//fz.list = make(FzDataMap)
	return fz
}

func (fz *FzClass) Show() *FzClass {
	for i, value := range fz.list {
		fmt.Println(i)
		fmt.Println(value)
	}
	return fz
}

//暂无卵用 方便大数据扩展
func (fz *FzClass) Add(row FzData) *FzClass {
	fz.list[row.mdid] = row
	return fz
}

//门店数据静态化 入库
func (fz *FzClass) toDb(db *sql.DB, row FzData) {
	var num int
	one, err := db.Query("SELECT  COUNT(*) AS num  FROM `static_fz` WHERE `md_id` = ? AND `month` = ? ", row.mdid, row.month)
	if err != nil {
		log.Println(err)
	}
	defer one.Close()
	for one.Next() {
		err := one.Scan(&num)
		if err != nil {
			log.Fatal(err)
		}
	}
	if num > 0 {
		//存在更新
		stmt, _ := db.Prepare("UPDATE `static_fz` SET `xjye`=?, `zsye`=?, `lcye`=? WHERE `md_id` = ? AND `month` = ? ")
		defer stmt.Close()
		stmt.Exec(row.xjye, row.zsye, row.lcye, row.mdid, row.month)
	} else {
		//不存在插入
		stmt, _ := db.Prepare("INSERT INTO `static_fz` (`md_id`, `xjye`, `zsye`, `lcye`,`month`) VALUES (?,?,?,?,?)")
		defer stmt.Close()
		stmt.Exec(row.mdid, row.xjye, row.zsye, row.lcye, row.month)
	}
}

/**
 * 某一公司负债处理
 * @param  {[type]} c chan int            [日志管道]
 * @param  {[type]} comp_id int           [公司ID]
 * @param  {[type]} fz_month int          [负债月份]
 */
func (fz *FzClass) oneComp(c chan string, comp_id int, fz_month string) {
	db := Mydb()
	defer db.Close()
	sql := "SELECT ed.`id`, SUM(IF(cc.type = 1, t.balance, 0)) AS xjye, SUM(IF(cc.type = 2, t.balance, 0)) AS zsye, IFNULL(tt.`lcye`,0) AS lcye FROM `customer_capital` `t` LEFT JOIN company_capital cc ON cc.id = t.capital_id LEFT JOIN customer_info ci ON ci.id = t.cu_id LEFT JOIN employ_dept ed ON ed.id = ci.store_id LEFT JOIN (SELECT ed.`id`,SUM(TRUNCATE(osd.pay_price / num * t.re_num, 1)) AS lcye FROM `customer_re_project` `t` LEFT JOIN customer_info ci ON ci.id = t.cu_id LEFT JOIN employ_dept ed ON ed.id = ci.store_id LEFT JOIN order_sale_detail osd ON osd.id = t.detail_id WHERE (ed.comp_id = ?) GROUP BY ed.id) tt ON tt.id = ed.id WHERE (ed.comp_id = ?) GROUP BY ed.id"
	rows, err := db.Query(sql, comp_id, comp_id)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	var rowData FzData
	rowData.month = fz_month
	for rows.Next() {
		err := rows.Scan(&rowData.mdid, &rowData.xjye, &rowData.zsye, &rowData.lcye)
		if err != nil {
			log.Fatal(err)
		}
		//fz.Add(rowData)
		fz.toDb(db, rowData)
		//返回管道信息写入
		c <- "公司:" + fmt.Sprintf("%d", comp_id) + "->门店:" + fmt.Sprintf("%d", rowData.mdid) + "(处理完成)"
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	close(c)
}

func (fz *FzClass) Run() {
	db := Mydb()
	defer db.Close()
	sql := "SELECT id FROM `company_info` WHERE `status` = 1"
	rows, err := db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	chs := make([]chan string, 0) //开多个管道接受消息
	fz_month := time.Now().Format("200601")
	var i int = 0
	var id int
	for rows.Next() {
		c := make(chan string)
		chs = append(chs, c)
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		go fz.oneComp(c, id, fz_month)
		i = i + 1
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
			writeResult("fz", x)
			fmt.Println(x) //消息回收处理 可扩展写入文件日志
		}
	}

}

var Fz *FzClass

func init() {
	Fz = NewFz()
}
