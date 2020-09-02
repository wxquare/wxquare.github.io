---
title: 了解tensorflow不同格式的模型及其转换方法
categories:
- AI
mathjax: true
---
　　tensorflow针对训练、预测、服务端和移动端等环境支持多种模型格式，这对于初学者来说可能比较疑惑。目前，tf中主要包括.ckpt格式、.pb格式SavedModel和tflite四种格式的模型文件。SavedModel用于tensorflow serving环境中，tflite格式模型文件用在移动端，后续遇到相关格式模型文件会继续补充。这里主要介绍常见的ckpt和pb格式的模型文件，以及它们之间的转换方法。

## CheckPoint(*.ckpt)
　　在使用tensorflow训练模型时，我们常常使用tf.train.Saver类保存和还原，使用该类保存和模型格式称为checkpoint格式。Saver类的save函数将图结构和变量值存在指定路径的三个文件中，restore方法从指定路径下恢复模型。当数据量和迭代次数很多时，训练常常需要数天才能完成，为了防止中间出现异常情况，checkpoint方式能帮助保存训练中间结果，避免重头开始训练的尴尬局面。有些地方说ckpt文件不包括图结构不能重建图是不对的，使用saver类可以保存模型中的全部信息。尽管ckpt模型格式对于训练时非常方便，但是对于预测却不是很好，主要有下面这几个缺点：
1. ckpt格式的模型文件依赖于tensorflow，只能在该框架下使用;
2. ckpt模型文件保存了模型的全部信息，但是在使用模型预测时，有些信息可能是不需要的。模型预测时，只需要模型的结构和参数变量的取值，因为预测和训练不同，预测不需要变量初始化、反向传播或者模型保存等辅助节点;
3. ckpt将模型的变量值和计算图分开存储，变量值存在index和data文件中，计算图信息存储在meta文件中,这给模型存储会有一定的不方便。

## frozen model(*.pb)
　　Google推荐将模型保存为pb格式。PB文件本身就具有语言独立性，而且能被其它语言和深度学习框架读取和继续训练，所以PB文件是最佳的格式选择。另外相比ckpt格式的文件，pb格式可以去掉与预测无关的节点，单个模型文件也方便部署，因此实践中我们常常使用pb格式的模型文件。那么如何将ckpt格式的模型文件转化为pb的格式文件呢？主要包含下面几个步骤，结合这几个步骤写了个通用的脚本，使用该脚本只需指定ckpt模型路径、pb模型路径和模型的输出节点，多个输出节点时使用逗号隔开。

- 通过传入的ckpt模型的路径得到模型的图和变量数据
- 通过 import_meta_graph 导入模型中的图
- 通过 saver.restore 从模型中恢复图中各个变量的数据
- 通过 graph_util.convert_variables_to_constants 将模型持久化
- 在frozen model的时候可以删除训练节点

```
# -*-coding: utf-8 -*-
import tensorflow as tf
from tensorflow.python.framework import graph_util
import argparse


def freeze_graph(input_checkpoint,output_pb_path,output_nodes_name):
    '''
    :param input_checkpoint:
    :param output_pb_path: PB模型保存路径
    '''
    saver = tf.train.import_meta_graph(input_checkpoint + '.meta', clear_devices=True)
    with tf.Session() as sess:
        saver.restore(sess, input_checkpoint) #恢复图并得到数据
        graph = tf.get_default_graph()
        # 模型持久化，将变量值固定
        output_graph_def = graph_util.convert_variables_to_constants(  
            sess=sess,
            input_graph_def=sess.graph_def,
            output_node_names=output_nodes_name.split(","))# 如果有多个输出节点，以逗号隔开

        print("++++++++++++++%d ops in the freeze graph." % len(output_graph_def.node)) #得到当前图有几个操作节点
        output_graph_def = graph_util.remove_training_nodes(output_graph_def)
        print("++++++++++++++%d ops after remove training nodes." % len(output_graph_def.node)) #得到当前图有几个操作节点

        # serialize and write pb model to Specified path
        with tf.gfile.GFile(output_pb_path, "wb") as f: 
            f.write(output_graph_def.SerializeToString()) 

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--ckpt_path', type=str, required=True,help='checkpoint file path')
    parser.add_argument('--pb_path', type=str, required=True,help='pb model file path')
    parser.add_argument('--output_nodes_name', type=str, required=True,help='name of output nodes separated by comma')

    args = parser.parse_args()
    freeze_graph(args.ckpt_path,args.pb_path,args.output_nodes_name)

```


参考：
https://blog.metaflow.fr/tensorflow-how-to-freeze-a-model-and-serve-it-with-a-python-api-d4f3596b3adc