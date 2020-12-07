---
title: golang mysql以及链接池的实现
categories:
- Golang
---

## 一、接口和依赖第三方包
- 标准库：[database/sql](https://golang.org/pkg/database/sql/)
- 数据库驱动(driver): [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- 第三方扩展包：[github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)

### database/sql
　　Golang提供了database/sql包用于对sql数据库的访问，它提供操作数据库的入口对象**sql.DB**。sql.DB表示操作数据库的抽象访问接口。sql包提供了操作数据库所有必要的结构体、函数和方法，sql包使得从一个数据库迁移到另一个数据库变得容易，只需更换一个驱动包即可。例如从sql server迁移到Mysql。

### github.com/go-sql-driver/mysql
　　Golang操作数据库需要安装第三方的数据库驱动包，例如mysql的github.com/go-sql-driver/mysql。Golang提供的database/sql/driver定义了数据库驱动的所有的接口。下面是目前Golang支持的数据库驱动的列表：https://github.com/golang/go/wiki/SQLDrivers

### sqlx第三方包
　　sqlx是针对database/sql包扩展，使得golang对数据库的访问更加方便。比较database/sql和sqlx的文档，可以发现sqlx尽可能保留了sql包功能的同时也扩展了更加方便的接口:
- sqlx文档：https://godoc.org/github.com/jmoiron/sqlx
- sql标准库文档：https://golang.org/pkg/database/sql/

## 二、基本用法

**新建sqlx.DB**:
```
package mysql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

// Config is mysql config.
type Config struct {
	Host        string
	Port        int
	User        string
	Pass        string
	Database    string
	Charset     string
	Active      int           //SetMaxOpenConns,recommendation 100
	Idle        int           //SetMaxIdleConns,recommendation 2
	IdleTimeout time.Duration //SetConnMaxLifetime,recommendation 5 second
}

// New a sqlx.DB.
func NewMysqlSqlDB(c *Config) (*sqlx.DB, error) {
	if c.Charset == "" {
		c.Charset = "utf-8"
	}
	DSN := fmt.Sprint(c.User, ":", c.Pass, "@tcp(", c.Host, ":", c.Port, ")/", c.Database, "?charset=", c.Charset)
	driverName := "mysql"
	db, err := sqlx.Connect(driverName, DSN)
	if err != nil {
		return db, err
	}
	if c.Active != 0 {
		db.SetMaxOpenConns(c.Active) //设置连接池最大打开数据库连接数，<=0表示不限制打开连接数，默认为0

	}
	if c.Idle != 0 {
		db.SetMaxIdleConns(c.Idle) //<=0表示不保留空闲连接，默认值2
	}

	if c.IdleTimeout != 0 {
		db.SetConnMaxLifetime(c.IdleTimeout) //设置连接超时时间
	}
	return db, err
}
```



我们通常将数据库查询逻辑封装在dao中，它sqlx.DB是他的成员：
```
package mysql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Dao struct {
	db *sqlx.DB
}

//New a Dao.
func NewDao(db *sqlx.DB) *Dao {
	d := &Dao{
		db: db,
	}
	return d
}
```


### select:
1. 查询一条数据，QueryRowx
2. 查询多条数据，QueryRowx,rows.Next(),rows.Close()
3. 解析少数几个字段，row.Scan，rows.Scan
4. 按照结构体解析，row,StructScan,rows.StructScan

```
func (self *Dao) QueryXX() {
	sql := "SELECT * FROM users WHERE id=?"
	var (
		id     int
		name   string
		salary int
	)
	if err := self.db.QueryRowx(sql, id).Scan(&id, &name, &salary); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(id, name, salary)
	}
	/*
		type Employee struct {
			id     int
			name   string
			salary int
		}
		employeeInfo := &Employee{}
		if err := self.db.QueryRowx(sql, id).StructScan(employeeInfo); err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(employeeInfo)
		}
	*/
}

//query
func (self *Dao) QueryXXXX() {
	sql := "SELECT * FROM users WHERE id < ?"
	rows, err := self.db.Queryx(sql, 10)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id     int
			name   string
			salary int
		)
		err = rows.Scan(&id, &name, &salary)
		if err != nil {
			fmt.Println(err)
		}
	}
	/*
		// rows.StructScan
		type Employee struct {
			id     int
			name   string
			salary int
		}
		employeeInfo := &Employee{}
		for rows.Next() {
			err = rows.StructScan(employeeInfo)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(employeeInfo.id, employeeInfo.name, employeeInfo.salary)
			}
		}
	*/

}
```


### insert、update，delete：
　　Exec()或ExecContext()方法的第一个返回值为一个实现了sql.Result接口的类型，我们可以用sql.Result中的LastInsertId()方法或RowsAffected()来判断SQL语句是否执行成功。注意LastInsertId()方法只有在使用INSERT语句且数据表有自增id时才有返回自增id值，否则返回0。
sql.Result的定义如下：
```
	type Result interface {
	    LastInsertId() (int64, error)//使用insert向数据插入记录，数据表有自增id时，该函数有返回值
	    RowsAffected() (int64, error)//表示影响的数据表行数
	}
```
```
func (self *Dao) InsertXX() {
	sql := "INSERT INTO users values(?,?,?)"
	rs, err := self.db.Exec(sql, 4, "yyy", 1000)
	if err != nil {
		fmt.Println(err)
		return
	}
	if id, _ := rs.LastInsertId(); id > 0 {
		fmt.Println("insert success")
	}
	/*也可以这样判断是否插入成功
	  if n,_ := rs.RowsAffected();n > 0 {
	      fmt.Println("insert success")
	  }
	*/
}
```

### 预编译：Prepared Statements
sql语句在db接收到最终执行完毕返回会经历三个过程：
- 词法和语义分析
- 优化sql语句，指定执行计划
- 执行并返回结果
　　有些时候，我们的一条sql语句可能会反复执行，或者每次执行的时候只有个别的值不同（比如query的where子句值不同，update的set子句值不同,insert的values值不同）。如果每次都需要经过上面的词法语义解析、语句优化、制定执行计划等，则效率就明显不行了。所谓预编译语句就是将这类语句中的值用占位符替代，可以视为将sql语句模板化或者说参数化，一般称这类语句叫Prepared Statements或者Parameterized Statements预编译语句的优势在于归纳为：**一次编译、多次运行，省去了解析优化等过程；此外预编译语句能防止sql注入**。

```
//Prepared Statements
//select,insert,update,delete
func (self *Dao) PreparedStms() {
	stmt, err := self.db.Prepare("SELECT * FROM users WHERE id = ?")
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(2)
	defer rows.Close()
	for rows.Next() {
		var (
			id       int
			username string
			salary   int
		)
		err = rows.Scan(&id, &username, &salary)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(id, username, salary)
		}

	}
}

```

### 事务transaction
　　默认会把提交的每一条SQL语句都当作一个事务来处理，如果多条语句一起执行，当其中某个语句执行错误，则前面已经执行的SQL语句无法回滚。对于一些要求比较严格的业务逻辑来说(如订单付款、用户转账等)，应该在同一个事务提交多条SQL语句，避免发生执行出错无法回滚事务的情况。
事务的隔离级别：
- read uncommitted （可以读取其它事务未提交的数据，造成脏读的问题）
- read committed （可以读取其它事务提交事务提交的数据，可能造成该事务两次读取的数据不一样，即不可重复读
- repeatable read（保证事务多次读取的数据相同，但是可能会造成幻读，例如该事物尝试插入数据是，因为别的事务插入数据导致插入失败，而该事务本身却很难发现）
- serializable（串行化，当把当前的会话设置为serializable时，其它会话对该表的写操作会被挂起。

```
//transactions
func (self *Dao) txOps() {
	tx, _ := self.db.Beginx()
	rs, err := tx.Exec("UPDATE users SET username = ? WHERE id = ?", "aaaa", 2)
	if err != nil {
		fmt.Println(err)
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
	}
	if n, _ := rs.RowsAffected(); n > 0 {
		fmt.Println("txops success!")
	}

}
```

参考：
https://juejin.im/post/5cb94e3a5188251ad954e6f7#heading-13
https://www.cnblogs.com/hanyouchun/p/6708037.html


