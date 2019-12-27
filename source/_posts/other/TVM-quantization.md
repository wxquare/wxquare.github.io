---
title: TVM学习笔记--量化(quantization)
categories:
- other
mathjax: true
---

## 什么量化
结合模拟信号AD转化说明

## 矩阵乘法和量化


## 神经网络量化，scale和零值


## 神经网络为什么要量化


## 深度学习怎么量化


## TVM量化



1. 模型量化加速的 需要什么硬件和推理平台的支持？？？？8bit
2. GEMMLOWP
3. symmetric and asymmetric quantization
4. uniform quantization
5.  


参考：
【RFC】https://github.com/apache/incubator-tvm/issues/2259
https://github.com/apache/incubator-tvm/pull/2116


一、学习模型量化相关知识
【1】https://jackwish.net/2019/neural-network-quantization-introduction-chn.html
【2】Tensor量化和calibrate：8-bit-inference-with-tensorrt
【3】浮点数在计算机中是怎么表示的？？？以及理解浮点数和定点数
http://www.ruanyifeng.com/blog/2010/06/ieee_floating-point_representation.html
【4】Quantizing deep convolutional networks for efficient inference: A whitepaper：https://arxiv.org/pdf/1806.08342.pdf

二、TVM量化相关知识
【INT8 quantization proposal】：https://discuss.tvm.ai/t/int8-quantization-proposal/516（2018.02.02）
【TVM quantizationRFC】 https://github.com/apache/incubator-tvm/issues/2259(2018.12.09)
【MKL-DNN】https://software.intel.com/en-us/articles/accelerate-lower-numerical-precision-inference-with-intel-deep-learning-boost

三、graph optimation.

```
def prerequisite_optimize(graph, params=None):
    """ Prerequisite optimization passes for quantization. Perform
    "SimplifyInference", "FoldScaleAxis", "FoldConstant", and
    "CanonicalizeOps" optimization before quantization. """
    optimize = _transform.Sequential([_transform.SimplifyInference(),
                                      _transform.FoldConstant(),
                                      _transform.FoldScaleAxis(),
                                      _transform.CanonicalizeOps(),
                                      _transform.FoldConstant()])

    if params:
        graph = _bind_params(graph, params)

    mod = _module.Module.from_expr(graph)
    with _transform.PassContext(opt_level=3):
        mod = optimize(mod)
    return mod["main"]

```


四、replay学习
https://docs.tvm.ai/langref/index.html



五、TVM unpack batch normalization by default.
https://discuss.tvm.ai/t/if-i-dont-want-to-unpack-batch-normalization/155