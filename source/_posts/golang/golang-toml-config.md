---
title: golang toml 配置文件
categories:
- Golang
---

　　对于完整的项目和服务来说都是需要配置文件的，json文件格式要求严格，不支持注释。最近在项目中使用的toml配置文件感觉还挺方便，它非常方便人工阅读，同时支持丰富的类型、层次结构和注释。toml文件的具体说明，参考https://github.com/toml-lang/toml。 目前golang有多种对toml文件的解释器(parser)，项目中使用的是https://github.com/BurntSushi/toml。 下面通过一个例子来说明是怎么使用toml配置文件的，具体步骤如下：
- 根据toml文件格式说明，定义项目所需要的配置文件，结构尽可能清晰，不要有太多的层次
- golang结构体定义toml文件对应的结构
- 使用toml parser解释器将toml文件的内容映射到结构体中

config.toml:

	# This is a TOML document.

	title = "TOML Example"

	[owner]
	name = "Tom Preston-Werner"
	dob = 1979-05-27T07:32:00-08:00 # First class dates

	[database]
	server = "192.168.1.1"
	ports = [ 8001, 8001, 8002 ]
	connection_max = 5000
	enabled = true

	[servers]

	  # Indentation (tabs and/or spaces) is allowed but not required
	  [servers.alpha]
	  ip = "10.0.0.1"
	  dc = "eqdc10"

	  [servers.beta]
	  ip = "10.0.0.2"
	  dc = "eqdc10"

	[clients]
	data = [ ["gamma", "delta"], [1, 2] ]

	# Line breaks are OK when inside arrays
	hosts = [
	  "alpha",
	  "omega"
	]


config.go:
```
package main

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

//https://github.com/BurntSushi/toml.git

type tomlConfig struct {
	Title   string
	Owner   ownerInfo
	DB      database `toml:"database"`
	Servers map[string]server
	Clients clients
}

type ownerInfo struct {
	Name string
	Org  string `toml:"organization"`
	Bio  string
	DOB  time.Time
}

type database struct {
	Server  string
	Ports   []int
	ConnMax int `toml:"connection_max"`
	Enabled bool
}

type server struct {
	IP string
	DC string
}

type clients struct {
	Data  [][]interface{}
	Hosts []string
}

var (
	cfg  *tomlConfig
	once sync.Once
)

func Config() *tomlConfig {
	once.Do(func() {
		filePath, err := filepath.Abs("config.toml")
		if err != nil {
			panic(err)
		}
		fmt.Printf("parse toml file once. filePath: %s\n", filePath)
		if _, err := toml.DecodeFile(filePath, &cfg); err != nil {
			panic(err)
		}
	})
	return cfg
}
```

main.go:
```
package main

import (
	"fmt"
)

func main() {
	conf := Config()
	fmt.Printf("%+v", conf)
}
```



