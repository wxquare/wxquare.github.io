---
title: 基于Github双分支和Hexo搭建博客
---
## 一.前言
   Github+Hexo搭建个人博客已经是非常成熟的技能了。在博客园有个博客，写的很烂，这次决定搭建一个自己的博客。根据网络上的教程，搭建的过程不是很复杂，但是也遇到一些问题，比如因为github pags和hexo版本的原因页面无法打开的情况。最重要的是很多教程没有考虑到本地hexo博客备份的内容，**本文通过建立两个分支，一个用于存储静态网页，一个用于备份博客方式，方便在不同的机器上写作。**
## 二.基本环境搭建
1. [安装git](https://git-scm.com/downloads)
2. [安装Nodejs](https://nodejs.org/en/)
3. 在本地建立博客目录，例如/github/wxquare.github.io，以下操作都是在该目录下面进行
4. 进入博客目录，执行下面几条命令后，完成本地博客的构建，浏览器输入http://localhost:4000/
``` bash
$ npm i -g hexo  安装hexo
$ hexo init      初始化  
$ hexo generate  生成静态网页
$ hexo server    启动服务器
```
5. 在Github上建立wxquare.github.io仓库,**创建master和hexo两个分支，master用于存储静态网页，hexo备份本地博客的内容**.
6. 执行下面的命令，初始化本地仓库，并且关联到Github上的wxquare.github.io仓库建立.
    git init
    git remote add origin git@github.com:wxquare/wxquare.github.io.git
    git pull origin master
    git branch hexo
    git checkout hexo
    git pull origin hexo
7. 配置**站点配置文件**_config.yml，将本地博客迁移到github上
	deploy: 
	 type: git
	 repo: https://github.com/wxquare/wxquare.github.io.git
	 branch: master
8. 执行*npm install hexo-deployer-git --save*，然后*hexo deploy*部署到github，浏览器输入https://wxquare.github.io/ 查看
9. 将本地的hexo博客备份到hexo分支上去 
   `git checkout hexo;git add .;git commit -m "init";git push origin hexo`
10. 写博客的基本流程是用markdown写文章，保存到/source/_posts目录下,执行`hexo clean`;`hexo generate`;`hexo server`；查看本地的博客，然后部署到github上去`hexo deploy`，最后备份本地hexo博客`git add .`;`git commit -m "add Github双分支+hexo搭建个人博客"`；`git push orign hexo`

## 三.博客个性化设置
1. 主题设置和修改。hexo初始化默认的主题是landscape，[https://hexo.io/themes/](https://hexo.io/themes/)提供了许多的主题，根据喜好为博客的主题，主题的文档提供了使用方法，设置相应的参数，调整为自己喜欢的格式。我这里选择的Next主题。
2. 不断更新....
