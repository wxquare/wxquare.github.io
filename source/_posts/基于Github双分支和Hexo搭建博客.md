---
title: 基于Github双分支和Hexo搭建博客
---

## 一.前言 ##
学习和总结是程序员非常重要的能力，本文将结合Hexo和github搭建个人博客，用于总结工作中重要的知识点和难点。整个搭建过程非常简单，网络上也有许多的教程，但是大多数教程都忽略了在不同机器上写博客需要重复环境配置的问题，也忽略了博客内容的备份问题。**本文通过在github上建立两个分支来解决上述问题，一个用于存储静态网页（master分支），另一个用于备份内容（hexo）分支。** 一旦博客网站出现问题，只需要克隆该项目，进入hexo分支继续写作，避免重复环境搭建，提高工作效率。
## 二.基本环境搭建流程 ##
1. 环境说明： Win7，git v2.14.2，node v8.9.3，hexo v3.4.4
2. [下载安装git](https://git-scm.com/downloads)
3. [下载安装Nodejs](https://nodejs.org/en/)
4. 在本地建立博客站点目录，/github/wxquare.github.io，进入该目录，以下操作都是在该目录下面进行
5. 执行下面命令后，构建本地博客。浏览器输入http://localhost:4000/验证是否正确。
``` 
   npm i -g hexo  安装hexo
   hexo init  初始化  
   hexo generate  生成静态网页
   hexo server 启动服务器 ```
6. 在Github上创建wxquare.github.io仓库，建立master和hexo两个分支，master用于存储静态网页，hexo备份本地博客的内容。
7. 执行下面的命令，初始化本地仓库，与Github上的wxquare.github.io仓库同步。
```
   git init 
   git remote add origin git@github.com:wxquare/wxquare.github.io.git 
   git pull origin master 
   git branch hexo 
   git checkout hexo 
   git pull origin hexo 
```
8. 修改站点配置文件_config.yml，将本地博客迁移到github上。
![](https://github.com/wxquare/wxquare.github.io/raw/hexo/source/photos/hexo_deploy.jpg)
8. 执行`npm install hexo-deployer-git --save`
9. 执行`hexo deploy`部署到github上，网络原因稍等一会儿，浏览器输入https://wxquare.github.io/ 查看效果
9. 将本地博客内容备份到hexo分支上去 
   `git checkout hexo;git add .;git commit -m "init";git push origin hexo`
10. 写博客的基本流程是用markdown写文章，保存到/source/_posts目录下,执行`hexo clean`;`hexo generate`;`hexo server`；查看本地的博客，然后部署到github上去`hexo deploy`，最后备份本地hexo博客`git add .`;`git commit -m "add Github双分支+hexo搭建个人博客"`；`git push orign hexo`

## 三.博客个性化设置 ##
1. 主题设置和修改。hexo初始化默认的主题是landscape，[https://hexo.io/themes/](https://hexo.io/themes/)提供了许多的主题，根据喜好为博客的主题，主题的文档提供了使用方法，设置相应的参数，调整为自己喜欢的格式。我这里选择的Next主题。
2. 进入博客站点目录，`git clone https://github.com/iissnan/hexo-theme-next themes/next`
3. 修改站点配置文件_config.yml，theme：next，重新编译，配置
4. 参考[http://theme-next.iissnan.com/](http://theme-next.iissnan.com/)，配置网站。 
5. 不断更新....

## 四.填坑  ##
1. 博客中的图片，将图片放在hexo分支的souce/photos目录下面，然后同步到github中，在hexo分支中找到图片，通过download按钮查看图片的地址。