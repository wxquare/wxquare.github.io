---
title: tensorflow模型权重量化(weight quantization)实战
categories:
- other
mathjax: true
---

　　
　　最近在尝试深度学习模型加速的工作，查了一些资料，发现模型推理加速的研究还挺多的，主要从四个方面进行，从头开始构建轻量高效的模型，例如mobileNets、squeezenet等；通过量化(quantization)、裁剪(pruning)和压缩(compression)来降低模型的尺寸；通过高效的计算平台加速推理(inference)的效率，例如Nvidia TensorRT、GEMMLOWP、Intel MKL-DNN等以及硬件定制。考虑到自身的能力，遵循从简单到复杂、通用到专用的原则，选择从模型量化(model quantization)入手，之后会陆续尝试其他优化手段。在一番尝试之后，挺遗憾的，因为tensorflow模型量化并没有使模型预测(inference)加速，根据tf成员在issue的回复，tf的模型量化主要针对移动端的优化，目前还没有针对x86和gpu环境的优化。**有成功通过模型量化加速推理过程的同学欢迎打脸留言**。


## 一、为什么要模型量化
　　为了尽可能保证深度学习模型的准确度(precision)，在训练和推理时候通常使用float32格式的数据。然而在实际商用中，有些模型由于层数和参数都比较多，推理预测需要很大计算量，导致推理(inference)的效率很低。模型量化(model quantization)是通用的深度学习优化的手段之一，它通过将float32格式的数据转变为int8格式，一方面降低内存和存储的开销，同时在一定的条件下(8-bit低精度运算 low-precision)也能提升预测的效率。目前还不太理解8-bit低精度运算，猜测这是模型量化没有实现推理加速的原因。模型量化适用于绝大数模型和使用场景，对于训练后的量化，不需要重新训练模型，可以很快将其量化为定点模型，而且几乎不会有精度损失，因此模型量化追求更小的模型和更快的推理速度。**实验中量化确实时模型下降为原来的1/4，但在推理效率并没有提升，甚至略有下降**。

## 二、什么是量化
### 2.1 实数量化
　　网络上关于模型量化的内容挺多的，量化本质上是一种仿射图(affine map)，它以表达式(1)将实数值表示映射为量化的uint8，当然也可以等效为表示式(2): 
```
real_value = A * quantized_value + B             (1) 
real_value = C * (quantized_value + D)           (2) 
```

　　除此之外，深度学习模型量化中有一个**约束条件，0必须准确的表示，不能有误差**。因为对于某些神经网络层，实数0精确表示对于优化实现非常有用，例如在具有填充的卷积层或池化层中，长度对输入数组进行零填充(zero-padding)来实现填充是有用的。实数值0对应的量化值称之为零点(zero-point)。实际上，如果0不能完全表示，当我们用0对应的量化值进行填充时，因为这与实际值0不完全对应，会导致结果不准确，引入偏差。因此有：
```
　　0=A∗zero_point+B
　　zero_point=−B/A
　　0=C∗(zero_point+D)
　　0=zero_point+D
　　D=−zero_point
```



　　结合上述条件，可以得出量化的最终表达式为(3)，它能做到0值的准确表示，zero_point是0对应的量化值。表示式(3)中有两个常量，zero_point是量化值，通常是uint8值，scale是一个正实数，通常为float32。
$$real\\_value = scale \* (quantized\\_value - zero\\_point)　　(3)$$

### 2.2 矩阵乘法量化
　　根据表达式(3)，我们可以将实数值(通常为float32)用量化值(通常为uint8)表示，下面将介绍怎么把它应用到矩阵乘法当中。假设有两个实数矩阵$lhs\\_real\\_matrix, rhs\\_real\\_matrix$，量化之后就会有对应的$lhs\\_scale, rhs\\_scale, lhs\\_zero\\_point, rhs\\_zero\\_point$，矩阵中的实数值可以用其量化值表示为：
```
	lhs_real_value[i] = lhs_scale * (lhs_quantized_value[i] - lhs_zero_point)
	rhs_real_value[i] = rhs_scale * (rhs_quantized_value[i] - rhs_zero_point)
```
　　在矩阵乘法中，每个值($result\\_real\\_value$)都由对应的ｉ个值相乘累加得到，根据表达式(4)和(5)很容易得到表示式(6),它表示$result\\_quantized\\_value$可由$lhs\\_quantized\\_value、rhs\\_quantized\\_value$计算得出。注意这里面有几个问题需要解决，如何减小式(6)中与zero_point减法的开销(overhead)？如何将(lhs_scale * rhs_scale / result_scale)实数运算用整数运算处理？这部分的内容参考gemmlowp的实现。
　　https://github.com/google/gemmlowp/blob/master/doc/quantization.md
```
result_real_value
  = Sum_over_i(lhs_real_value[i] * rhs_real_value[i])
  = Sum_over_i(
        lhs_scale * (lhs_quantized_value[i] - lhs_zero_point) *
        rhs_scale * (rhs_quantized_value[i] - rhs_zero_point)
    )
  = lhs_scale * rhs_scale * Sum_over_i(
        (lhs_quantized_value[i] - lhs_zero_point) *
        (rhs_quantized_value[i] - rhs_zero_point)
    )                    (4)

result_real_value = result_scale * (result_quantized_value - result_zero_point)
result_quantized_value = result_zero_point + result_real_value / result_scale  (5)

result_quantized_value = result_zero_point +
    (lhs_scale * rhs_scale / result_scale) *
        Sum_over_i(
            (lhs_quantized_value[i] - lhs_zero_point) *
            (rhs_quantized_value[i] - rhs_zero_point)
        )          (6)

```

## 三、tensorflow模型量化方案

　　**训练后量化(post training Quantization)**。在许多情况下，我们希望在不重新训练模型的前提下，只是通过压缩权重或量化权重和激活输出来缩减模型大小，从而加快预测速度。“训练后量化”就是这种使用简单，而且在有限的数据条件下就可以完成量化的技术。训练后量化操作简单，只需要使用量化工具将训练好的模型执行量化类型，即可实现模型的量化。训练后量化包括“只对权重量化”和“对权重和激活输出都量化”，对于很多网络而言，都可以产生和浮点型很接近的精度。


　　**只对权重量化(weight only quantization)**。一个简单的方法是只将权重的精度从浮点型减低为8bit整型。由于只有权重进行量化，所以无需验证数据集就可以实现。一个简单的命令行工具就可以将权重从浮点型转换为8bit整型。如果只是想为了方便传输和存储而减小模型大小，而不考虑在预测时浮点型计算的性能开销的话，这种量化方法是很有用的。

　　**量化权重和激活输出（Quantizing weights and activations）**。我们可以通过计算所有将要被量化的数据的量化参数，来将一个浮点型模型量化为一个8bit精度的整型模型。由于激活输出需要量化，这时我们就得需要标定数据了，并且需要计算激活输出的动态范围，一般使用100个小批量数据就足够估算出激活输出的动态范围了。

　　**训练时量化（Quantization Aware Training)**。训练时量化方法相比于训练后量化，能够得到更高的精度。训练时量化方案可以利用Tensorflow的量化库，在训练和预测时在模型图中自动插入模拟量化操作来实现。由于训练时量化相对麻烦，加上权重量化没有实现加速的期望，所以没有尝试训练时量化，根据文档显示，其大概包括以下几个步骤：
1. 可以在预训练好的模型基础上继续训练或者重新训练，建议在保存好的浮点型模型的基础上精调
2. 修改估计器，添加量化运算，利用tf.contrib.quantize中的量化rewriter向模型中添加假的量化运算
3. 训练模型，输出对于权重和激活输出都带有各自量化信息（尺度、零点）的模型
4. 转换模型，利用tf.contrib.lite.toco convert定义的转换器，将带有量化参数的模型被转化成flatbuffer文件，该文件会将权重转换成int整型，同时包含了激活输出用于量化计算的信息
5. 执行模型，转换后的带有整型权重的模型可以利用TFLite interpreter来执行，也可以在CPU上运行模型


## 四、tensorflow模型权重量化实验
　　一开始尝试模型量化是因为有个复杂的视频分割模型推理效率很低，期望通过模型量化实现加速，在复杂模型上尝试失败之后，我用label_image的例子再次验证，结果显示也没有加速的效果。这里主要试验了训练后量化，尝试了只对权重量化和权重和激活量化，发现后者比前者性能更差，这里描述权重量化的过程。整个过程是比较简单的，tensorflow有两种量化方式，推荐使用第二种，编译命令行工具进行量化。
1. 在tensorflow r1.0的版本中有个量化的脚本可以提供量化的功能：
```
$wget "https://storage.googleapis.com/download.tensorflow.org/models/inception_v3_2016_08_28_frozen.pb.tar.gz"
$tar -xzf tensorflow/examples/label_image/data
$ work_dir=/home/terse/code/programming/tensorflow/quantization
$ python tensorflow/tools/quantization/quantize_graph.py \
--input=$work_dir/inception_v3_2016_08_28_frozen.pb \
--output=$work_dir/inception_quantized0.pb \
--output_node_names=InceptionV3/Predictions/Reshape_1 \
--mode=weights 
```

2. 在较新版本的tf中，quantize_graph.py量化的脚本已经废弃了需要编译tensorflow的源码生成
```
tensorflow-1.14.0编译transform_graph工具
$ bazel build tensorflow/tools/graph_transforms:transform_graph
$ bazel-bin/tensorflow/tools/graph_transforms/transform_graph \
--in_graph=$work_dir/inception_v3_2016_08_28_frozen.pb \
--out_graph=$work_dir/inception_quantized1.pb \
--outputs=InceptionV3/Predictions/Reshape_1 \
--transforms='quantize_weights'
```

3. 使用summarize_graph分析量化前后的模型区别，权重量化、模型减小、增加了一些和量化和反量化的节点。
``` 
tensorflow-1.14.0编译transform_graph工具
$ bazel build tensorflow/tools/graph_transforms:summarize_graph
$ bazel-bin/tensorflow/tools/graph_transforms/summarize_graph \
--in_graph=$work_dir/inception_quantized1.pb \
--print_structure=true
```
4. 使用权重量化的模型做推理验证
```
$ bazel build tensorflow/examples/label_image：label_image
$ bazel-bin/tensorflow/examples/label_image/label_image \
--image=$work_dir/grace_hopper.jpg \
--labels=$work_dir/imagenet_slim_labels.txt \
--graph=$work_dir/inception_quantized1.pb
```

## 五、 为什么模型量化没有使推理加速
　　关于tensorflow模型量化没有实现模型加速的，我查了一些资料，发现出现类似的问题不在少数。根据tensorflow团队成员的回复，截了几个member的答复，大意是目前量化目前针对移动端的优化，当然也有一些移动端的人说速度下降了。tensorflow未来有可能针对intel x86，gpu量化优化，但不知道什么时候支持。


　　The quantization is aimed at mobile performance, so most of the optimizations are for ARM not x86. We're hoping to get good quantization on Intel eventually, but we don't have anyone actively working on it yet.

　　Quantized ops currently only work on the CPU, because most GPUs don't support eight-bit matrix multiplications natively. I have just seen that the latest TitanX Pascal cards offer eight-bit support though, so I'm hoping we will be able to use that in the future.


参考：
1. https://zhuanlan.zhihu.com/p/33535898
2. https://arxiv.org/abs/1806.08342
3. https://github.com/google/gemmlowp/blob/master/doc/quantization.md
4. https://github.com/tensorflow/tensorflow/issues/2807


