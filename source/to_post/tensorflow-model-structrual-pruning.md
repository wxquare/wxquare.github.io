
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