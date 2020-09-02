---
title: 了解tensorflow中的模型基准测试工具
categories:
- AI
mathjax: true
---


	
　　深度学习模型落地需要考虑决定推理（inference）过程所需的计算资源（成本）和效率（系统的吞吐量和延时），有时甚至需要进行适当的模型裁剪和压缩工作。理论上说，模型结构一旦确定是可以计算它的复杂度和计算量，但这有些繁琐。实际中可以借助一些工具帮助预估模型实际的性能，比较模型优化前后的差别，主要使用到的是benchmark_model和summarize_graph。


## 一、benchmark_model模型推理速度分析
　　在深度学习模型工程落地时，我们追求在成本可控的前提下提高良好的用户体验，因此模型的推理效率和计算代价是重要的衡量指标。通常用FLOPs（floating point operations）描述模型的计算力消耗，它表示浮点运算计算量，用来衡量算法/模型的复杂度。我们是可以从原理上计算出模型需要的FLOPs，参考：https://www.zhihu.com/question/65305385。 除了从理论计算之外，还可以使用tensorflow中的 benchmark_model 工具来进行粗略估计，它可以帮助估算出模型所需的浮点操作数(FLOPS)，然后你就可以使用这些信息来确定你的模型在你的目标设备上运行的可行性。除此之外，比较容易混淆的概念是FLOPS（floating point operations per second），意指每秒浮点运算次数，理解为计算速度，它是衡量硬件性能的指标对于来说TESLA P40可以每秒处理12T个FLOP，普通单核CPU每秒大概处理100亿次的FLOP。当有了计算操作消耗的估计之后，它就对你计划的目标设备上有所帮助，如果模型的计算操作太多，那么就需要优化模型减小FLOP数量。

　　例如下面的例子中，我们通过benchmark_model分析resetNet20-cifar10，大概有82.15M的FLOPs，该机器每秒执行21.89B，因此该模型大概需要4ms的计算时间。在使用benchmark_model之前，需要使用tensorflow源码进行编译。

```
编译benchmark_model
$ bazel build -c opt tensorflow/tools/benchmark:benchmark_model
$ bazel-bin/tensorflow/tools/benchmark/benchmark_model \
--graph=model_original.pb \
--input_layer="net_input" \
--input_layer_shape="1,32,32,3" \
--input_layer_type="float" \
--output_layer="net_output" \
--show_flops=true \
--show_run_order=false \
--show_time=false \
--num_threads=1
```


#### 预估FLOPs
```
2019-10-11 21:30:31.288678: I tensorflow/tools/benchmark/benchmark_model.cc:636] FLOPs estimate: 82.15M
2019-10-11 21:30:31.288744: I tensorflow/tools/benchmark/benchmark_model.cc:638] FLOPs/second: 21.89B
```


#### 查看不同类型节点消耗的时间：
```
========================= Summary by node type ==========================================
 [Node type]	  [count]	  [avg ms]	    [avg %]	    [cdf %]	  [mem KB]	[times called]
          <>	       65	     4.110	    47.269%	    47.269%	     0.000	       65
FusedBatchNorm	       19	     2.028	    23.324%	    70.592%	   240.384	       19
      Conv2D	       22	     2.003	    23.036%	    93.629%	   868.352	       22
         Pad	        2	     0.239	     2.749%	    96.377%	   115.456	        2
        Relu	       19	     0.082	     0.943%	    97.320%	     0.000	       19
       Const	       65	     0.071	     0.817%	    98.137%	     0.000	       65
        NoOp	        1	     0.066	     0.759%	    98.896%	     0.000	        1
         Add	        9	     0.059	     0.679%	    99.574%	     0.000	        9
        Mean	        1	     0.010	     0.115%	    99.689%	     0.256	        1
     Softmax	        1	     0.008	     0.092%	    99.781%	     0.000	        1
_FusedMatMul	        1	     0.007	     0.081%	    99.862%	     0.040	        1
     _Retval	        1	     0.005	     0.058%	    99.919%	     0.000	        1
     Squeeze	        1	     0.005	     0.058%	    99.977%	     0.000	        1
        _Arg	        1	     0.002	     0.023%	   100.000%	     0.000	        1

Timings (microseconds): count=1000 first=7287 curr=7567 min=7198 max=18864 avg=8794.03 std=1249
Memory (bytes): count=1000 curr=1224488(all same)
```

- node type：进行操作的节点类型。
- start：运算符的启动时间，展示了其在操作顺序中的位置。
- first: 以毫秒为单位。默认情况下 TensorFlow 会执行 20 次运行结果来获得统计数据，这个字段则表示第一次运行基准测试所需的操作时间。
- avg ms：以毫秒为单位。表示整个运行的平均操作时间。
- %：一次运行占总运行时间的百分比。这对理解密集计算区域非常有用。
- cdf%：整个过程中表格中当前运算符及上方全部运算符的累积计算时间。这对理解神经网络不同层之间的性能分布非常重要，有助于查看是否只有少数节点占用大部分时间。
- mem KB：当前层消耗的内存大小。
- Name：节点名称。


## 二、summarize_graph 模型大小分析
　　服务端深度模型落地时主要关注模型的预测效率，移动端模型落地需要考虑模型的大小。通过summarize_graph工具可以帮助我们简要分析模型的参数量和包含哪些op。设置--print_structure=true可以观察到模型的结构，这也可以通过tensorboard来可视化实现。
```
tensorflow-1.14.0编译summarize_graph工具
$ bazel build -c opt tensorflow/tools/graph_transforms:summarize_graph
$ bazel-bin/tensorflow/tools/graph_transforms/summarize_graph \
--in_graph=reset20_cifar10_original.pb \
--print_structure=true

```

```
    Found 1 possible inputs: (name=net_input, type=float(1), shape=[?,32,32,3]) 
    No variables spotted.
    Found 1 possible outputs: (name=net_output, op=Softmax) 
    Found 272572 (272.57k) const parameters, 0 (0) variable parameters, and 0 control_edges
    Op types used: 194 Const, 77 Identity, 22 Conv2D, 19 Relu, 19 FusedBatchNorm, 11 Add, 6 Slice, 5 Pad, 5 Reshape, 4 Sub, 4 MatchingFiles, 3 Switch, 2 Squeeze, 2 ShuffleDataset, 2 ShuffleAndRepeatDataset, 2 StridedSlice, 2 Shape, 2 TensorSliceDataset, 2 RealDiv, 2 PrefetchDataset, 2 ParallelMapDataset, 2 ParallelInterleaveDataset, 2 Transpose, 2 OneHot, 2 BatchDatasetV2, 2 Cast, 2 Maximum, 2 DecodeRaw, 1 GreaterEqual, 1 All, 1 Assert, 1 BiasAdd, 1 Softmax, 1 ExpandDims, 1 FixedLengthRecordDataset, 1 FloorMod, 1 Mul, 1 ReverseV2, 1 Less, 1 MatMul, 1 RandomUniformInt, 1 RandomUniform, 1 Mean, 1 Placeholder, 1 Merge
```


https://tensorflow.juejin.im/mobile/optimizing.html