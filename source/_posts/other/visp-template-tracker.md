---
title: visp template tracker 算法概述
categories:
- other
mathjax: true
---
　　最近在项目中使用到visp库的template tracker，这里简单记录对算法的理解和认识。由于不熟悉图像视频领域专业知识，可能会出现专业术语理解错误，主要根据源代码讲解实现过程。
参考：
- https://visp.inria.fr/template-tracking/
- https://visp-doc.inria.fr/doxygen/visp-daily/tutorial-tracking-tt.html

## 概述
　　首先可通过 https://visp.inria.fr/template-tracking/ 上的视频了解模板追踪算法的能力。它和kcf之类的追踪算法还不太一样，在kcf追踪算法中，在追踪之前，我们只需要告诉追踪器的追踪目标，通常情况下，我们不要求像素级别的进度的要求。而template tracker参考模板（reference template）计算视频中两帧之间的单应矩阵Homography，通过单应矩阵计算目标区域在当前帧的位置，从而实现追踪的效果。

## 数学描述
　　visp库中为模板追踪算法提供了SSD、ZNNC和在VISP 3.0.0时引入的MI(mutual information) 代价函数。这里以SSD代价函数描述模板追踪算法。因此，模板追踪算法在数学上是一个最优化问题，通过SSD代价函数，寻找最优的标记帧到当前帧的单应矩阵。模板追踪算法的数学描述也比较简单：
$$ H\_t = \arg \min \limits\_{H}\sum\_{x∈ROI}((I^*(x)-I\_t(w(x,H)))^2 $$
- $I^*$表示标记帧(参考帧），$I_t$表示当前帧
- ROI表示参考区域（参考模板，reference template)
- $H$ 表示参考帧$I^*$到当前帧的的单应矩阵Homography
- $x$ 表示图像中的一个像素点
- $w(x,H)$ 表示标记帧上像素点$x$根据单应矩阵$H$到当前帧的映射关系

那么如何解这个数学问题呢，这里使用的是经典的**高斯牛顿法(Gauss–Newton algorithm)**:
参考：https://en.wikipedia.org/wiki/Gauss%E2%80%93Newton_algorithm
$$ H_{t+1} = H_t + (J_r^TJ_r)^{-1}J_r^Tr(H_t)  $$


## 关键实现介绍
### 1.计算关键帧中的参考区域中(reference template）中每个像素点的雅克比矩阵：
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
- 计算$△H = -(J^TJ)^{-1}J_ne_n$
- 计算新的H
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

## 存在问题及其优化
- 效果，原生的模板追踪算法不能解决遮挡问题，实际中可以通过模板去掉ROI中的部分点解决遮挡问题
- 效率，原生的算法效率比较低，可通过并行策略提高效率，实际中很容易实现数倍的提升。