---
title: tensorflow中的模型剪枝(channel-pruning)实战
categories:
- other
mathjax: true
---


论文：
## 一、概述
   最近在一直在做模型压缩相关工作，之前分别尝试了权重量化(weight quantization)和权重稀疏(weight sparsification)，但是这两种压缩方法都是需要推理引擎和硬件的优化才能加速推理过程，而tensorflow在推理的时候没有针对量化和稀疏矩阵的优化，因此效果一般。吸取前两次的经验，这次尝试了结构化压缩channel pruning，它能直接减少的模型计算的FLOPs，不依赖于推理引擎。channel pruning来自论文ICCV2017论文 Channel Pruning for Accelerating Very Deep Neural Networks。 

	

## 二、channel pruning 原理分析
https://blog.csdn.net/u014380165/article/details/79811779

## 三、pocketflow channel pruning实践

## 四、channel pruning前后模型分析 

## 四、conv2d 1X1 在channel pruning中的引用




四、问题汇总

参考：
https://blog.csdn.net/u014380165/article/details/79811779



time ./scripts/run_local.sh nets/resnet_at_cifar10_run.py     --learner channel  --exec_mode eval --save_path ./models/original_model.ckpt


time ./scripts/run_local.sh nets/resnet_at_cifar10_run.py  --learner channel  --exec_mode eval --save_path ./models/pruned_model.ckpt


conv1x1 在 channel pruing中的应用。

