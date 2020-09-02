---
title: 初识TVM，相比于tensorflow的2倍性能提升
categories:
- AI
mathjax: true
---

　　
　　最近在做深度学习模型加速的工作，先后尝试了模型权重量化(quantization)、模型权重稀疏（sparsification）和模型通道剪枝(channel pruning)等压缩方法，但效果都不明显。权重量化和稀疏属于非结构化的压缩，需要推理引擎和硬件的优化才能实现推理加速，通道剪枝能直接减少FLOPs，确实能卷积网络的效率，在ResNet56网络中能大概提升卷积50%的速度。在工程实践中，除了通过模型压缩提升推理性能，还可以通过优化推理引擎提高推理效率。目前存在多种开源的推理引擎，我首先尝试了TVM。


### 为什么选择TVM
　　为提升深度学习模型的推理效率，设备平台制造商针对自己的平台推出优化的推理引擎，例如NAVIDA的tensorRT，Intel的OpenVINO，Tencent针对移动端应用推出NCNN等。目前，深度学习模型应用广泛，在服务端和移动端都有应用，甚至于特殊的嵌入式场景想，它们都有加速模型推理的需求。个人感觉针对不同平台选择不同的推理引擎，学习成本太高。我这里选择尝试TVM，主要有以下几个原因：
- 尝试了过一些模型压缩方法，效率提升有限
- 有些是模型压缩方法需要推理引擎和硬件的支持的，例如量化
- tensorflow推理效率有限，需要更好的推理引擎
- 针对平台选择不同推理引擎，学习成本太高
- 需要能支持跨平台的推理引擎，未来可能在定制的嵌入式芯片上运行深度学习模型
- 除了TVM之外，还存在XLA之类方案，选择TVM也是因为tianqi等大佬主导的项目，相信大佬！


### 初次体验TVM，相比于tensorflow2倍的性能提升
　　看了几篇TVM介绍文章后，了解到它是从深度学习编译器的角度来做推理引擎，目前技术领域还比较新，具体技术细节以后有机会会深入学习，这里主要想体验一下使用TVM做深度模型推理，重点是推理效率的提升，因为是骡子还是马得拉出来遛遛。参考官方文档进行编译安装，整个过程还是比较简单的，结果显示相比于tensorflow大概100%的性能提升。实验环境是ubuntu 19.04，x86_64架构。
1. 安装llvm,也可源码编译
```
$ sudo apt-get install llvm
```
2. 编译TVM
```
$ git clone --recursive https://github.com/dmlc/tvm.git
$ cd tvm $$ mkdir build
$ cp cmake/config.cmake build
# 编辑config.cmake 然后将USE_LLVM OFF 改为 set(USE_LLVM /usr/bin/llvm-config)
$ cd build
$ cmake ..
$ cmake -j4
```
3. 编辑.bashrc配置Python环境
```
export TVM_HOME=/home/xxxx/code/tvm
export PYTHONPATH=$TVM_HOME/python:$TVM_HOME/topi/python:$TVM_HOME/nnvm/python
```
4. 官方[Compile Tensorflow Models](https://docs.tvm.ai/tutorials/frontend/from_tensorflow.html#sphx-glr-tutorials-frontend-from-tensorflow-py)
直接运行出现了两个问题，下载文件时和SSL相关，另外一个是缺少antlr
```
# install antlr
$ pip install antlr4-python3-runtime
# debug ssl
import ssl
ssl._create_default_https_context = ssl._create_unverified_context
# run demo
$ python from_tensorflow.py
```
5. 在代码中加入时间测试，实验测试结果。TVM与测试时间为0.277s，tensorflow为0.586s。
```
============ TVM ============ 0.2770531177520752
African elephant, Loxodonta africana (score = 0.58335)
tusker (score = 0.33901)
Indian elephant, Elephas maximus (score = 0.02391)
banana (score = 0.00025)
vault (score = 0.00021)
============= Tensorflow ===== 0.58619508743286133
===== TENSORFLOW RESULTS =======
African elephant, Loxodonta africana (score = 0.58394)
tusker (score = 0.33909)
Indian elephant, Elephas maximus (score = 0.03186)
banana (score = 0.00022)
desk (score = 0.00019)

```

## 未填的坑
　　过程遇到一个坑，查了TVM社区，没有很好的解答，看起来好像会和性能有关，希望路过的大佬能帮忙解决。https://discuss.tvm.ai/t/cannot-find-config-for-target-llvm-when-using-autotvm-in-tensorflow-example-for-cpu/1544
```
WARNING:autotvm:Cannot find config for target=llvm, workload=('conv2d', (1, 8, 8, 2048, 'float32'), (1, 1, 2048, 384, 'float32'), (1, 1), (0, 0), (1, 1), 'NHWC', 'float32'). A fallback configuration is used, which may bring great performance regression.
WARNING:autotvm:Cannot find config for target=llvm, workload=('conv2d', (1, 8, 8, 2048, 'float32'), (1, 1, 2048, 448, 'float32'), (1, 1), (0, 0), (1, 1), 'NHWC', 'float32'). A fallback configuration is used, which may bring great performance regression.
WARNING:autotvm:Cannot find config for target=llvm, workload=('conv2d', (1, 8, 8, 2048, 'float32'), (1, 1, 2048, 192, 'float32'), (1, 1), (0, 0), (1, 1), 'NHWC', 'float32'). A fallback configuration is used, which may bring great performance regression.

```

参考：
1. tvm install: https://docs.tvm.ai/install/from_source.html
2. tvm tutorial: [Compile Tensorflow Models](https://docs.tvm.ai/tutorials/frontend/from_tensorflow.html#sphx-glr-tutorials-frontend-from-tensorflow-py)
3. 未填的坑：https://discuss.tvm.ai/t/cannot-find-config-for-target-llvm-when-using-autotvm-in-tensorflow-example-for-cpu/1544