---
title: tensorflow通道剪枝(channel-pruning)实战
categories:
- other
mathjax: true
---


## 一、概述
　　最近在做模型压缩(model compress)相关工作，之前分别尝试了权重量化(weight quantization)和权重稀疏(weight sparsification)，遗憾的是它们都需要推理引擎和硬件的特定优化才能实现推理加速的效果，而tensorflow在x86架构的CPU下并没有没有针对量化和稀疏矩阵的优化，因此效果一般。吸取前两次的经验，这次尝试了结构化压缩通道剪枝(channel pruning)，它通过删减模型中融入的通道信息，减少的模型前向计算FLOPs，实现推理加速。通道剪枝来自论文ICCV2017论文 Channel Pruning for Accelerating Very Deep Neural Networks。 这里会首先简单介绍channel pruning的原理，然后通过PocketFlow压缩工具对ResNet56进行通道剪枝，结果显示channel pruning在精度不怎么损失的基础上，模型近50%的FLOPs。由于剪枝后模型中存在许多的conv2d 1x1卷积，实际提升推理效率大概20%。

## 二、channel pruning 基本原理
　　虽然论文末尾谈到channel pruning可以应用到模型训练中，但是文章的核心内容还是对训练好的模型进行channel pruning，也是文章中说的inference time。通常剪枝正如其名字channel pruning核心思想是移除一些冗余的channel简化模型。


![channel-pruning示意图](/images/channel_pruning.jpg)


https://blog.csdn.net/u014380165/article/details/79811779

## 三、pocketflow 实践

## 四、channel pruning前后模型分析 

## 四、conv2d 1X1 在channel pruning中的引用




四、问题汇总

参考：
https://blog.csdn.net/u014380165/article/details/79811779



time ./scripts/run_local.sh nets/resnet_at_cifar10_run.py     --learner channel  --exec_mode eval --save_path ./models/original_model.ckpt


time ./scripts/run_local.sh nets/resnet_at_cifar10_run.py  --learner channel  --exec_mode eval --save_path ./models/pruned_model.ckpt


conv1x1 在 channel pruing中的应用。


参考：
[1].[Channel Pruning for Accelerating Very Deep Neural Networks](https://arxiv.org/abs/1707.06168)
[2].[PocketFLow](https://github.com/wxquare/PocketFlow)
