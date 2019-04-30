# Git global setup
git config --global user.name  "tersewang"
git config --global user.email "tersewang@tencent.com"

# Create a new repository
git clone http://git.code.oa.com/tersewang/terse-test.git
cd terse-test
touch README.md
git add README.md
git commit -m "add README"
git push -u origin master

# Existing folder or Git repository
cd existing_folder
git init
git remote add origin http://git.code.oa.com/tersewang/terse-test.git
git add .
git commit
git push -u origin master


# 查看git配置项目
git config --global --list
git config --list

# git 设置代理
git config --global https.proxy http:://web-proxy.tencent.com:8080
git config --global http.proxy http://dev-proxy.oa.com:8080

# 取消代理
git config --global --unset http.proxy 
git config --global --unset https.proxy

dev-proxy.oa.com

https://stackoverflow.com/questions/21085607/git-error-the-requested-url-returned-error-504-gateway-timeout-while-accessing


二、
git submodule