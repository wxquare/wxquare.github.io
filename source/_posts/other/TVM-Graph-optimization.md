---
title: TVM学习笔记--图优化和relay(Graph-level optimization)
categories:
- other
mathjax: true
---


 TVM优化主要包括两个级别，图优化(graph-level) 和算子优化（operator)。这里主要想总结一下最近了解到的关于TVM的图优化。深度学习模型通常都是通过计算图（computation graph)来描述运算过程的。计算图中的节点表示各种同的算子(opertor),边表示算子之间的依赖关系。TVM论文中提到深度学习模型的计算图和编译器的中间描述(IR)很相似，只是计算图中数据通常是多维的tensor。顺着这个思路，通过一些优化手段，也可以把计算图做功能等价的变换，实现优化性能。学习TVM图优化，我从下面三个内容进行：
- TVM是怎么表示计算图的的？
- TVM在图优化上做了哪些工作？以及效果怎么样？
- 怎么添加优化的pass？
 
## 一、TVM是怎么表示计算图的。
一、relay介绍。Introduction to Relay IR
二、adding an operator to relay
三、Relay Pass Infrastructure
四、Adding a compiler Pass to Relay
五、(Relay: A New IR for Machine Learning Frameworks)[https://arxiv.org/pdf/1810.00952.pdf]

## 二、TVM已经做了哪些图优化的方法。






