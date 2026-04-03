# AI目录专属配置

## AI文章特殊规范
- 深度学习相关文章归入子目录：
  - `computer-vision/` - 计算机视觉相关
  - `tensorflow/` - TensorFlow框架相关
  - `tvm/` - TVM编译器相关
- 系统设计类AI文章（如22-ai-system-design.md）放在AI根目录
- Agent设计类文章（如23-dod-agent-design.md）放在AI根目录
- 文章命名：重要文章用数字前缀（如22-xxx.md），技术笔记用描述性名称

## 引用规范
- 引用论文：使用标准格式 [作者, 年份]
- 引用代码：提供GitHub链接或代码仓库链接
- 引用图片：优先使用Excalidraw或Mermaid，放在source/diagrams/
- 引用外部资源：确保链接有效，添加访问日期

## 技术术语统一
- 深度学习（不用DL）
- 神经网络（不用NN）
- 机器学习（不用ML）
- 卷积神经网络（不用CNN）
- 循环神经网络（不用RNN）
- 英文术语首次出现时加中文注释

## 代码示例规范
- Python代码必须指定版本要求（如需要Python 3.8+）
- 提供完整的依赖列表
- 代码示例要可运行，避免伪代码
- 复杂算法提供时间复杂度和空间复杂度分析

## 图表规范
- 架构图使用Excalidraw
- 流程图使用Mermaid
- 数据流图使用Mermaid
- 所有图表文件命名要有意义，避免使用默认名称
