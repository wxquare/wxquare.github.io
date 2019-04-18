---
title: Python S3对象存储
categories:
- Python
---

## 一、s3基本概念
　　最近在项目中处理除了使用数据库存储数据，还接触到对象存储的概念，感觉还挺好用的。它能方便的解决海量图片、小视频、语音等小文件的存储。调查后发现，目前市场上存在许多对象存储的云服务，比较有名是亚马逊的S3。第一次使用S3对象存储，有几个概念需要明白：

- Accesskey：开发者拥有的身份识别的ID，可以理解为用户名，申请服务时平台提供
- Secretkey：开发者用户对应的密钥，可以理解为密码
- Bucket： 存储桶，S3中用于存储数据的容器，每个对象都存储在一个bucket中。
- Object： 对象指的是存储在S3中具体的文件。


## 二、安装bobo sdk
　　为了方便项目的部署，习惯性将项目依赖的sdk直接安装在项目根目录的3rdparty目录下面：
```
	pip install --target=./3rdparty bobo
```

## 三、s3 对象存储基本使用
```
import sys
sys.path.append('./3rdparty')

import boto
from boto.s3.key import Key


class RWGClient(object):
	'''
		http://boto.cloudhackers.com/en/latest/s3_tut.html#creating-a-connection
	'''

	def　__init__(self,access_key,secret_key,host_name,bucket_name,is_secure=False):
		self.aws_access_key_id = access_key
		self.aws_secret_access_key = secret_key
		self.host = host_name
		self.bucket = bucket_name

	def __get_connection():
		return boto.connect_s3(
			aws_access_key_id = self.aws_access_key_id,
			aws_secret_access_key = self.aws_secret_access_key,
			host = self.host
			is_secure = False
		)


	def upload_file_to_rgw(local_file,rgw_file):

	'''
		upload local_file to rgw_file
		todo: deal with error ??? 
	'''
		c = self.__get_connection()
		b = c.get_bucket(self.bucket)
		k = Key(b)
		k.key = rgw_file
		k.set_contents_from_filename(local_file)



	def download_from_rgw(rgw_file,local_file):
	'''
		download a rgw_file to localfile
		todo: deal with error??
	'''
		c = self.__get_connection()
		b = c.get_bucket(self.bucket)
		k = Key(b)
		k.key = rgw_file
		k.get_contents_to_filename(local_file)
		

```


## 四、对象存储使用问题汇总