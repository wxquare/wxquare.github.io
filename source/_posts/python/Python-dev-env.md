---
title: Python 开发环境
categories:
- Python
---

### 1. 保证python2和python3的兼容性
- Linux系统默认带有Python2，且很多命令依赖于Python2，不要卸载Python2，否则会造成很多问题
- 很多程序都会使用Python3，尤其是算法程序，因此需要在一个系统同时兼容Python不同版本
- Python3推荐使用Anaconda，因为它集成了很多算法包，例如numpy、sklean，pandas和Matplotlib等
- Anaconda 使用Conda管理和安装算法包，非常方便。
- <font color=red>在用户环境下默认使用python3，在root用户下默认使用python2，并且使用python2和python3指向各自的解释器。</font>

### 2. sublime3 和 anaconda插件
- 可直接通过package control安装anaconda插件
- 也可直接将github的anaconda 克隆到sublime安装包目录下面
https://github.com/DamnWidget/anaconda.git


### 3. pylint 和 pylinter
- python 安装pylint
- sublime 安装pylinter插件
- 配置信息
```
{
    // When versbose is 'true', various messages will be written to the console.
    // values: true or false
    "verbose": false,
    // The full path to the Python executable you want to
    // run Pylint with or simply use 'python'.
    "python_bin": "python",
    // The following paths will be added Pylint's Python path
    "python_path": [
        "/home/terse/anaconda3/bin/python3"
                   ],
    // Optionally set the working directory
    "working_dir": null,
    // Full path to the lint.py module in the pylint package
    //"pylint_path": "/home/terse/anaconda3/bin/pylint",
    // Optional full path to a Pylint configuration file
    "pylint_rc": null,
    // Set to true to automtically run Pylint on save
    "run_on_save": false,
    // Set to true to use graphical error icons
    "use_icons": false,
    "disable_outline": false,
    // Status messages stay as long as cursor is on an error line
    "message_stay": false,
    // Ignore Pylint error types. Possible values:
    // "R" : Refactor for a "good practice" metric violation
    // "C" : Convention for coding standard violation
    // "W" : Warning for stylistic problems, or minor programming issues
    // "E" : Error for important programming issues (i.e. most probably bug)
    // "F" : Fatal for errors which prevented further processing
    "ignore": [],
    // a list of strings of individual errors to disable, ex: ["C0301"]
    "disable": [],
    "plugins": []
}
```
