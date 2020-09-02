---
title: DaSiamRPN pytorch转tflite模型
---

　　DaSiamRPN是2018年在追踪领域的深度学习模型，公开测试结果都非常好，可以说是目前最好的追踪算法，公开的部分代码是使用Pytorch的预测部分，还没有公开训练部分代码。我们需求是将其移植到终端，初步方案是通过将其转为tensorfl lite，然后在移动端部署。由于接触深度学习时间短，对pytorch和tensorflow框架不熟悉，加上DaSiamRPN模型本身比较复杂，在模型转换过程中遇到不少的问题，目前我们已经完成了DaSiamRPN tflite 的python版本。


## 一、主流模型转换方法
两种模型转换的方法：
1. pytorch转keras，然后转tensorflow lite  
https://heartbeat.fritz.ai/deploying-pytorch-and-keras-models-to-android-with-tensorflow-mobile-a16a1fb83f2  
2. onnx，pytorch转onnx，onnx转tensorflow，tensorflow转tensorflow lite。
https://github.com/onnx/tutorials/blob/master/tutorials/PytorchTensorflowMnist.ipynb  
最开始是用第一种方法的，但是转出来的结果不对，失败原因可能是因为自己定义相同kera参数有些不同导致失败，最终放弃。最终，通过第二种方法onnx做中转，验证可行。


## 二、pytorch转tensorflow
用onnx做模型转换需要注意下面三点：  
1. forward函数决定从原模型导出模型数据的哪些部分。DaSiamRPN模型比较复杂，通过定义不同的forward我们导出了三个tflite模型。  
2. 确定导出模型的输入格式  
3. 导出的tensorflow模型文件是frozen graph格式
```     
from os.path import realpath, dirname, join
import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.autograd import Variable
import tensorflow as tf
import onnx
from onnx_tf.backend import prepare

class SiamRPN(nn.Module):             
    def __init__(self, size=2, feature_out=512, anchor=5):
        configs = [3, 96, 256, 384, 384, 256]
        configs = list(map(lambda x: 3 if x==3 else x*size, configs))
        feat_in = configs[-1]
        super(SiamRPN, self).__init__()
        
        self.featureExtract = nn.Sequential(
            nn.Conv2d(configs[0], configs[1] , kernel_size=11, stride=2),
            nn.BatchNorm2d(configs[1]),
            nn.MaxPool2d(kernel_size=3, stride=2),
            nn.ReLU(inplace=True),

            nn.Conv2d(configs[1], configs[2], kernel_size=5),
            nn.BatchNorm2d(configs[2]),
            nn.MaxPool2d(kernel_size=3, stride=2),
            nn.ReLU(inplace=True),

            nn.Conv2d(configs[2], configs[3], kernel_size=3),
            nn.BatchNorm2d(configs[3]),
            nn.ReLU(inplace=True),

            nn.Conv2d(configs[3], configs[4], kernel_size=3),
            nn.BatchNorm2d(configs[4]),
            nn.ReLU(inplace=True),
            
            nn.Conv2d(configs[4], configs[5], kernel_size=3),
            nn.BatchNorm2d(configs[5]),
        )
        
        self.anchor = anchor
        self.feature_out = feature_out
        self.conv_r1 = nn.Conv2d(feat_in, feature_out*4*anchor, 3)
        self.conv_r2 = nn.Conv2d(feat_in, feature_out, 3)
        self.conv_cls1 = nn.Conv2d(feat_in, feature_out*2*anchor, 3)
        self.conv_cls2 = nn.Conv2d(feat_in, feature_out, 3)
        self.regress_adjust = nn.Conv2d(4*anchor, 4*anchor, 1)
        self.r1_kernel = []
        self.cls1_kernel = []
        self.cfg = {}
    # pytorch forward
    def forward(self, x):
        x_f = self.featureExtract(x)
        r1_kernel_raw = self.conv_r1(x_f)
        cls1_kernel_raw = self.conv_cls1(x_f)
        return r1_kernel_raw,cls1_kernel_raw
  
class SiamRPNvot(SiamRPN):
    def __init__(self):
        super(SiamRPNvot, self).__init__(size=1, feature_out=256)
        self.cfg = {'lr':0.45, 'window_influence': 0.44, 'penalty_k': 0.04, 'instance_size': 271, 'adaptive': False} # 0.355

# load net
net = SiamRPNvot()
net.load_state_dict(torch.load(join(realpath(dirname(__file__)), 'SiamRPNVOT.model')))

# export temple_model.onnx and temple_model_pb
dummy_input = Variable(torch.randn(1, 3, 127, 127)) # one black and white 28 x 28 picture will be the input to the model
torch.onnx.export(net, dummy_input, "temple_model.onnx")
onnx_model = onnx.load("temple_model.onnx")
tf_rep = prepare(onnx_model)
print(tf_rep.inputs,tf_rep.outputs)
print(tf_rep.tensor_dict)
tf_rep.export_graph("temple_model.pb")
outputs = [node for node in tf_rep.outputs]
print(outputs)

#命令行将tensorflow模型转tensorflow lite模型
#toco --output_file=temple_model.tflite   --graph_def_file=temple_model.pb  --input_arrays=0  --output_arrays=transpose_21,transpose_24
```

## 三、tensorflow转tensorflow lite
理论上tenforflow转tensorflow lite是简单的，实践中也遇到一些问题  
1. tensorflow官网针对不同的场景提供四种tensorflow转tensorflow lite的方法，onnx导出的模型文件格式是frozen_graph。  
https://www.tensorflow.org/api_docs/python/tf/lite/TFLiteConverter  
2. frozen_graph文件转tensorflow lite需要提供input_arrays和output_arrays参数，这里是试出来的。在导出模型的时候打印tf_rep.inputs,tf_rep.outputs，tf_rep.tensor_dict，通过这里得出input_array是0,output_arrays是62和63，但是要写对应的名字，transpose_21，transpose_24。  
我是使用toco命令行进行模型转换：  
toco --output_file=temple_model.tflite   --graph_def_file=temple_model.pb  --input_arrays=0  --output_arrays=transpose_21,transpose_24


## 四、测试使用tflite模型文件
　经过上面的几个步骤就可以得到tflite模型文件，在移植到终端之前，我们先写了一些python代码进行正确性的验证。
```
temple_interpreter = tf.contrib.lite.Interpreter(model_path="temple_model.tflite")
temple_interpreter.allocate_tensors()
input_details = self.temple_interpreter.get_input_details()
output_details = self.temple_interpreter.get_output_details()
input_shape = input_details[0]['shape']
input_data = x.data.numpy() #输出数据的numpy格式
self.temple_interpreter.set_tensor(input_details[0]['index'], input_data)
self.temple_interpreter.invoke()
y1 = temple_interpreter.get_tensor(output_details[0]['index'])
y2 = temple_interpreter.get_tensor(output_details[1]['index'])
```

 通过分析输入x和输出y1和y2即可判断tflite模型文件的正确性


以上记录刚接触深度学习模型和框架时遇到的坑！

