---
title: 搭建基于VScode的Windows开发环境
---

# 一.下载安装 [vscode](https://code.visualstudio.com/)  
<center>[***https://code.visualstudio.com/***](https://code.visualstudio.com/)</center>

# 二.搭建Cpp开发调试环境
1. 下载安装MinGW-w64，[https://sourceforge.net/projects/mingw-w64/files/latest/download](https://sourceforge.net/projects/mingw-w64/files/latest/download)
2. 添加进入环境变量PATH。D:\coding\tools\MinGW-w64\mingw64\bin
3. 使用vscode打开一个C++工程的目录
4. 安装Microsoft C/C++ extension
5. 生成和配置c_cpp_properties.json，实现代码补全和导航的功能
6. 生成和配置task.json文件，编译代码
7. 生成和配置launch.json，调试代码
8. 阅读参考：[C/C++ for VS Code (Preview)](https://code.visualstudio.com/docs/languages/cpp)

# 三.搭建Golang开发调试环境
1. 下载安装Golang语言,[https://redirector.gvt1.com/edgedl/go/go1.9.2.windows-amd64.msi](https://redirector.gvt1.com/edgedl/go/go1.9.2.windows-amd64.msi)
2. 配置GOROOT, GOROOT=D:\coding\tools\golang\go1.9.2
3. 配置GOPATH, GOPATH=D:\coding\tools\golang\GOPATH;D:\coding\tools\golang\goWorkspace
4. 使用 `go env` 检查设置是否正确
5. 使用 code 打开 goWorkspace目录，安装go插件![](https://i.imgur.com/cF4nqqS.jpg)
6. 第一次编写Go代码的时候，需要安装所需的工具，例如gocode，godef等。设置好GOPATH的环境变量，打开一个Go代码，在右下角会看到“Analysis Tools Missing”的提示，点击它就会安装所需的工具。也可以手动安装
	<pre>
		go get -u -v github.com/nsf/gocode
	    go get -u -v github.com/rogpeppe/godef   
	    go get -u -v github.com/golang/lint/golint   
	    go get -u -v github.com/lukehoban/go-outline  
	    go get -u -v sourcegraph.com/sqs/goreturns 
	    go get -u -v golang.org/x/tools/cmd/gorename  
	    go get -u -v github.com/tpng/gopkgs  
	    go get -u -v github.com/newhook/go-symbols  
	    go get -u -v golang.org/x/tools/cmd/guru
	</pre>
7.  vscode的Go插件有一些定制化的配置，可以通过*File->Perferences->User Setting*进行设置
	<pre>
	{
	    "window.zoomLevel": 0,
	    "editor.wordWrap": "on",
	    "editor.minimap.renderCharacters": false,
	    "editor.minimap.enabled": false,
	    "terminal.external.osxExec": "iTerm.app",
	    //"go.useLanguageServer": true,
	    "go.docsTool": "gogetdoc",
	    "go.buildOnSave": true,
	    "go.lintOnSave": true,
	    "go.vetOnSave": true,
	    "go.buildTags": "",
	    "go.buildFlags": [],
	    "go.lintFlags": [],
	    "go.vetFlags": [],
	    "go.coverOnSave": false,
	    "go.useCodeSnippetsOnFunctionSuggest": false,
	    "go.formatOnSave": true,
	    "go.formatTool": "gofmt",
	    "go.goroot": "D:/coding/tools/golang/go1.9.2",
	    "go.gopath": "D:/coding/tools/golang/GOPATH",
	    "go.gocodeAutoBuild": false,
	    "git.ignoreMissingGitWarning": true,
	    "go.autocompleteUnimportedPackages": true
	}
	</pre>