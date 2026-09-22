---
title: 边缘端视觉算法异构加速实战：OpenCL、OpenCV UMat 与 KCF 性能测试
date: 2020-08-13
categories:
  - AI 与 Agent
  - AI
---

## 引言：为什么要在移动端尝试 OpenCL

当时正在做 KCF（Kernelized Correlation Filters）目标跟踪算法的移动端优化。KCF 的计算量较大，而移动端设备的 CPU 计算资源和功耗预算有限，因此尝试把其中一部分适合并行化的操作迁移到 GPU 上执行。KCF 的算法背景可参考原始论文和预印本 [18][19]。

本文不是一份完整的 OpenCL 教程，也不是一组可以代表所有移动设备的通用 benchmark，而是基于 Android、OpenCV 和 OpenCL 的一次探索性测试。原始测试主要回答以下问题：

1. OpenCV 的 `UMat` 是否能把算子迁移到 GPU，以及 CPU/GPU 数据传输的成本是多少？
2. OpenCL 常用 API 的初始化、内存分配、数据传输和 Kernel 调度分别需要多长时间？
3. global work size 对简单内存拷贝 Kernel 的性能有什么影响？
4. 多个 Command Queue 是否一定能够带来并行收益？

需要特别说明的是，OpenCV 的 `UMat` 是 Transparent API 的抽象，并不保证每个算子都在 GPU 上执行；OpenCL 运行时可能根据算子、数据类型和设备能力回退到 CPU [7][8]。同样，编译 OpenCV 时打开 `WITH_OpenCL=ON`，也不等于设备上一定存在可用的 OpenCL 驱动。

## 一、测试环境与方法

原始实验使用的主要环境如下：

| 项目 | 原始配置或说明 |
| ------ | ------ |
| OpenCV | 3.4.6 [11] |
| Android NDK | android-ndk-r16b |
| ABI | arm64-v8a |
| C++ 标准库 | `gnustl_static` |
| OpenCL | 设备厂商提供的移动端实现 |
| 测试设备 | 三星 GALAXY On7、小米 6、小米 MIX 2S |
| 主要算子 | `Mat ↔ UMat`、`cvtColor`、Buffer 拷贝、NDRange Kernel、Command Queue |

这是一套历史环境。NDK r18 已移除 GNU libstdc++，因此 `gnustl_static` 不能作为当前 Android 工程的默认配置 [13]。如果要复现实验，应额外记录 Android 版本、OpenCL 驱动版本、OpenCL C 版本、设备频率、编译器版本、图像通道数和温度状态；仅记录手机型号不足以保证结果可复现。

对于移动端性能测试，应区分以下几类时间：

```text
初始化时间 = Context、Program、Kernel、Buffer 等对象的创建和编译
Kernel 时间 = 设备真正执行 Kernel 的时间
端到端时间 = 上传 + Kernel + 下载 + 同步
```

GPU 端到端耗时可以近似写成：

```text
T_gpu = T_init + T_upload + T_kernel + T_download + T_sync
```

如果只测 `T_kernel`，就不能直接把结果解释为完整算法的加速比。建议使用单调时钟测量主机端总耗时，并在开启 `CL_QUEUE_PROFILING_ENABLE` 后通过 Event profiling 获取设备端时间 [5]。Android 级别的调度、频率和热状态，则应结合 Perfetto 或 Simpleperf 观察 [14][15]。

## 二、OpenCL 基本执行模型

OpenCL 的基本执行路径可以概括为：

```text
Platform → Device → Context → Command Queue
                              ↓
                    Program → Kernel
                              ↓
                 Buffer / Image / Event
```

几个容易混淆的概念如下：

- **work-item**：执行 Kernel 的最小逻辑实例，可以通过 `get_global_id()` 获取自己的全局索引。
- **work-group**：一组可以协作执行的 work-item，大小受到设备和 Kernel 资源限制。
- **global work size**：本次 NDRange 中 work-item 的总数量。
- **local work size**：每个 work-group 中的 work-item 数量。
- **Command Queue**：主机向设备提交内存操作和 Kernel 的队列，默认通常是 in-order queue。
- **Event**：描述异步命令的完成状态，也可以用于建立命令之间的依赖关系。

`clEnqueueNDRangeKernel` 的 global work size 和 local work size 需要结合设备上限、Kernel 资源使用和内存访问模式选择，不能简单认为 work-item 越多越快 [1][3][4]。如果要研究 work-group 调优，应固定数据规模，同时独立比较不同的 global work size、local work size 和向量化方式。

### 2.1 内存地址空间与数据驻留

OpenCL Kernel 中常见的地址空间包括 global、constant、local 和 private。它们首先是编程模型中的访问范围，不应简单等同于某一种固定的物理存储：具体映射由设备和驱动决定，但不同地址空间的可见范围、生命周期和同步规则不同 [1][2]。

- **global memory**：所有 work-item 和多个 Kernel 实例都可以访问，容量通常较大，但访问延迟也更高。图像输入、输出和中间 Buffer 通常位于这一地址空间。
- **constant memory**：Kernel 只读的数据区域，适合存放不会随 work-item 改变的小型参数表或卷积系数。
- **local memory**：同一个 work-group 内共享的区域，适合把会被重复访问的数据暂存起来。它只在 work-group 内可见，不能用来跨 work-group 通信。
- **private memory**：每个 work-item 独有的变量。局部变量过多时，设备可能将其溢出到较慢的存储区域，因此寄存器和局部变量数量也会影响 occupancy。

对移动端图像算法而言，如果每次调用都在 `Mat` 与设备 Buffer 间往返，GPU 可能被传输和同步成本抵消；让预处理、特征计算和模型更新连续使用同一批 `UMat` 或 Buffer，才有机会摊薄这些成本。

同步也有层次差异。`barrier()` 只能协调同一个 work-group 中的 work-item，不能让不同 work-group 在 Kernel 内安全地交换数据；主机侧的 Event wait list 则用于安排不同命令之间的依赖。`clFlush` 负责推动已提交命令进入设备，`clFinish` 则等待队列中所有命令完成，二者都不应被随意混作性能计时点 [1][3][5]。因此，性能代码应尽量使用精确的 Event 依赖，避免用过多全局同步把本可重叠的阶段串行化。

## 三、编译带 OpenCL 的 OpenCV SDK

KCF 使用了不少 OpenCV 函数，因此最初的方案是编译一个启用 OpenCL 的 OpenCV SDK。原始构建命令如下，路径参数需要根据本地环境调整：

```bash
cmake \
  -DCMAKE_BUILD_WITH_INSTALL_RPATH=ON \
  -DCMAKE_TOOLCHAIN_FILE="/home/xxx/code/mobile/third_party/opencv-3.4.6/platforms/android/android.toolchain.cmake" \
  -DANDROID_NDK="/home/xxx/code/mobile/tools/android-ndk-r16b" \
  -DANDROID_SDK="/home/xxx/code/mobile/tools/android_sdk/tools" \
  -DANDROID_NATIVE_API_LEVEL=19 \
  -DANDROID_ABI="arm64-v8a" \
  -DANDROID_ARM_NEON=TRUE \
  -DANDROID_STL=gnustl_static \
  -DCMAKE_BUILD_TYPE=Release \
  -DOPENCV_EXTRA_MODULES_PATH="/home/xxx/code/mobile/third_party/opencv_contrib-3.4.6/modules" \
  -DCMAKE_INSTALL_PREFIX="/home/xxx/code/mobile/third_party/opencv-3.4.6/install_20190623_OpenCL" \
  -DBUILD_opencv_java=ON \
  -DBUILD_ANDROID_PROJECTS=OFF \
  -DBUILD_ANDROID_EXAMPLES=OFF \
  -DBUILD_DOCS=OFF \
  -DBUILD_PERF_TESTS=OFF \
  -DBUILD_TESTS=OFF \
  -DBUILD_FAT_JAVA_LIB=OFF \
  -DWITH_OpenCL=ON \
  -DWITH_CUDA=OFF \
  -DWITH_MATLAB=OFF \
  -DBUILD_opencv_aruco=OFF \
  -DBUILD_opencv_calib3d=OFF \
  -DBUILD_opencv_features2d=OFF \
  ..
```

当前 Android 项目更建议参考 NDK 的 CMake 集成文档 [12]，并使用仍受支持的 C++ 运行时配置。上面的命令保留在本文中，是为了说明原始实验环境，而不是推荐当前项目直接复制使用。

## 四、OpenCV UMat 与数据传输测试

### 4.1 运行时确认是否使用 OpenCL

在测试 `UMat` 之前，应先确认 OpenCL 运行时和设备状态：

```cpp
#include <opencv2/core/ocl.hpp>

bool prepareOpenCL() {
    if (!cv::ocl::haveOpenCL()) {
        return false;
    }

    cv::ocl::setUseOpenCL(true);
    if (!cv::ocl::useOpenCL()) {
        return false;
    }

    cv::ocl::Device device = cv::ocl::Device::getDefault();
    std::cout << "OpenCL device: " << device.name() << std::endl;
    std::cout << "OpenCL version: " << device.version() << std::endl;
    return device.available();
}
```

`UMat` 和 `cv::ocl` 的具体 API 见 OpenCV 文档 [7][8][9]。实际工程还应记录设备名称、驱动版本和算子是否发生 CPU fallback。否则，即使代码使用了 `UMat`，也不能仅凭 API 名称判断 GPU 一定参与了计算。

### 4.2 Mat 与 UMat 的拷贝

原始测试使用 `image.copyTo(u_img)` 和 `u_img.copyTo(out)` 测量 CPU 与设备内存之间的拷贝耗时。这个测试有参考价值，但需要区分以下情况：

1. 第一次调用可能包含运行时初始化、内存分配或 Kernel 准备成本。
2. 后续调用可能复用了已经分配的内存。
3. 移动端常见统一内存架构，所谓“CPU 到 GPU 拷贝”可能主要体现为映射、缓存同步或设备运行时管理，不应简单类比独立显卡的 PCIe 拷贝。
4. 需要记录图像的数据类型、通道数、步长和内存是否连续。

原始测试数据如下。表中的“首次”与“平均”必须分别理解为冷启动和稳态测量，不能直接混为同一个加速比。

| 手机型号 | CPU 型号 | GPU 型号 | OpenCL 版本 | 首次 Mat→UMat | 平均 Mat→UMat | 首次 UMat→Mat | 平均 UMat→Mat | 图像尺寸/文件大小 | Host→Device | Device→Host |
| ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ |
| 三星 GALAXY On7 | 骁龙 410 MSM8916 | Adreno 306 | 2 | 25.2 ms | 0.8 ms | 1.5 ms | 0.8 ms | 720×480，159 KB | 221 MB/s | 258 MB/s |
| 三星 GALAXY On7 | 骁龙 410 MSM8916 | Adreno 306 | 2 | 30.18 ms | 2.88 ms | 5.5 ms | 2.9 ms | 1920×1080，约 6 MB | 2.14 GB/s | 2.14 GB/s |
| 小米 6 | 骁龙 835 | Adreno 540 | 2 | 16.602 ms | 0.754 ms | 2.85 ms | 0.795 ms | 1920×1080，约 6 MB | 7.9 GB/s | 8.06 GB/s |
| 小米 6 | 骁龙 835 | Adreno 540 | 2 | 17.010 ms | 0.332 ms | 1 ms | 0.265 ms | 720×480，159 KB | 632 MB/s | 898.2 MB/s |
| 小米 MIX 2S | 骁龙 845 | Adreno 630 | 2 | 8.7 ms | 2.1 ms | 6.1 ms | 0.9 ms | 1920×1080，约 6 MB | 6.6 GB/s | 6.62 GB/s |
| 小米 MIX 2S | 骁龙 845 | Adreno 630 | 2 | 3.3 ms | 0.5 ms | 2.2 ms | 0.4 ms | 720×480，约 1.6 MB | 654 MB/s | 682 MB/s |

这些数据缺少采样次数、统计分布、温度和内存复用策略，因此适合用于观察趋势，不宜作为设备之间的严格排名。

### 4.3 `cvtColor` 测试

原始代码的 CPU 版本在计时前完成了图像读取，OpenCL 版本则在进入计时循环前完成了 `Mat → UMat` 拷贝，计时循环主要覆盖 `cvtColor`。因此，下面的 OpenCL 平均值更接近“已经完成数据准备后的算子稳态时间”，不是完整的端到端时间。

| 手机型号 | 执行路径 | 图像尺寸 | 首次运行 | 平均运行 |
| ------ | ------ | ------ | ------: | ------: |
| 三星 GALAXY On7 | CPU | 1920×1080 | 3.2 ms | 1.8 ms |
| 三星 GALAXY On7 | OpenCL | 1920×1080 | 273 ms | 0.6 ms |
| 三星 GALAXY On7 | CPU | 720×480 | 1.2 ms | 0.62 ms |
| 三星 GALAXY On7 | OpenCL | 720×480 | 274 ms | 0.25 ms |
| 小米 MIX 2S | CPU | 1920×1080 | 3 ms | 1.3 ms |
| 小米 MIX 2S | OpenCL | 1920×1080 | 154 ms | 0.36 ms |
| 小米 MIX 2S | CPU | 720×480 | 0.5 ms | 0.21 ms |
| 小米 MIX 2S | OpenCL | 720×480 | 80.5 ms | 0.09 ms |

从原始记录看，OpenCL 首次运行有明显的初始化或 Kernel 准备开销，稳态算子时间较低。但如果每一帧都需要上传和下载，最终是否有收益必须使用同一条端到端流水线重新测量。OpenCV 的 OpenCL 优化说明也强调，数据驻留和算子支持情况会影响实际收益 [10]。

### 4.4 让移动端基准测试可复现

移动端性能测试最容易出现的问题，不是不会调用计时函数，而是测试协议没有把影响因素固定下来。一次更可靠的测试至少应包含以下步骤：

1. **固定输入**：记录图像尺寸、通道数、数据类型、步长和是否连续；不要让不同测试隐含使用不同的图片压缩格式或不同的解码时间。
2. **区分冷启动和稳态**：第一次运行单独记录，用于观察 Context、Program、Kernel 和内存分配成本；随后执行若干次预热，再统计稳态运行时间。
3. **重复采样**：不要只报告一次或简单平均值。建议保存最小值、中位数、P95 和最大值，必要时同时保存每次样本，便于发现偶发的调度抖动。
4. **统一计时边界**：CPU 和 GPU 测试必须覆盖同样的工作范围。若 GPU 测试排除了上传和下载，就应明确称为“Kernel 稳态时间”，不能与 CPU 端到端时间直接比较。
5. **验证输出**：每次性能测试都应抽样比较 CPU 和 GPU 的输出，例如使用 `memcmp`、最大绝对误差或像素级误差统计。性能变快但结果错误，不是优化成功。
6. **控制设备状态**：记录电量、温度、前台应用、屏幕状态和测试持续时间。Android 设备在连续运行后可能降频，短时间的峰值结果不一定代表长期运行能力 [16][17]。

统计结果还应该带上测试协议，例如：`warmup=20`、`samples=100`、`metric=P50/P95`、`include_transfer=true`、`output_verified=true`，以免把 Kernel 时间和完整流水线时间混为一谈。

在工具选择上，主机端可以用单调时钟测量完整流程，OpenCL 设备端使用 Event profiling 拆分命令区间，Android 系统层则用 Perfetto 观察线程、频率和调度轨迹 [5][14]。Simpleperf 更适合回答 CPU 热点在哪里、线程是否被调度，以及 CPU 版本是否已经受到缓存或频率影响 [15]。三层数据结合起来，才能把“感觉 GPU 更快”转化为可解释的性能证据。

## 五、OpenCL 核心 API 性能测试

原始 API 测试结果如下。这里的数值是单次测试记录，文章没有保存完整的重复次数、平均值、P95 和热状态，因此不应把它们当作稳定基线。

| 手机型号 | CPU 型号 | GPU 型号 | OpenCL 版本 | API | 原始测试数据 |
| ------ | ------ | ------ | ------ | ------ | ------ |
| 小米 6 | 骁龙 835 | Adreno 540 | 2 | GPU 内存分配 `clCreateBuffer` | 1 MB：430 μs；5 MB：1000 μs；10 MB：2000 μs |
| 小米 6 | 骁龙 835 | Adreno 540 | 2 | CPU→GPU `clEnqueueWriteBuffer` | 1 MB：105 μs；5 MB：400 μs；10 MB：700 μs |
| 小米 6 | 骁龙 835 | Adreno 540 | 2 | GPU→CPU `clEnqueueReadBuffer` | 1 MB：60 μs；5 MB：400 μs；10 MB：600 μs |
| 小米 6 | 骁龙 835 | Adreno 540 | 2 | Kernel 编译 `clBuildProgram` | 69682 μs |
| 小米 6 | 骁龙 835 | Adreno 540 | 2 | 创建 Kernel `clCreateKernel` | 50 μs |
| 小米 6 | 骁龙 835 | Adreno 540 | 2 | 提交 `clEnqueueNDRangeKernel` | 首次约 5000 μs，之后约 800 μs |

需要区分“提交命令的主机开销”和“设备执行 Kernel 的时间”。`clEnqueueNDRangeKernel` 返回得快，不代表 Kernel 已经执行完成；只有等待 Event 或使用设备端 profiling，才能获得更可靠的执行时间 [3][5]。

## 六、global work size 与内存拷贝效率

### 6.1 CPU 循环拷贝与 `memcpy`

测试图像大小为：

```text
3840 × 2160 × 3 = 24883200 bytes
```

原始记录如下：

```cpp
char* out = new char[bmp_size];
for (int i = 0; i < bmp_size; ++i) {
    out[i] = bmp_data[i];
}
```

运行时间约为 13 ms。

```cpp
memcpy(out, bmp_data, bmp_size);
```

运行时间约为 3 ms。

CPU 循环与 `memcpy` 的比较还需要说明缓存状态、编译优化级别和输出是否被后续使用；重复读取同一块数据可能得到明显不同于冷缓存的结果。

### 6.2 修正后的 OpenCL 拷贝 Kernel

原始 Kernel 存在全角逗号、参数数量不一致、可能遗漏尾部数据和重复写入等问题。下面给出一个只表达“每个 work-item 处理一个字节”的简化示例；它用于说明边界和访问语义，不能直接替代针对具体 GPU 的最优实现。

```c
__kernel void copy_bytes(__global const uchar* input,
                         __global uchar* output,
                         const ulong size) {
    size_t gid = get_global_id(0);
    if (gid < size) {
        output[gid] = input[gid];
    }
}
```

对应的 Buffer 访问方向应与 Kernel 一致：

```cpp
cl::Buffer input_buffer(context, CL_MEM_READ_ONLY, bmp_size);
cl::Buffer output_buffer(context, CL_MEM_WRITE_ONLY, bmp_size);
```

主机端的 `global work size` 可以向上取整，Kernel 通过边界判断忽略多出的 work-item：

```cpp
size_t local_size = 256;  // 需要根据设备能力复测
size_t global_size = ((bmp_size + local_size - 1) / local_size) * local_size;

queue.enqueueNDRangeKernel(
    kernel,
    cl::NullRange,
    cl::NDRange(global_size),
    cl::NDRange(local_size),
    nullptr,
    &event);
event.wait();
```

`local_size` 不是一个对所有 GPU 都适用的固定常数，应结合 `CL_DEVICE_MAX_WORK_GROUP_SIZE`、Kernel 资源占用、内存对齐和设备厂商建议值测试。OpenCL 的 NDRange 和 work-group 约束见 [1][2][4]。

### 6.3 原始 work-item 测试结果

原始测试使用一张 3840×2160 的图像，在小米 MIX 2S 上观察不同 global work size 的结果：

| global work size | 运行时间 |
| ------: | ------: |
| 1 | 2972 ms |
| 2 | 1526 ms |
| 4 | 792 ms |
| 8 | 418 ms |
| 16 | 252 ms |
| 32 | 166 ms |
| 64 | 122 ms |
| 128 | 104 ms |
| 256 | 64 ms |
| 512 | 60 ms |
| 1024 | 92 ms |
| 2048 | 662 ms |
| 4096 | 237 ms |
| 10240 | 180 ms |
| 102400 | 171 ms |
| 248832 | 167 ms |
| 2488320 | 16 ms |
| 24883200 | 15 ms |

原始结果显示，在该设备和该实现上，不同 global work size 之间存在明显差异。但 `2488320` 和 `24883200` 的异常低耗时需要重新复测，不能仅凭一次主机端计时解释为“work-item 越多越好”。可能影响结果的因素包括：

- local work size 由驱动自动选择；
- Kernel 是否真正完成并被正确计时；
- 主机端毫秒级输出造成的精度损失；
- 内存缓存、设备频率和热状态变化；
- Kernel 是否存在未覆盖的数据范围或错误的边界处理。

因此，本文只能得出“global work size 需要实测调优”的结论，不能得出某个固定值适用于其他设备。建议使用 Event profiling、多次运行、中位数/P95 和正确性校验重新确认。

## 七、Command Queue 与并发测试

原始问题是：如果有 `n` 个任务，每个任务包括 CPU→GPU 拷贝、Kernel 执行和 GPU→CPU 拷贝，使用一个 Queue 和多个 Queue 是否存在差异？

答案取决于 Queue 的顺序属性、设备是否支持重叠执行、内存传输路径和驱动调度策略。多个 Queue 并不自动意味着多个任务会并行；如果设备或驱动最终把操作串行化，Queue 数量增加只会带来额外管理开销。Queue 属性和异步命令模型见 [1][6]。

### 7.1 正确的比较方式

单 Queue 和多 Queue 必须满足相同的条件：

1. 相同的任务数量、输入数据和 Buffer 大小。
2. 不把 Queue、Program、Kernel 和 Buffer 的创建时间混入稳态执行时间。
3. 两种方案都统计完整的 Upload、Kernel、Download 和最终等待时间。
4. 使用 Event profiling 查看不同 Queue 的时间区间是否真正重叠。
5. 分别比较 in-order Queue 和设备支持时的 out-of-order Queue。

一个更清晰的任务级测量结构如下：

```cpp
auto start = monotonic_time_ns();

queue.enqueueWriteBuffer(input_buffer, false, 0, bmp_size,
                         bmp_data, nullptr, &write_event);
queue.enqueueNDRangeKernel(kernel, cl::NullRange,
                           cl::NDRange(global_size),
                           cl::NDRange(local_size),
                           {write_event}, &kernel_event);
queue.enqueueReadBuffer(output_buffer, false, 0, bmp_size,
                        host_output, {kernel_event}, &read_event);

read_event.wait();
auto end = monotonic_time_ns();
```

多 Queue 测试则为每个任务建立自己的 Queue、Buffer 和 Event，但仍然以所有任务完成后的总时间作为端到端指标。不能只测 Queue 创建时间，也不能只测其中一个 Kernel 的等待时间。

### 7.2 对原始结论的重新表述

原始实验观察到多个 Queue 的总耗时略低于单 Queue，部分 work-item 设置下可能有约 15% 的差异，并观察到 GPU 利用率曲线形态不同。这可以作为“该设备驱动在特定任务布局下可能存在调度差异”的线索，但还不足以证明多个 Queue 普遍更快。

要把这个观察变成可靠结论，还需要补充 Queue 属性、任务总耗时、Event 时间线、重复次数、温度和 GPU 利用率采集方式。Android 设备长时间运行时还可能发生热降频，热状态会显著影响结果 [16][17]。

### 7.3 从单 Kernel 调优到设备侧流水线

简单的字节拷贝 Kernel 主要受内存带宽和调度开销影响，不一定能代表 KCF 中包含 FFT、逐元素运算或相关计算的 Kernel。优化时应先确认热点属于计算受限、内存受限还是同步受限，再选择方向。

如果相邻 work-item 访问连续地址，通常更容易形成高效的内存访问；但改成 `uchar4`、`float4` 等向量类型，还要满足数据对齐、通道布局和设备支持，并用输出校验和设备端时间验证。图像数据还要关注行步长，不能把 `width × height × channel` 机械地当成唯一布局。

local memory 也不是默认的性能加速器。只有当数据会被同一 work-group 重复使用，并且搬入 local memory 的成本能够被多次访问摊薄时，显式缓存才可能有收益。如果每个元素只读取一次，额外的搬运和 barrier 反而会增加成本。调优时应分别测量 global-only、local-cache 和不同 local work size，而不是依据经验直接加入 local memory。

对于多个连续阶段，更有价值的方向通常是双缓冲或多缓冲流水线：缓冲区 A 在执行 Kernel 时，缓冲区 B 可以准备下一批输入；当设备支持并且依赖关系正确时，上传、计算和下载才可能出现重叠。每一个阶段都应通过 Event wait list 表达依赖，不要用无差别的 `clFinish` 把所有队列重新串起来 [1][5][6]。最终比较的也不是“创建了几个 Queue”，而是固定吞吐目标下的总耗时、峰值内存、错误率和功耗。

## 八、对 KCF 移动端优化的启示

本文的 UMat、内存拷贝和简单 Kernel 测试，最终都应该服务于 KCF 的端到端优化。OpenCV 也提供了 KCF 跟踪器接口，可作为算法实现和 API 行为的补充参考 [20]。建议对 KCF 的每个阶段单独测量：

| KCF 阶段 | CPU 耗时 | OpenCL Kernel | 上传/下载 | 调用次数 | 是否值得迁移 |
| ------ | ------: | ------: | ------: | ------: | ------ |
| 图像预处理 |  |  |  |  |  |
| 特征提取 |  |  |  |  |  |
| FFT/相关计算 |  |  |  |  |  |
| 模型更新 |  |  |  |  |  |
| 单帧总耗时 |  |  |  |  |  |

一个算子适合迁移到 GPU，通常需要同时满足以下条件：

- 计算量或并行度足够大；
- 数据在 GPU 上可以连续驻留，减少来回拷贝；
- Kernel 启动和同步成本可以被摊薄；
- 算子的 OpenCL 实现经过正确性和性能验证；
- 长时间运行时不会因为温度或功耗导致收益消失。

如果 KCF 只把一个很小的算子单独搬到 GPU，`T_upload + T_download + T_sync` 很可能超过 Kernel 本身的收益。更合理的优化方向通常是把一组连续算子放在同一设备侧流水线中，再以单帧端到端延迟、吞吐量、功耗和稳定性共同评价。

### 8.1 KCF 算子迁移的决策流程

KCF 优化不应从“把所有 OpenCV 调用都改成 UMat”开始，而应从 CPU baseline 的热点分析开始。可以采用以下流程：

1. 先在 CPU 版本上记录一帧跟踪过程的阶段耗时和调用次数，确认真正占用时间的函数，而不是只优化最容易改写的函数。
2. 对候选算子检查 OpenCV T-API 是否有对应实现；如果没有，就评估手写 OpenCL Kernel 的维护成本和正确性风险。某些算子可能静默回退到 CPU，必须通过运行时状态和 profile 结果确认 [7][8][10]。
3. 估算数据边界。若候选算子前后都需要在 `Mat` 和 `UMat` 之间转换，先计算转换成本；若多个候选算子可以连续处理同一批设备数据，则优先迁移成一个阶段。
4. 为每个阶段建立 CPU、GPU Kernel 和端到端三个指标，同时加入输出误差、内存占用和温度变化。只有当端到端时间稳定下降，才把优化合并到主流程。
5. 保留运行时开关和 CPU fallback。设备没有 OpenCL、驱动异常、温度过高或输入尺寸过小时，应能够回到 CPU 路径，而不是让跟踪流程因为 GPU 初始化失败而不可用。

对于 KCF，频域计算和大规模逐元素运算通常比很小的控制逻辑更值得分析，但这只是候选方向，不是对某个实现的性能承诺。实际选择还取决于 ROI 尺寸、特征通道数、FFT 实现、调用频率和数据是否能长期驻留设备端 [18][19]。

### 8.2 上线前检查清单

在将 OpenCL 优化合入移动端算法前，可以用下面的清单做一次复核：

- 是否记录了 CPU、Kernel 和端到端三种时间？
- 是否分别报告了冷启动和稳态结果？
- 是否使用中位数和 P95，而不是只给一次平均值？
- 是否验证了所有输出数据的正确性？
- 是否确认了 `UMat` 没有意外 fallback 到 CPU？
- 是否减少了 `Mat ↔ UMat` 的往返转换？
- 是否在长时间运行和高温状态下复测？
- 是否保留了 CPU fallback、超时和错误码处理？
- 是否在多个 Android 设备上检查了结果方向，而不是只优化一台手机？

这份清单的核心是确保性能收益能在真实输入、温度和生命周期中稳定复现。

## 九、结论

这次实验得到的主要经验可以总结为：

1. OpenCL 的初始化、Program 编译、内存分配和 Kernel 首次执行可能有明显的一次性成本。
2. `UMat` 能够简化 OpenCV 的异构计算接入，但不保证每个操作都由 GPU 执行，需要检查运行时设备和 fallback 情况。
3. CPU/GPU 数据传输可能抵消单个 Kernel 的性能收益，必须同时报告 Kernel 时间和端到端时间。
4. global work size、local work size、内存访问方式和设备驱动共同决定 Kernel 性能，不存在跨设备通用的最优 work-item 数量。
5. 多 Command Queue 是否有效取决于设备、驱动、Queue 属性和任务依赖，必须通过 Event 时间线验证，而不能只观察 Queue 数量。
6. 对 KCF 这类移动端算法，最终指标应是完整算法的单帧延迟、吞吐、功耗、温度和长期稳定性，而不是某个孤立算子的最佳耗时。

## 参考资料

### OpenCL

1. Khronos Group，[The OpenCL 3.0 Unified Specification](https://registry.khronos.org/OpenCL/specs/3.0-unified/html/)。
2. Khronos Group，[The OpenCL C 3.0 Language Specification](https://registry.khronos.org/OpenCL/specs/3.0-unified/html/OpenCL_C.html)。
3. Khronos Group，[OpenCL SDK Reference Pages](https://registry.khronos.org/OpenCL/sdk/3.0/docs/man/html/)。
4. Khronos Group，[clEnqueueNDRangeKernel Reference](https://registry.khronos.org/OpenCL/sdk/3.0/docs/man/html/clEnqueueNDRangeKernel.html)。
5. Khronos Group，[clGetEventProfilingInfo Reference](https://registry.khronos.org/OpenCL/sdk/3.0/docs/man/html/clGetEventProfilingInfo.html)。
6. Khronos Group，[clCreateCommandQueueWithProperties Reference](https://registry.khronos.org/OpenCL/sdk/3.0/docs/man/html/clCreateCommandQueueWithProperties.html)。

### OpenCV

7. OpenCV，[cv::UMat Class Reference](https://docs.opencv.org/4.x/d7/d1e/classcv_1_1UMat.html)。
8. OpenCV，[OpenCL Support in OpenCV Core](https://docs.opencv.org/4.x/dc/d83/group__core__opencl.html)。
9. OpenCV，[cv::ocl::Device Class Reference](https://docs.opencv.org/4.x/d8/d45/classcv_1_1ocl_1_1Device.html)。
10. OpenCV，[OpenCL Optimizations](https://github.com/opencv/opencv/wiki/OpenCL-optimizations)。
11. OpenCV，[OpenCV 3.4.6 Source Tree](https://github.com/opencv/opencv/tree/3.4.6)。

### Android 性能与构建

12. Android Developers，[Use CMake with the NDK](https://developer.android.com/ndk/guides/cmake)。
13. Android NDK，[Changelog r18](https://github.com/android/ndk/wiki/Changelog-r18)。
14. Perfetto，[Documentation](https://perfetto.dev/docs/)。
15. Android Developers，[Simpleperf](https://developer.android.com/ndk/guides/simpleperf)。
16. Android Developers，[Thermal Management](https://developer.android.com/games/optimize/adpf/thermal)。
17. Android Developers，[Performance](https://developer.android.com/topic/performance)。

### KCF 与目标跟踪

18. João F. Henriques, Rui Caseiro, Pedro Martins, Jorge Batista, [High-Speed Tracking with Kernelized Correlation Filters](https://ieeexplore.ieee.org/document/6870486), IEEE Transactions on Pattern Analysis and Machine Intelligence, 2015。
19. João F. Henriques et al., [High-Speed Tracking with Kernelized Correlation Filters](https://arxiv.org/abs/1404.7584), arXiv:1404.7584。
20. OpenCV，[Tracking API Reference](https://docs.opencv.org/4.x/dc/d6b/group__tracking.html)。
