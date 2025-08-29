---
title: TVM学习笔记--模型量化(int8)及其测试数据
date: 2020-08-13
categories:
- AI
mathjax: true
---

　　坚持了接近一年的视频算法相关的项目，老板最终还是喊停了。并没有感到特别意外，只是在对一个东西突然有些兴趣或者说入门的时候，戛然而止，多少有些不甘心和遗憾，但是以后会在业余继续学习的，也希望自己在2020年能把工作逐渐聚焦到这块吧。

　　接触TVM到有两个原因。一是需要支持多种优化手段的推理引擎，例如量化、图优化、稀疏优化、模型压缩剪枝等。尝试过在tensorflow的quantization和非结构性剪枝(no-structural pruning)，加速效果非常一般，因为这些优化手段需要推理引擎的支持，但是当时我们都是纯后台出身，也没人掌握这个内容。再之后尝试channel pruning，终于取得了一些进展，但是30%的提升leader并不满意。二是需要支持多种平台的推理引擎，例如NV GPU/x86/ARM GPU等。由于组内业务迟迟没有好的落地场景，尝试了多种手段，需要的把深度模型部署在不同的平台上。记得有次，花了两周的时间把DaSiamRPN模型移植到终端上。从零开始pytorch、onnx、tflite、android，期间踩了许多的坑，结果在移动端运行需要4秒时间来处理一帧图像。。。期间同事也曾通过tensorRT部署模型，效率反而下降。一次偶然的机会了解到TVM，当时感觉它可能是比较适合我们团队的需求的。


　　由于我之前学习信号处理的，比较容易理解量化。模型量化quantization也在深度学习在部署落地时提高效率的常用的方法。之前有写过关于[tensorflow模型量化](https://zhuanlan.zhihu.com/p/86440423)的方法，写得不好，对于想学习模型量化知识的可以参考下面链接进行学习：

**模型量化相关：**
【1】[神经网络量化简介](https://jackwish.net/2019/neural-network-quantization-introduction-chn.html)
【2】[Tensort量化:8-bit-inference-with-tensort](http://on-demand.gputechconf.com/gtc/2017/presentation/s7310-8-bit-inference-with-tensorrt.pdf)
【3】[阮一峰：浮点数的二进制表示](http://www.ruanyifeng.com/blog/2010/06/ieee_floating-point_representation.html)
【4】[Quantizing deep convolutional networks for efficient inference](https://arxiv.org/pdf/1806.08342.pdf)

**TVM量化相关RFC**
【INT8 quantization proposal】：https://discuss.tvm.ai/t/int8-quantization-proposal/516（2018.02.02）
【TVM quantizationRFC】 https://github.com/apache/incubator-tvm/issues/2259(2018.12.09)

	
　　目前，官网上还没有关于模型量化的教程和文档，对于刚接触新手来说可能有些麻烦，这里提供提供一个参考代码，方便新手学习。除此之外，也测试了TVM的int8量化性能，结果显示TVM的量化加速效果不是很好，甚至略有下降，需要配合autotvm一起使用。[测试代码地址](https://github.com/wxquare/programming/tree/master/blog/TVM_quantization)。测试结果如下，希望对大家了解TVM有帮助。


| 模型 | 原始框架 | 原始框架运行时间 | TVM FP32 | TVM int8 | TVM int8+AutoTVM |
| ------ | ------ | ------ | ------ | ------ | ------ |
|resnet18v1| mxnet 1.5.1|	27.8ms | 46.9ms| 51.10ms | 25.83ms |
|Inceptionv1| tensorflow 1.13 |	560ms | 164ms| 185ms | 116ms |
