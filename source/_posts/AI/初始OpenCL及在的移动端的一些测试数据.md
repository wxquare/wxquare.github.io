---
title: 初识OpenCL及在移动端的一些测试数据
categories:
- AI
---

　　最近在做kcf算法在移动端优化的相关工作，由于kcf算法计算量太大，而移动端计算性能有限，因此打算将kcf部分耗时操作通过GPU计算进行提升算法的性能。由于接触GPU和OpenCL的时间比较短，原理性的东西理解得也不深刻，本文主要在移动端测试了一些GPU和OpenCL的数据，无法分析内在原因，方便后续移动端算法优化。主要工作如下：
1. 编译了OpenCL的opencv版本sdk，测试了mat到umat相互内存拷贝和cvtcolor函数的性能。
2. 测试了OpenCL核心API的性能
3. 以内存拷贝核函数为例，测试OpenCL work_item数量与效率的关系。
4. 测试OpenCL多commandqueue的性能


## 一、opencv+OpenCL
### 1.1 编译opencv+OpenCL的sdk
　　KCF算法总使用了不少的opencv函数，开始想的是编译一个包含OpenCL的opencv的sdk，然后通过调用该sdk从而实现使用GPU加速算法的目的。编译opencv+OpenCL的sdk当时踩了不少坑，多番尝试之后，使用下面的命令是可以成功编译。分别下载opencv-3.4.6、android-ndk-r16b、opencv_contrib-3.4.6,在opencv中建build目录，运行下面命令，命令中使用一些路径相关的参数要根据环境适当修改。

```
cmake -DCMAKE_BUILD_WITH_INSTALL_RPATH=ON -DCMAKE_TOOLCHAIN_FILE="/home/xxx/code/mobile/third_party/ opencv-3.4.6/platforms/android/android.toolchain.cmake" -DANDROID_NDK="/home/xxx/code/mobile/tools/android-ndk-r16b" -DANDROID_SDK="/home/xxx/code/mobile/tools/android_sdk/tools" -DANDROID_NATIVE_API_LEVEL=19 -DANDROID_ABI="arm64-v8a" -DANDROID_ARM_NEON=TRUE -DANDROID_STL=gnustl_static -DCMAKE_BUILD_TYPE=Release -DOPENCV_EXTRA_MODULES_PATH="/home/xxx/code/mobile/third_party/opencv_contrib-3.4.6/modules" -DCMAKE_INSTALL_PREFIX="/home/xxx/code/mobile/third_party/opencv-3.4.6/install_20190623_OpenCL" -DBUILD_opencv_java=ON -DBUILD_ANDROID_PROJECTS=OFF -DBUILD_ANDROID_EXAMPLES=OFF -DBUILD_DOCS=OFF -DBUILD_PERF_TESTS=OFF -DBUILD_TESTS=OFF -DBUILD_FAT_JAVA_LIB=OFF -DWITH_OpenCL=ON -DWITH_CUDA=OFF -DWITH_MATLAB=OFF -DBUILD_opencv_aruco=OFF -DBUILD_opencv_calib3d=OFF -DBUILD_opencv_features2d=OFF .. 
```
### 1.2 测试mat到umat的相互转换的性能
　　在编译好opencv sdk之后，首先简单测试了一下sdk是否使用到了GPU资源。测试图片从CPU拷贝到GPU的的性能，opencv提供两组API。UMat::copyTo(OutputArray dst)和Mat::getMat(int access_flags)，实际测试中发现copyto那组性能比get的性能更好些，mat.getUmat函数会报错，还不知道什么原因。  
```
	void testMatCopyToUmat(const char* img, int times) {
	    cv::Mat image = cv::imread(img, cv::IMREAD_UNCHANGED);
	    cv::Mat out;
	    cv::UMat u_img;
	    if (u_img.empty()){
	        //
	    }
	    struct timeval start, end;
	    struct timeval last_time;
	    gettimeofday(&start, NULL);
	    last_time = start;
	    for (int i = 0; i < times; i++) {
	        image.copyTo(u_img);
	        //cv::cvtColor(image, out, cv::COLOR_BGR2GRAY);
	        gettimeofday(&end, NULL);
	        P("mat.copyToUmat:%d,run times:%d, spend:%d us", i,times, (end.tv_sec - last_time.tv_sec) * 1000000 + 
	                                       (end.tv_usec - last_time.tv_usec));
	        last_time = end;
	    }
	    gettimeofday(&end, NULL);
	    P("mat.copyToUmat: run times:%d, spend:%d ms", times, (end.tv_sec - start.tv_sec) * 1000 + 
	                                       (end.tv_usec - start.tv_usec)/1000);
	}
	void testUMatCopyToMat(const char* img, int times) {
	    cv::Mat image = cv::imread(img, cv::IMREAD_UNCHANGED);
	    cv::Mat out;
	    struct timeval start, end,last_time;
	    cv::UMat u_img;
	    image.copyTo(u_img);

	    gettimeofday(&start, NULL);
	    last_time = start;
	    for (int i = 0; i < times; i++) {
	        u_img.copyTo(out);
	        gettimeofday(&end, NULL);
	        P("mat.copyToUmat:%d,run times:%d, spend:%d us", i,times, (end.tv_sec - last_time.tv_sec) * 1000000 + 
	                                       (end.tv_usec - last_time.tv_usec));
	        last_time = end;
	    }
	    gettimeofday(&end, NULL);
	    P("mat.copyToUmat: run times:%d, spend:%d ms", times, (end.tv_sec - start.tv_sec) * 1000 + 
	                                       (end.tv_usec - start.tv_usec)/1000);
	}
```


| 手机型号 | CPU型号 | GPU型号 | OpenCL版本 | 首次mat拷贝umat | mat拷贝umat | 首次umat拷贝mat | umat拷贝mat | 图片格式 | 上行带宽 | 下行带宽 | 
| ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ |
|三星GALAXY On7|高通 骁龙410 MSM8916|	Adreno306|	2|	25.2ms|	0.8ms|	1.5ms|	0.8ms|	720*480 159KB|	运行1000次，221M/s|	运行1000次，258M/s|
|三星GALAXY On7|高通 骁龙410 MSM8916|	Adreno306|	2|	30.18ms|	2.88ms|	5.5ms|	2.9ms|	1920*1080 6MB|	运行1000次，2.14G/s|	运行1000次，2.14G/s|
|小米6 MI6|	骁龙 835|	高通 Adreno540|	2|16.602ms|	0.754ms|	2.85ms|	0.795ms|	1920*1080 6MB|	运行1000次，7.9G/s|	运行1000次，8.06G/s|
|小米6 MI6|	骁龙 835|	高通 Adreno540|	2|17.010ms|	0.332ms|	1ms|0.265ms|	720*480 159KB|	运行1000次，632M/S|	运行1000次，898.2M/s|	
|小米mix2s|	骁龙 845|	高通 Adreno630|	2|8.7ms|	2.1ms|	6.1ms|0.9ms|	1920*1080 6MB|	运行1000次，6.6G/S|	运行1000次，6.62G/s|	
|小米mix2s|	骁龙 845|	高通 Adreno630|	2|3.3ms|	0.5ms|	2.2ms|0.4ms|	720*480 1579KB|	运行1000次，654M/S|	运行1000次，682M/s|																

### 1.3 测试OpenCL cvtcolor函数性能
　　在测试完CPU和GPU内存拷贝的性能之外，之后测试了cvtcolor函数的性能，由于动态加载，OpenCL函数首次加载时特别耗时，大概需要200ms。除此之外，在不同规格的图片上，OpenCL的计算性能大概是cpu的2到3倍。
```
void cpu(const char* img, int times) {
    cv::Mat image; 
    cv::Mat out;
    struct timeval start, end,last;
    for (int i = 0; i < times; i++) {
        image = cv::imread(img, cv::IMREAD_UNCHANGED);
        gettimeofday(&start, NULL);
        cv::cvtColor(image, out, cv::COLOR_BGR2GRAY);
        gettimeofday(&end, NULL);
        P("run times:%d, spend:%d us", i, (end.tv_sec - start.tv_sec) * 1000000 +
                                       (end.tv_usec - start.tv_usec));
    }
}
 
void OpenCL(const char* img, int times) {
    cv::UMat u_img;
    cv::Mat image; 
    cv::UMat out;
    cv::Mat out1;

    std::vector<cv::UMat> v;
    for(int i=0;i<times;i++){
      image = cv::imread(img, cv::IMREAD_UNCHANGED);
      cv::UMat u_img;
      image.copyTo(u_img);
      v.push_back(u_img);
    }
    struct timeval start, end,last;
    for (int i = 0; i < times; i++) {
        gettimeofday(&start, NULL);
        cv::cvtColor(v[i], out, cv::COLOR_BGR2GRAY);
        gettimeofday(&end, NULL);
        P("run times:%d, spend:%d us", i, (end.tv_sec - start.tv_sec) * 1000000 +
                                       (end.tv_usec - start.tv_usec));
    }
}
```
测试数据：

| 手机型号 | cpu/gpu | 图片格式 | 首次运行时间 | 平均时间 |
| ------ | ------ | ------ | ------| ------ |
| 三星GALAXY On7|	cpu | 1920x1080 |	3.2ms|	1.8ms|																							
| 三星GALAXY On7|	OpenCL	| 1920x1080 |	273ms|	0.6ms|
| 三星GALAXY On7|	cpu	| 720x480 |	1.2ms|	0.62ms|																							
| 三星GALAXY On7|	OpenCL	| 720x480 |	274ms|	0.25ms|																							
| 小米mix2s |	cpu	| 1920x1080 |	3ms|	1.3ms|																							
| 小米mix2s |	OpenCL	| 1920x1080 |	154ms|	0.36ms|																							
| 小米mix2s |	cpu	| 720x480 |	0.5ms|	0.21ms|																							
| 小米mix2s |	OpenCL	| 720x480 |	80.5ms|	0.09ms|																							



## 二、OpenCL核心API性能测试
| 手机型号 | cpux型号 | GPU型号 | OpenCL版本 | API | 测试数据 |
| ------ | ------ | ------ | ------ | ------ | ------ |
|小米6 MI6 | 骁龙 835 | 高通 Adreno540 |	2 | gpu内存分配(clCreateBuffer) | 1M 430us,5M 1000us,10M 2000us |																			
|小米6 MI6 | 骁龙 835 | 高通 Adreno540 |	2 | cpu到gpu内存拷贝(writeBuffer) |	1M 105us,5M 400us,10M 700us |																			
|小米6 MI6 | 骁龙 835 | 高通 Adreno540 |	2 | gpu到cpu内存拷贝(ReadBuffer)	 | 1M 60us,5M 400us,10M 600us |																			
|小米6 MI6 | 骁龙 835 | 高通 Adreno540 |	2 | 核函数编译clBuildProgram	| 69682 us |																						
|小米6 MI6 | 骁龙 835 | 高通 Adreno540 |	2 | 创建核对象clCreateKernel	 | 50us	|																					
|小米6 MI6 | 骁龙 835 | 高通 Adreno540 |	2 | 核函数clEnqueueNDRangeKernel | 首次运行5000us，之后大概800us |																						



## 三、 测试OpenCL work_item数量与效率的关系
　　在OpenCL编程中，work_item和work_group的设置对程序的性能有较大的影响。这里以内存拷贝为例测试OpenCL中work_item数量与效率的关系。通过一张3840x2160的图片拷贝，分别测试了CPU和GPU内存拷贝的性能，测试了在不同work_item条件下GPU内存拷贝性能的性能。从测试结果来看，不同work_item对opencl的性能有较大的影响。测试结果显示，最开始时work_item数量曾倍数关系，之后会在100ms抖动，最好的情况是work_item数量与bmpsize大小相同。测试机器为小米mix2s。

### 3.1循环拷贝
bmpsize = 3840x2160x3,运行时间13ms
```
 char* out = new char[bmp_size];
  for(int i=0;i<bmp_size;i++){
    // P("%d",i);
    out[i] = bmp_data[i];
  }
```

### 3.2memcpy拷贝
bmpsize = 3840x2160x3,运行时间3ms
```
	memcpy(out,bmp_data,bmp_size);
```

### 3.3 opencl拷贝
核函数：
```
__kernel void convert_image(__global const uchar* in, 
            __global uchar* out，const int channel,
            const int width, const int height){
    int thread_count = get_global_size(0);
    int size = width * height * channel;
    int each_thread = size / thread_count;
    int tid = get_global_id(0);
    ; out[tid] = in[tid];
    for(int i=tid*each_thread;i<(tid+1)*each_thread;i++){
        out[i] = in[i];
    }
}
```
关键代码与work_item的设置：
```
    P("thread_count=%d", thread_count);
    gettimeofday(&start,NULL);
    err = queue.enqueueNDRangeKernel(kernel, cl::NullRange, cl::NDRange(thread_count, 1),
                               cl::NullRange, NULL, &event);
    event.wait();
    gettimeofday(&end,NULL);
    P("opecl wait:%d ms", (end.tv_sec - start.tv_sec) * 1000 + (end.tv_usec - start.tv_usec)/1000);
```
| work_item数量 | 运行时间 |
| ------ | ------ |
| 1 | 2972ms | 
| 2 | 1526ms |
| 4 | 792ms |
| 8 | 418ms |
| 16 | 252ms |
| 32 | 166ms |
| 64 | 122ms |
| 128 | 104ms |
| 256 | 64ms |
| 512 | 60ms |
| 1024 | 92ms |
| 2048 | 662ms |
| 4096 | 237ms |
| 10240 | 180ms |
| 102400 | 171ms |
| 248832 | 167ms |
| 2488320 | 16ms |
| 24883200 | 15ms |

**疑问:当work_item为256或者512是个较好的值，但是不明白为什么2488320和24883200值效果会更好。**

## 四、多commandqueue性能测试
　　在学习和测试OpenCL的过程中，有一个疑问能否使用多个commandqueue来做任务的并行。假设有n个任务，每个任务包含CPU到GPU内存拷贝，核函数执行，和GPU到CPU的内存拷贝。分别测试了使用一个commandqueue和n个commandqueue的性能，测试结果显示多个commandqueue会比使用单个commandqueue性能略好一些，但是差别不大。除此之外，与work_item的设置有关，多个commandqueue可能比单个commandqueue性能性能提升15%。从GPU利用率来说，单个commandqueuGPU曲线呈锯齿形状，而多个commandque呈梯形。部分代码如下：
单个commandqueue：
```
void test(const char* cl_file, const char* name, 
      const char* bmp_data, const int bmp_size, 
      const int width, const int height, const int channels,
      const int line_size, const int thread_count,
      const int run_times) 
{
  cl::Platform platforms = cl::Platform::getDefault();
  //P("platform count:%d", platforms.size());
  cl::Context context(CL_DEVICE_TYPE_GPU, NULL);
  std::vector<cl::Device> devices = context.getInfo<CL_CONTEXT_DEVICES>();
  P("Device count:%d", devices.size());
  std::ifstream in(cl_file, std::ios::in);
  std::stringstream buffer;
  buffer << in.rdbuf();
  cl_int err = CL_SUCCESS;
  cl::Program program_ = cl::Program(context, buffer.str());
  err = program_.build(devices);
  if (err != CL_SUCCESS) {
    P("build error");
    return;
  }  
  cl::Kernel kernel(program_, name, &err);
  if (err != CL_SUCCESS) {
    P("build error");
    return;
  }

  cl::CommandQueue queue(context, devices[0], 0, &err);
  if (err != CL_SUCCESS) {
    P("CommandQueue create error");
    return;
  }

  struct timeval start, end;
  cl::Event event;
  err = CL_SUCCESS;
  for(int i = 0;i<run_times;i++){
  {
   
    //see: https://github.khronos.org/OpenCL-CLHPP/classcl_1_1_buffer.html
    cl::Buffer in_buf(context, CL_MEM_WRITE_ONLY, bmp_size);
    cl::Buffer out_buf(context, CL_MEM_READ_ONLY, bmp_size);
    err = queue.enqueueWriteBuffer(in_buf, true, 0, bmp_size, bmp_data, NULL, &event);

    kernel.setArg(0, in_buf);
    kernel.setArg(1, out_buf);
    kernel.setArg(2, line_size);
    kernel.setArg(3, channels);
    kernel.setArg(4, width);
    kernel.setArg(5, height);

    P("thread_count=%d", thread_count);
    gettimeofday(&start,NULL);
    err = queue.enqueueNDRangeKernel(kernel, cl::NullRange, cl::NDRange(thread_count, 1),
                               cl::NullRange, NULL, &event);
    event.wait();
    gettimeofday(&end,NULL);
    P("opecl wait:%d ms", (end.tv_sec - start.tv_sec) * 1000 + 
                                         (end.tv_usec - start.tv_usec)/1000);

  
    char* h_out_buf = new char[bmp_size];
    err = queue.enqueueReadBuffer(out_buf, true, 0, bmp_size, h_out_buf, NULL, &event);
    if(0!=memcmp(h_out_buf, bmp_data, bmp_size)){
      P("data not same");
      return;
    }else{
      P("data same");
    }
  }
}
}
```


多个commandqueue：
```
void test_mutil_command_queue(const char* cl_file, const char* name, 
      const char* bmp_data, int bmp_size, 
      const int width, const int height, const int channels,
      const int line_size, const int thread_count,
      const int run_times) 
{
  cl::Platform platforms = cl::Platform::getDefault();
  //P("platform count:%d", platforms.size());
  cl::Context context(CL_DEVICE_TYPE_GPU, NULL);
  std::vector<cl::Device> devices = context.getInfo<CL_CONTEXT_DEVICES>();
  P("Device count:%d", devices.size());
  // cl::CommandQueue queue(context, devices[0], 0);
  //
  std::ifstream in(cl_file, std::ios::in);
  std::stringstream buffer;
  buffer << in.rdbuf();
  //
  //cl::Program::Sources source{
  //    std::make_pair(buffer.str().c_str(), buffer.str().size()) };
  cl_int err = CL_SUCCESS;
  cl::Program program_ = cl::Program(context, buffer.str());
  err = program_.build(devices);
  if (err != CL_SUCCESS) {
    P("build error %d",err);
    return;
  }  

  struct timeval start, end,end2;
  cl::Event event;
  err = CL_SUCCESS;
  gettimeofday(&start, NULL);
  std::vector<cl::CommandQueue> vQueue;
  std::vector<cl::Event> vEvents;
  std::vector<cl::Buffer> vInBuffers;
  std::vector<cl::Buffer> vOutBuffers;
  std::vector<char*> vHostOutBuf;
  std::vector<cl::Kernel> vKernels;
  std::vector<char*> vBmpdatas;

  for(int i=0;i<run_times;i++){
    cl::Event event;
    cl::CommandQueue queue(context, devices[0], 0, &err);
      if (err != CL_SUCCESS) {
      P("CommandQueue create error");
      return;
    }
    vQueue.push_back(queue);
    vEvents.push_back(event);
    cl::Buffer in_buf(context, CL_MEM_WRITE_ONLY, bmp_size);
    cl::Buffer out_buf(context, CL_MEM_READ_ONLY, bmp_size);
    vInBuffers.push_back(in_buf);
    vOutBuffers.push_back(out_buf);
    char* h_out_buf = new char[bmp_size];
    vHostOutBuf.push_back(h_out_buf);

    cl::Kernel kernel(program_, name, &err);
    if (err != CL_SUCCESS) {
      P("build error");
      return;
    }
    kernel.setArg(0, vInBuffers[i]);
    kernel.setArg(1, vOutBuffers[i]);
    kernel.setArg(2, line_size);
    kernel.setArg(3, channels);
    kernel.setArg(4, width);
    kernel.setArg(5, height);
    vKernels.push_back(kernel);

  }
  gettimeofday(&end, NULL);
  P("opecl create queue: spend:%d ms", (end.tv_sec - start.tv_sec) * 1000 + 
                                       (end.tv_usec - start.tv_usec)/1000);

  for(int i=0;i<run_times;i++){
    err = vQueue[i].enqueueWriteBuffer(vInBuffers[i], false, 0, bmp_size, bmp_data, NULL, &vEvents[i]);
  }
  for(int i=0;i<run_times;i++){
    vEvents[i].wait();
  }

  for(int i=0;i<run_times;i++){
    err = vQueue[i].enqueueNDRangeKernel(vKernels[i], cl::NullRange, cl::NDRange(thread_count, 1),
                                 cl::NullRange, NULL, &vEvents[i]);
  }

  for (int i = 0; i < run_times; ++i){ 
    vEvents[i].wait();
  }

  for(int i=0;i<run_times;i++){
    err = vQueue[i].enqueueReadBuffer(vOutBuffers[i], false, 0, bmp_size, vHostOutBuf[i], NULL, &vEvents[i]);
  }
  for(int i=0;i<run_times;i++){
    vEvents[i].wait();
  }
  
  for(int i=0;i<run_times;i++){
    if (0!=memcmp(vHostOutBuf[i], bmp_data, bmp_size)){
      P("data not same");
    }else{
      P("data same");
    }
  }
}
```