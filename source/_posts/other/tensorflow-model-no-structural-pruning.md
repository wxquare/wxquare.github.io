---
title: tensorflow模型权重稀疏(weight sparsification)实战
categories:
- other
mathjax: true
---



## 一、概述
　　深度模型通常会有更好的预测精度，但是它面临计算开销过大的问题。模型压缩(model compress)是提高深度模型推理效率的一种解决方案，它期望在不损失精度或者精度损失可控的范围内，加速推理效率，减低内存开销。目前，模型压缩算法主要包括权**重量化(quantization)、剪枝(pruning)、低秩分解等**。上周尝试了[tensorflow中的模型量化](https://wxquare.github.io/2019/09/16/other/tensorflow-model-quantization/)，发现量化需要硬件或者推理引擎的对低精度8-bit计算支持，目前tensorflow在x86和gpu环境下还没有很好的支持，因此量化只帮助实现了模型大小下降，没有实现推理的加速。model pruning学习的材料是tensorflow repo中的tensorflow/contrib/model_pruning，实际了解后发现它属于pruning中no-structural pruning，其加速效果依赖具体的硬件实现，加速效果一般，tensorflow 中对稀疏矩阵运算没有特别好的优化（依赖于底层的 SparseBLAS 实现，目前还没有特别好的）。model pruning中还有一种structural pruning 则不改变计算方式，可以直接使用，加速效果相对较好，之后也会继续尝试。


## 二、tensorflow/contrib/model_pruning原理
　　[Michael Zhu and Suyog Gupta, “To prune, or not to prune: exploring the efficacy of pruning for model compression”, 2017 NIPS ](https://arxiv.org/pdf/1710.01878.pdf) 
　　tensorflow中model_pruning理论来自上面这篇文章。文章中指出目前有些深度学习网络模型是过度设计（over-parameterized）。为了使其在资源受限的环境下高效的进行推理预测，要么减少网络的隐藏单元（hidden unit）同时保持模型密集连接结构，要么采用针对大模型进行模型剪枝（model pruning）。文章中的模型行剪枝是一种非结构化的剪枝（no-structural pruning），它在深度神经网络的各种连接矩阵中引入稀疏性（sparsity），从而减少模型中非零值参数的数量。文章比较了大而稀疏（large-sparse）和较小密集（small-dense）这两种模型，认为前者是优于后者的。除此之外，文章提出了一种新的渐进剪枝技术（gradual pruning technique），它能比较方便的融入到模型训练的过程中，使其调整比较小。


　　tensorflow中的模型剪枝是一种训练时剪枝。对于需要被剪枝的网络模型，对于网络中每个需要被剪枝的层（layer)添加一个二进制掩码变量（binary mask variable ），该变量的大小和形状和改层的权重张量（weight tensor）相同。在训练图中加入一些ops，它负责对该层的权重值（weights）的绝对值进行排序，通过mask将最小的权重值屏蔽为0。在前向传播时该掩模的对应位与选中权重进行相与输出feature map，如果该掩模对应位为0则对应的权重相与后则为0，在反向传播时掩模对应位为0的权重参数则不参与更新。除此之外，文章提出了一种新的自动逐步修剪算法（automated gradual pruning），它实际上是定义了一种稀疏度变化的规则，初始时刻，稀疏度提升较快，而越到后面，稀疏度提升速度会逐渐放缓，这个主要是基于冗余度的考虑。因为初始时有大量冗余的权值，而越到后面保留的权值数量越少，不能再“大刀阔斧”地修剪，而需要更谨慎些，避免“误伤无辜”。其表达式如下，官方文档中列出了一些的剪枝超参数，主要的有下面几个。
$$s\_{t}=s\_{f}+\left(s_{i}-s_{f}\right)\left(1-\frac{t-t_{0}}{n\Delta t}\right)^{3}  $$

- initial_sparsity：初始稀疏值$s_i$
- target_sparsity：目标稀疏值$s_f$
- sparsity_function_begin_step：开始剪枝的step $t_0$
- sparsity_function_end_step: 剪枝停止的step
- pruning_frequency：剪枝的频率$\Delta t$，文章提出在100到1000之间通常比较好
- sparsity_function_exponent: 剪枝函数的指数，表示式中已描述为默认的3，表示由快到慢，为1时表示线性剪枝



## 三、tensorflow中的model_pruning实践
　　tensorflow中model_pruning的源码位于tensorflow/contrib/model_pruning。
1. 准备tensorflow-1.14.0源码
2. 编译model_pruning
```
$bazel build -c opt tensorflow/contrib/model_pruning/examples/cifar10:cifar10_train
```
3. 通过设置一些参数，开始针对cifar10剪枝
```
$bazel-out/k8-py2-opt/bin/tensorflow/contrib/model_pruning/examples/cifar10/cifar10_train \
--train_dir=/home/terse/code/programming/tensorflow/model_pruning/train \
--pruning_hparams=name=cifar10_pruning,\
initial_sparsity=0.3,\
target_sparsity=0.9,\
sparsity_function_begin_step=100,\
sparsity_function_end_step=10000
```

4. 可通过tensorboard查看剪枝过程。可以清楚的看出随着训练步骤的增加，conv1和conv2的sparsity在不断的增长。 在GRAPHS 页面，双击conv节点，可以看到在原有计算图基础上新增了mask和threshold节点用来做 model pruning
```
$tensorboard --logdir=/home/terse/code/programming/tensorflow/model_pruning/train
```

5. 模型剪枝之后将剪枝的ops从训练图中删除。
```
$bazel build -c opt tensorflow/contrib/model_pruning:strip_pruning_vars
$bazel-out/k8-py2-opt/bin/tensorflow/contrib/model_pruning/strip_pruning_vars \
--checkpoint_dir=/home/terse/code/programming/tensorflow/model_pruning/train \
--output_node_names=softmax_linear/softmax_linear_2 \
--output_dir=/home/terse/code/programming/tensorflow/model_pruning \
--filename=pruning_stripped.pb
```

## 四、model_pruning源码简单分析
　　使用tensorflow的model_pruning进行模型剪枝，主要包括两方面的工作，一是apply_mask，二是在训练图中增加剪枝的节点（pruning ops）。这里分别截取了其中的两段代码。
```
  # cifar10_pruning.py  apply_mask to the graph
  with tf.variable_scope('conv1') as scope:
    kernel = _variable_with_weight_decay(
        'weights', shape=[5, 5, 3, 64], stddev=5e-2, wd=0.0)

    conv = tf.nn.conv2d(
        images, pruning.apply_mask(kernel, scope), [1, 1, 1, 1], padding='SAME')
    
    biases = _variable_on_cpu('biases', [64], tf.constant_initializer(0.0))
    pre_activation = tf.nn.bias_add(conv, biases)
    conv1 = tf.nn.relu(pre_activation, name=scope.name)
    _activation_summary(conv1)
```

```
	 #Adding pruning ops to the training graph
	with tf.graph.as_default():
	
	  # Create global step variable
	  global_step = tf.train.get_or_create_global_step()
	
	  # Parse pruning hyperparameters
	  pruning_hparams = pruning.get_pruning_hparams().parse(FLAGS.pruning_hparams)
	
	  # Create a pruning object using the pruning specification
	  p = pruning.Pruning(pruning_hparams, global_step=global_step)
	
	  # Add conditional mask update op. Executing this op will update all
	  # the masks in the graph if the current global step is in the range
	  # [begin_pruning_step, end_pruning_step] as specified by the pruning spec
	  mask_update_op = p.conditional_mask_update_op()
	
	  # Add summaries to keep track of the sparsity in different layers during training
	  p.add_pruning_summaries()
	
	  with tf.train.MonitoredTrainingSession(...) as mon_sess:
	    # Run the usual training op in the tf session
	    mon_sess.run(train_op)
	
	    # Update the masks by running the mask_update_op
	    mon_sess.run(mask_update_op)

```
## 五、总结和未解决的问题
1. tensorflow中的模型剪枝属于no-structral，本质上是使权重稀疏化(weight sparsification),实践中发现它没有使推理加速，据其加速效果依赖具体的硬件实现，加速效果一般，tensorflow 中对稀疏矩阵运算没有特别好的优化（依赖于底层的 SparseBLAS 实现，目前还没有特别好的）
2. 实践中发现不管稀疏度为多少，其剪枝后的模型大小都是相同的，是不是tensorflow对稀疏的模型也是按照非稀疏格式存储的？
3. issue:[model_pruning: Why 50% and 90% zeros of the stripped models are the same size? #32805](https://github.com/tensorflow/tensorflow/issues/32805)
4. issue: [CNN.Model pruning: no gain in speeding up of inference #22732](CNN.Model pruning: no gain in speeding up of inference #22732)


参考：
1. [https://github.com/tensorflow/tensorflow/tree/r2.0/tensorflow/contrib/model_pruning](https://github.com/tensorflow/tensorflow/tree/r2.0/tensorflow/contrib/model_pruning)
2. [Michael Zhu and Suyog Gupta, “To prune, or not to prune: exploring the efficacy of pruning for model compression”, 2017 NIPS ](https://arxiv.org/pdf/1710.01878.pdf)
3. https://zhuanlan.zhihu.com/p/48069799