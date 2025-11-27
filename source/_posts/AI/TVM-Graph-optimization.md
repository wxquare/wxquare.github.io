---
title: TVM学习笔记--了解Relay和图优化
date: 2020-08-13
categories:
- AI
mathjax: true
---


　　TVM主要包括两个部分，一个是Relay和图优化(graph-level)，另一个就是算子（operator）级别优化，这里简单写最近了解到的关于relay和图优化方面的东西。我们都知道深度学习网络通常都是通过计算图来描述的，计算图中的节点表示各种同的算子(opertor),边表示算子之间的依赖关系。Relay可以理解为一种可以描述深度学习网络的函数式编程语言，通过relay可以描述复杂的深度网络，文中提到了control flow。最近一段时间的时间学习直观的感受的Relay编写网络模型和其它框架没什么太多的区别，但是提供的文本形式的中间表示，对开发和调试有很大的帮助。另外，它提供了许多用于图优化的pass，供大家学习和参考。测试代码都在0.6版本上调试通过。
	代码地址：https://github.com/wxquare/programming/tree/master/blog/TVM_graph_optimization


 ## 一、Hello Relay
既然Relay是一种可以描述计算的函数式语言，逛社区的发现一段代码，可以当作Relay的第一个程序。 
API参考:https://docs.tvm.ai/api/python/relay/index.html

    ```
    from tvm import relay
    import tvm.relay.op

    x = relay.expr.var('x', relay.scalar_type('int64'), dtype = 'int64')
    one = relay.expr.const(1, dtype = 'int64')
    add = relay.op.tensor.add(x, one)    
    func = relay.expr.Function([x], add, relay.scalar_type('int64'))

    mod = relay.Module.from_expr(func)  # note this API
    print("Relay module function:\n", mod.astext(show_meta_data=False))
    graph, lib, params = tvm.relay.build(mod, 'llvm', params={})
    print("TVM graph:\n", graph)
    print("TVM parameters:\n", params)
    print("TVM compiled target function:\n", lib.get_source())

    ```
## 二、使用Relay定义卷积单元
在学习Relay的时候参考了https://zhuanlan.zhihu.com/p/91283238 这篇文章。但是可能因为版本的问题，很多API多不兼容了，因此修改了一些地方，建议读者也可以去看一下。
```
    import tvm
    from tvm.relay import transform
    import tvm.relay as relay
    import numpy as np
    from tvm.contrib import graph_runtime


    def batch_norm_infer(data,
                        gamma=None,
                        beta=None,
                        moving_mean=None,
                        moving_var=None,
                        **kwargs):
        name = kwargs.get("name")
        kwargs.pop("name")
        if not gamma:
            gamma = relay.var(name + "_gamma")
        if not beta:
            beta = relay.var(name + "_beta")
        if not moving_mean:
            moving_mean = relay.var(name + "_moving_mean")
        if not moving_var:
            moving_var = relay.var(name + "_moving_var")
        return relay.nn.batch_norm(data,
                                gamma=gamma,
                                beta=beta,
                                moving_mean=moving_mean,
                                moving_var=moving_var,
                                **kwargs)[0]

    def conv2d(data, weight=None, **kwargs):
        name = kwargs.get("name")
        kwargs.pop("name")
        if not weight:
            weight = relay.var(name + "_weight")
        return relay.nn.conv2d(data, weight, **kwargs)


    def conv_block(data, name, channels, kernel_size=(3, 3), strides=(1, 1),
                padding=(1, 1), epsilon=1e-5):
        conv = conv2d(
            data=data,
            channels=channels,
            kernel_size=kernel_size,
            strides=strides,
            padding=padding,
            data_layout='NCHW',
            name=name+'_conv')
        bn = batch_norm_infer(data=conv, epsilon=epsilon, name=name + '_bn')
        act = relay.nn.relu(data=bn)
        return act


    data_shape = (1, 3, 224, 224)
    kernel_shape = (32, 3, 3, 3)
    dtype = "float32"
    data = relay.var("data", shape=data_shape, dtype=dtype)
    act = conv_block(data, "graph", 32, strides=(2, 2))
    func = relay.Function(relay.analysis.free_vars(act),act)


    mod = relay.Module.from_expr(func)
    mod = relay.transform.InferType()(mod)
    shape_dict = {
        v.name_hint : v.checked_type for v in mod["main"].params}
    np.random.seed(0)
    params = {}
    for k, v in shape_dict.items():
        if k == "data":
            continue
        init_value = np.random.uniform(-1, 1, v.concrete_shape).astype(v.dtype)
        params[k] = tvm.nd.array(init_value, ctx=tvm.cpu(0))

    target = "llvm"
    ctx = tvm.context(target, 0)
    print("Relay module function:\n", mod.astext(show_meta_data=False))
    print("TVM parameters:\n", params.keys())

    with relay.build_config(opt_level=3):
        graph, lib, params = relay.build(mod, target, params=params)

    print("TVM graph:\n", graph)
    print("TVM parameters:\n", params.keys())
    # print("TVM compiled target function:\n", lib.get_source())
    module = graph_runtime.create(graph, lib, ctx)
    data_tvm = tvm.nd.array((np.random.uniform(-1, 1, size=data_shape)).astype(dtype))
    module.set_input('data', data_tvm)
    module.set_input(**params)
    module.run()
    output = module.get_output(0)

```
## 三、Relay Graph Optimization
前面两个例子介绍了怎么使用relay构建网络，这个部分介绍怎么使用relay做图优化。上面例子代码中没有直接图优化的代码，而是包含在relay.build中。通过追踪代码，我们这部分的逻辑集中在 https://github.com/apache/incubator-tvm/blob/v0.6/src/relay/backend/build_module.cc 这个文件的optimize函数中。这里罗列了代码用到的pass，relay提供了方便的的文本形式中间描述，感兴趣的可以自己试一下每个pass之后，发生了哪些变化。

- relay::qnn::transform::Legalize())，这个pass和qnn有关
- transform::Legalize()，我理解的这个是和目标有关的优化，一个表达式虽然在语义上等效于另一个，但可以在目标上具有更好的性能。这个在需要在异构环境下生效。
- transform::SimplifyInference() 。
简化推理阶段的数据流图。在语义上等于输入表达式的简化表达式将被返回。例如将BatchNorm展开以及去掉 dropout。
- transform::EliminateCommonSubexpr(fskip))，去除公共子表达式。
- transform::CombineParallelConv2D(3)，将多个conv2d运算符合并为一个，这部分优化会将具有相同输入的卷积合并成一个大的卷积运算。
- transform::CombineParallelDense(3))，将多个dense运算符组合为一个
- transform::FoldConstant()，常量传播优化。
- transform::FoldScaleAxis()
- transform::CanonicalizeCast()，
将特殊运算符规范化为基本运算符。这样可以简化后续分析，例如将bias_add扩展为expand_dims和broadcast_add
- transform::CanonicalizeOps()
- transform::AlterOpLayout()，layout 变换
- transform::FuseOps()，算子融合，根据一些规则，将expr中的运算符融合为较大的运算符。


## 四、使用Python API Relay 图优化
 TVM核心代码是采用C++编写的，但是也提供了Python接口，这方面初学者体验的使用。Relay图优化核心功能都提供了对应的API，因此可以尝试一下，非常简单。

```
def my_optimize(func,params=None):

    if params:
        graph = _bind_params(func, params)

    # https://docs.tvm.ai/api/python/relay/transform.html
    optimize = relay.transform.Sequential([relay.transform.SimplifyInference(),
                                      relay.transform.FoldConstant(),
                                      relay.transform.FoldScaleAxis(),
                                      relay.transform.CanonicalizeOps(),
                                      relay.transform.FoldConstant()])

    mod = relay.Module.from_expr(graph)
    mod = optimize(mod)
    return mod["main"]

mod['main'] = my_optimize(mod['main'], params)
print("Relay module function:\n", mod.astext(show_meta_data=False))

```
这里可以对比优化前后的IR.

```
Relay module function:
 v0.0.4
def @main(%data: Tensor[(1, 3, 224, 224), float32], %graph_conv_weight: Tensor[(32, 3, 3, 3), float32], %graph_bn_gamma: Tensor[(32), float32], %graph_bn_beta: Tensor[(32), float32], %graph_bn_moving_mean: Tensor[(32), float32], %graph_bn_moving_var: Tensor[(32), float32]) -> Tensor[(1, 32, 112, 112), float32] {
  %0 = nn.conv2d(%data, %graph_conv_weight, strides=[2, 2], padding=[1, 1], channels=32, kernel_size=[3, 3]) /* ty=Tensor[(1, 32, 112, 112), float32] */;
  %1 = nn.batch_norm(%0, %graph_bn_gamma, %graph_bn_beta, %graph_bn_moving_mean, %graph_bn_moving_var) /* ty=(Tensor[(1, 32, 112, 112), float32], Tensor[(32), float32], Tensor[(32), float32]) */;
  %2 = %1.0;
  nn.relu(%2) /* ty=Tensor[(1, 32, 112, 112), float32] */
}
# =====================================
Relay module function:
 v0.0.4
def @main(%data: Tensor[(1, 3, 224, 224), float32]) -> Tensor[(1, 32, 112, 112), float32] {
  %0 = nn.conv2d(%data, meta[relay.Constant][0] /* ty=Tensor[(32, 3, 3, 3), float32] */ /* ty=Tensor[(32, 3, 3, 3), float32] */, strides=[2, 2], padding=[1, 1], channels=32, kernel_size=[3, 3]) /* ty=Tensor[(1, 32, 112, 112), float32] */;
  %1 = multiply(%0, meta[relay.Constant][1] /* ty=Tensor[(32, 1, 1), float32] */ /* ty=Tensor[(32, 1, 1), float32] */) /* ty=Tensor[(1, 32, 112, 112), float32] */;
  %2 = add(%1, meta[relay.Constant][2] /* ty=Tensor[(32, 1, 1), float32] */ /* ty=Tensor[(32, 1, 1), float32] */) /* ty=Tensor[(1, 32, 112, 112), float32] */;
  nn.relu(%2) /* ty=Tensor[(1, 32, 112, 112), float32] */
}

// meta data omitted. you can use show_meta_data=True to include meta data

```


参考与进阶学习：
[1]. https://www.zhihu.com/question/331611341/answer/875630325
[2]. https://zhuanlan.zhihu.com/p/91283238
[3]. https://docs.tvm.ai/dev/relay_intro.html
[4]. https://docs.tvm.ai/dev/relay_add_op.html
[5]. https://docs.tvm.ai/dev/relay_add_pass.html
[6]. https://arxiv.org/pdf/1810.00952.pdf


