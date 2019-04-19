---
title: Python 使用mysql
categories:
- Python
---

## 一、ubuntu安装mysql
安装和运维数据是一件非常麻烦的事情，以后尽量使用云数据库，将时间花在业务和算法上面。
```
	sudo apt-get install mysql-server
	sudo apt install mysql-client
	sudo apt install libmysqlclient-dev
```

Mysql配置文件：
	sudo vi /etc/mysql/mysql.conf.d/mysqld.cnf

查看初始的用户和密码：
sudo cat /etc/mysql/debian.cnf
	# Automatically generated for Debian scripts. DO NOT TOUCH!
	[client]
	host     = localhost
	user     = debian-sys-maint
	password = lla3zI2SYolLD1ww
	socket   = /var/run/mysqld/mysqld.sock
	[mysql_upgrade]
	host     = localhost
	user     = debian-sys-maint
	password = lla3zI2SYolLD1ww
	socket   = /var/run/mysqld/mysqld.sock

Mysql登录：
	mysql -udebian-sys-maint -plla3zI2SYolLD1ww

简单的mysql命令：
	show databases;
	use dbs;
	show tables;

Mysql数据库重启：
	service mysql restart

## 二、 选择pymysql
目前开源的mysql连接工具，这里选择pymysql。
为了方便部署，没有将pymysql安装到系统目录，而是将pymysql安装到工程目录下面。
pip install --target=/home/wxquare/python/projx/3rdparty pymysql

ERROR 1698 (28000): Access denied for user 'root'@'localhost'
https://stackoverflow.com/questions/39281594/error-1698-28000-access-denied-for-user-rootlocalhost

## 三、 pymysql简单使用

```
import sys
sys.path.append('./3rdparty')

import pymysql


class MysqlClient(object):
	def __init__(self,host,port,user,password,db,charset='utf8'):

		self.host = host
		self.port = port
		self.user = user
		self.password = password
		self.db = db
		self.charset = charset


	def get_connection(self):

		return pymysql.connect(
				host = self.host,
				port = self.port,
				user = self.user,
				password = self.password,
				db = self.db,
				charset = self.charset
			)


	def query(self,sql_str):
		'''
			select
		'''
		con = self.get_connection()
		cur = con.cursor(cursor=pymysql.cursors.DictCursor)  # default tuple cursor
		cur.execute(sql_str)
		rows = cur.fetchall() # fetch all data
		cur.close()
		con.close()
		return rows


	def execute(self,sql_str):
		'''
			insert,update,delete  note:INSERT、UPDATE、DELETE neet commit
		'''
		con = self.get_connection()
		cur = con.cursor()
		try:
			cur.execute(sql_str)
			con.commit()
		except Exception:
			con.rollback()
		finally:
			cur.close()
			con.close()
```


## 四、mysql使用问题汇总：
1. navicate无限期试用。https://my.oschina.net/u/3509764/blog/910748


参考：https://shockerli.net/post/python3-pymysql