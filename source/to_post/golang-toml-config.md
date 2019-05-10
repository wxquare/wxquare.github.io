---
title: golang toml 配置文件
categories:
- Golang
---

　　对于完整的项目和服务来说都是需要配置文件的，json文件格式要求严格，不支持注释。最近在项目中使用的toml配置文件感觉还挺方便，它非常方便人工阅读，同时支持丰富的类型、层次结构和注释。toml文件的具体说明，参考https://github.com/toml-lang/toml。 目前golang有多种对toml文件的解释器(parser)，项目中使用的是https://github.com/BurntSushi/toml。 下面通过一个例子来说明是怎么使用toml配置文件的。


