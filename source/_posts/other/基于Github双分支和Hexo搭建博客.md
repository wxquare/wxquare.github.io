---
title: 基于Github双分支和Hexo搭建博客
date: 2023-08-13
categories:
- other
---

## 基本步骤
- hexo 文档：https://hexo.io/zh-cn/docs/
- 安装git、node、npm、hexo、next主题
- 在github建立wxquare.github.io仓库，两个branch，分别为master和hexo
- 使用markdown在hexo branch 写文章，hexo generate生成静态文件，并通过hexo deploy 部署到远端
- 申请域名wxquare.top，绑定wxquare.github.io
- https://wxquare.github.io/

## 写文章发布blog的流程
  - 在hexo branch /source/_posts 下使用markdown写文章
  - 使用hexo genergate 生成静态文件
  - hexo server 查看本地效果
  - hexo deploy 到远端
  - 提交修改文件到hexo

   ``` 
       npm i -g hexo  安装hexo
       hexo init  初始化  
       hexo generate  生成静态网页
       hexo server 启动服务器 （浏览器输入http://localhost:4000/验证是否正确。）
       hexo deploy 部署到远端 （wxquare.github.io）
   ```

## hexo 配置
1. 修改站点配置文件_config.yml，使得能将本地博客部署到github上
    ```
    deploy:
      type: git
      repo: https://github.com/wxquare/wxquare.github.io.git
      branch: master
    ```

## next主题 配置
1. 主题设置和修改。hexo初始化默认的主题是landscape，[https://hexo.io/themes/](https://hexo.io/themes/)提供了许多的主题，根据喜好为博客的主题，主题的文档提供了使用方法，设置相应的参数，调整为自己喜欢的格式。我这里选择的Next主题
2. 安装next主题：https://theme-next.js.org/docs/getting-started/
3. 主题设置：https://theme-next.js.org/docs/theme-settings/


## 其它
1. 增加分类
2. hexo 增加支持markdown公式：http://stevenshi.me/2017/06/26/hexo-insert-formula/
3. 博客中的图片，将图片放在hexo分支的source/images目录下面，markdown和blog中均可以看到
4. Hexo博客Next主题添加统计文章阅读量功能：https://bjtu-hxs.github.io/2018/06/12/leancloud-config/
