---
title: 基于Github双分支和Hexo搭建博客
categories:
- other
---

## 1. 基本步骤

- 参考：https://zhuanlan.zhihu.com/p/26625249
- 安装git、node.js、hexo
- 在github建立repo，命名为wxquare.github.io，两个branch，分别为master和hexo
- 使用markdown在hexo branch 写文章，hexo generate生成静态文件，并通过hexo deploy 部署到远端
- 申请域名wxquare.top，绑定wxquare.github.io
- 
    
## 2. 执行下面命令，初始化本地博客
   ``` 
       npm i -g hexo  安装hexo
       hexo init  初始化  
       hexo generate  生成静态网页
       hexo server 启动服务器 （浏览器输入http://localhost:4000/验证是否正确。）
       hexo deploy 部署到远端 （wxquare.github.io）
   ```

## 3. 修改站点配置文件_config.yml，使得能将本地博客迁移到github上
    ```
    deploy:
      type: git
      repo: https://github.com/wxquare/wxquare.github.io.git
      branch: master
    ```
## 4. 正常写文章发布的流程
   - 在hexo branch /source/_posts 下使用markdown写文章
   - 使用hexo genergate 生成静态文件
   - hexo server 查看本地效果
   - hexo deploy 到远端
   - 提交修改文件到hexo

## 5. 博客配置进阶
1. 主题设置和修改。hexo初始化默认的主题是landscape，[https://hexo.io/themes/](https://hexo.io/themes/)提供了许多的主题，根据喜好为博客的主题，主题的文档提供了使用方法，设置相应的参数，调整为自己喜欢的格式。我这里选择的Next主题。
2. 进入博客站点目录，`git clone https://github.com/iissnan/hexo-theme-next themes/next`
3. 修改站点配置文件_config.yml，theme：next，重新编译，配置
4. 参考[http://theme-next.iissnan.com/](http://theme-next.iissnan.com/)，配置网站。 
5. 增加分类
6. hexo 增加支持markdown公式：http://stevenshi.me/2017/06/26/hexo-insert-formula/
7. 博客中的图片，将图片放在hexo分支的souce/photos目录下面，然后同步到github中，在hexo分支中找到图片，通过download按钮查看图片的地址。
8. Hexo博客Next主题添加统计文章阅读量功能：https://bjtu-hxs.github.io/2018/06/12/leancloud-config/


参考：GitHub+Hexo 搭建个人网站详细教程：https://zhuanlan.zhihu.com/p/26625249
