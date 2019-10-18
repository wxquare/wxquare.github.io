---
title: tensorflow中的模型剪枝(channel-pruning)
categories:
- other
mathjax: true
---



## 一、channel pruning 介绍

## 二、channel pruning 原理分析
https://blog.csdn.net/u014380165/article/details/79811779

## 三、pocketflow channel pruning实践


四、问题汇总

参考：
https://blog.csdn.net/u014380165/article/details/79811779



time ./scripts/run_local.sh nets/resnet_at_cifar10_run.py     --learner channel  --exec_mode eval --save_path ./models/original_model.ckpt


time ./scripts/run_local.sh nets/resnet_at_cifar10_run.py  --learner channel  --exec_mode eval --save_path ./models/pruned_model.ckpt


conv1x1 在 channel pruing中的应用。

