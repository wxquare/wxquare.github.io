---
title: 了解模板追踪算法和高斯牛顿迭代法
categories:
- other
mathjax: true
---
　　最近在项目中使用到visp库的模板追踪算法(template tracker)，由于接触算法的时间比较短，这里简单记录对算法的理解和认识。模板追踪算法原理比较简单，当代价函数为SSD时，抽象为数学中的非线性最优化问题，这里采用高斯牛顿法求解。高斯牛顿法应该是通用的一种求最优化问题的算法，高斯牛顿法核心是迭代公式，不断迭代更新出新的参数值。visp模板算法效率本身不高，因此在实现的时候提供了一些可调的优化的参数，例如金字塔、采样率、迭代次数、误差等。在项目中，visp模板追踪算法在参考模板没有遮挡的情况下，效果基本满足要求，但是在有遮挡的情况，会存在比较大的问题，因此我们针对遮挡情况，进行了特别的优化。除此之外，我们优化了一个并行版本的模板追踪算法，提升追踪效率。


## 概述
　　在了解visp模板追踪算法之前，可通过官网上的[视频](https://visp.inria.fr/template-tracking/)了解追踪算法的能力。它和kcf之类的追踪算法还不太一样，在kcf追踪算法中，我们需要告诉追踪器的追踪目标，通常情况下，我们不要求像素级别的进度的要求。而template tracker参考模板（reference template）计算视频中两帧之间的单应矩阵Homography，通过单应矩阵计算目标区域在当前帧的位置，从而实现追踪的效果。

## 数学描述
　　visp库中为模板追踪算法提供了SSD、ZNNC和在VISP 3.0.0时引入的MI(mutual information) 代价函数。这里以SSD代价函数描述模板追踪算法。模板追踪算法在数学描述为一个最优化问题，通过SSD代价函数，缩小误差，寻找最优的标记帧到当前帧的单应矩阵。模板追踪算法的数学描述如下：
$$ H\_t = \arg \min \limits\_{H}\sum\_{x∈ROI}((I^*(x)-I\_t(w(x,H)))^2 $$
- $I^*$表示标记帧(参考帧），$I_t$表示当前帧
- ROI表示参考区域（参考模板，reference template)
- $H$ 表示参考帧$I^*$到当前帧的的单应矩阵Homography
- $x$ 表示图像中的一个像素点
- $w(x,H)$ 表示标记帧上像素点$x$根据单应矩阵$H$到当前帧的映射关系

　　这里使用经典的**高斯牛顿法(Gauss–Newton algorithm)迭代法**求解，关于高斯牛顿法这里就不赘述了，最关键的是其迭代公式，感兴趣可以参考下面两篇文章：
- https://en.wikipedia.org/wiki/Gauss%E2%80%93Newton_algorithm
- https://zhuanlan.zhihu.com/p/42383070

　　其迭代公式如下，$J$表示雅克比矩阵，$J^T$表示$J$的转置，$H_t$表示迭代的初始值，$H_k$表示上一次迭代的结果，$r(H_k)$表示上一次迭代的残差residual。

$$ H_{t+1} = H_t + (J^TJ)^{-1}J^Tr(H_k)  $$


## 关键实现步骤
　　了解了模板追踪算法的数学描述和高斯牛顿迭代算法，其基本实现应该是不难的，它本质上是一个迭代算法主要分为以下几步：
step1. 设定初始的$H$矩阵，第一帧为单一矩阵，之后上一帧的结果.
step2. 对于第$k$次迭代计算雅克比$J$, 残差$r(H_k)$，得到$\triangledown H=-(J^TJ)^{-1}J^Tr(H_k)$.
step3. 如果$\triangledown H$ 足够小或者达到最大循环次数就停止迭代  
step4. 如果不满足迭代停止条件$H_{k+1}=H_{k} +\triangledown H$ 
step5. 迭代结束时，$H_{t+1}=H_{k}$

### 1.计算关键帧中的参考区域中(reference template）中每个像素点的雅克比:
- 计算关于x方向的梯度
- 计算关于y方向的梯度
- 对ROI中的每个点uv计算$J=[d_xu,d_xv,d_x,d_yu,d_yv,d_y,-d_xu^2-d_yuv,-d_xuv-d_yv^2]$
```
	# img0 表示标记帧
    dx = cv2.Sobel(img0, cv2.CV_64F, 1, 0, ksize)
    dy = cv2.Sobel(img0, cv2.CV_64F, 0, 1, ksize)
    img0 = cv2.GaussianBlur(img0, (ksize, ksize), 1)
	
	# uv表示标记帧参考区域的每个像素点
    juv = [dx[uv] * u, dx[uv] * v, dx[uv], dy[uv] * u, dy[uv] * v, dx[uv],
           -dx[uv] * u * u - dy[uv] * u * v, -dx[uv] * u * v - dy[uv] * v * v]
	J = np.array(juv).T

	# MJ=-(JT*J)^-1 *JT
    MJ = -np.dot(np.linalg.pinv(np.dot(J.T, J)), J.T)
```
### 2.迭代计算当前帧的H的矩阵
- 迭代条件停止的条件，两次迭代误差小于一个指定值，例如$10^{-8}$
- 第一次为单位矩阵，之后为上一帧的追踪结果
- 根据H矩阵将关键帧上上参考区域的点映射到当前帧: uv1 = np.dot(H, uv)
- 计算关键帧上参考区域到当前帧的误差e：E = img0[uv] - img1[uv1] 
- 计算$\triangledown H = -(J^TJ)^{-1}J_ne_n$
- 计算新的$H$
```
        # for deltaH
        MJ = -np.dot(np.linalg.pinv(np.dot(J.T, J)*lambdaJTJ), J2.T)
        #MJ = -np.dot(np.linalg.pinv(np.dot(J2.T, J2)*lambdaJTJ), J2.T)
        deltaH =alpha* np.dot(MJ, E2)

        # for newH
        dh = np.insert(deltaH, 8, np.zeros(1), 0).reshape((3, 3))
        dhinv = np.linalg.pinv(np.identity(3) + dh)
        newH = np.dot(H, dhinv)
```

## 实际实现考虑点及其存在的问题
为提高模板追踪算法的效率，visp库在实现模板追踪算法的时候设置了一些可调的参数：
- 对参考模板中的像素点进行采样处理setSampling
- 迭代时设置学习率，setLambda默认为0.001
- 设置最大迭代次数，setIterationMax(200)
- 设置金字塔的层数，tracker.setPyramidal(2, 1)

　　实际使用visp模板追踪算法中，发现当参考模板处有物体遮挡时，效果不好，因此需要做进一步的处理。另外，我们在工程实践时，为了提高追踪的效率，升级了一个并行版本的追踪，能提高数倍的追踪效率。

参考链接：
- https://visp.inria.fr/template-tracking/
- https://visp-doc.inria.fr/doxygen/visp-daily/tutorial-tracking-tt.html
- https://en.wikipedia.org/wiki/Gauss%E2%80%93Newton_algorithm
- https://zhuanlan.zhihu.com/p/42383070