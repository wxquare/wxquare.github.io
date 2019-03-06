### 1、pkg-config：解决编译和链接时候出现的undefined symbol问题
pkg-config能方便使用第三方库和头文件和库文件，其运行原理 
 
- 它首先根据PKG_CONFIG_PATH环境变量下寻找库对应的pc文件  
- 然后从pc文件中获取该库对应的头文件和库文件的位置信息
  
例如在项目中需要使用opencv库，该库包含的头文件和库文件比较多  

- 首先查看是否有对应的opencv.pc find /usr -name opencv.pc  
- 查看该路径是否包含在PKG_CONFIG_PATH  
- 使用pkg-config --cflgs --libs opencv 查看库对应的头文件和库文件信息  
- pkg-config --modversion opencv 查看版本信息

参考链接：[https://blog.csdn.net/luotuo44/article/details/24836901](https://blog.csdn.net/luotuo44/article/details/24836901)

### 2、ldd：帮助解决运行时出现的问题
**现象**：  
- <font color=red >error while loading shared libraries: libopencv_cudabgsegm.so.3.4: cannot open shared object file: No such file or directory </font>  
- ldd ./xxx，发现库文件not found  

      libopencv_cudaobjdetect.so.3.4 => not found  
      libopencv_cudalegacy.so.3.4 => not found

**ld.so 动态共享库搜索顺序**：  
1. ELF可执行文件中动态段DT_RPATH指定；gcc加入链接参数“-Wl,-rpath”指定动态库搜索路径；  
2. 环境变量LD_LIBRARY_PATH指定路径；  
3. /etc/ld.so.cache中缓存的动态库路径。可以通过修改配置文件/etc/ld.so.conf 增删路径（修改后需要运行ldconfig命令）；  
4. 默认的 /lib/;  
5. 默认的 /usr/lib/  

**解决办法**：  
- 确认系统中是包含这个库文件的  
- pkg-config --libs opencv 查看opencv库的路径  
- export LD_LIBRARY_PATH=/usr/local/lib64，增加运行时加载路径  

 参考链接：[https://www.cnblogs.com/amyzhu/p/8871475.html](https://www.cnblogs.com/amyzhu/p/8871475.html)
