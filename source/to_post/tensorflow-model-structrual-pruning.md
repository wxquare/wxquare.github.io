
## 运行官方的demo
(base) terse@ubuntu:~/code/python/PocketFlow$ ./scripts/run_local.sh nets/resnet_at_cifar10_run.py --learner=channel


理解背后的原理


2019-09-26 20:22:17 (3.49 MB/s) - ‘models_resnet_20_at_cifar_10.tar.gz.3’ saved [12132]

models/
models/model.ckpt-97656.index
models/checkpoint
models/model.ckpt-97656.meta
models/model.ckpt-97656.data-00000-of-00001
INFO:tensorflow:Restoring parameters from ./models/model.ckpt-97656
INFO:tensorflow:model restored from ./models/model.ckpt-97656
INFO:tensorflow:loss: 0.41850602626800537
INFO:tensorflow:accuracy: 0.9079999327659607

https://zhuanlan.zhihu.com/p/59635378

前两个：Channel pruning for accelerating very deep neural networks - ICCV 2017
最后一个：Discrimination-aware Channel Pruning for Deep Neural Networks - NIPS 2018

INFO:tensorflow:Channel pruning the model/resnet_model/conv2d_57/Conv2D layer,       the pruning rate is 1
INFO:tensorflow:loss: 1.2905842065811157
INFO:tensorflow:accuracy: 0.7403646111488342
INFO:tensorflow:Pruning accuracy 0.7403646111488342
INFO:tensorflow:The current model flops is 97141321.10468754
INFO:tensorflow:Pruned flops 97141321.10468754
INFO:tensorflow:The accuracy is 0.7403646111488342 and the flops after pruning is 97141321.10468754
INFO:tensorflow:The speedup ratio is 0.3854528482613462
INFO:tensorflow:The original model flops is 252018688.0
INFO:tensorflow:The pruned flops is 97141321.10468754
