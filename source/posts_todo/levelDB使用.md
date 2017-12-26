# leveldb简介
leveldb是google用C++开发的快速的键值对（KV）存储数据库，提供从字符串到字符串的有序映射。具有以下几个特点:     

- k值和v值支持任意长度的字节数组
- 数据以k值的字典序有序存储
- 用户可以自定义k值的比较函数,感觉和STL中的map非常类似
- 基本的操作Put(key,value),Get(key),Delete(key)
- 多个操作可以使用批处理操作，批处理操作是原子
- 数据使用Google的Snappy压缩库自动压缩
- 它不是一个SQL数据库，不支持SQL查询，不是关系型的数据模型，不支持索引等技术
- 没有内置的客户端和服务器模块

# leveldb下载安装
	git clone https://github.com/google/leveldb.git //下载leveldb
	cd leveldb    //编译leveldb，在out-shared，out-static下生成动态库和静态库
	make
	#安装
	sudo cp -r /leveldb/include/leveldb /usr/local/include
	sudo cp /leveldb/out-shared/libleveldb.so.1.20 /usr/lib
    sudo ln -s /usr/lib/libleveldb.so.1.20 /usr/lib/libleveldb.so.1
    sudo ln -s /usr/lib/libleveldb.so.1.20 /usr/lib/libleveldb.so
    sudo ldconfig  #更新lib

# leveldb简单使用	
	#include <iostream>
	#include <cassert>
	#include <cstdlib>
	#include <string>
	// 包含必要的头文件
	#include <leveldb/db.h>

	using namespace std;

	int main(void)
	{
	    // 数据库性能参数设置
	    leveldb::Options options;
	    options.create_if_missing = true; // 如果数据库不存在就创建
	
	    //创建数据库或者打开数据库
	    leveldb::DB *db = nullptr;
	    leveldb::Status status = leveldb::DB::Open(options, "./testdb", &db);
	    assert(status.ok());
	
	    std::string key = "key1";
	    std::string value = "value1";
	    std::string get_value;
	    // 写入 key1 -> value1
	    status = db->Put(leveldb::WriteOptions(), key, value);
	
	    // 写入成功，就读取 key:people 对应的 value
	    if (status.ok())
	        status = db->Get(leveldb::ReadOptions(), "key1", &get_value);
	
	    // 读取成功就输出
	    if (status.ok())
	        cout << get_value << endl;
	    else
	        cout << status.ToString() << endl;
	
	    delete db;
	    return 0;
	}

使用静态库编译链接：
	g++ -std=c++11 testleveldb.cpp -o testleveldb ./lib/libleveldb.a -lpthread
# leveldb封装