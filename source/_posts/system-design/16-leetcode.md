---
title: LeetCode 核心题型分类与解题模式速查手册
date: 2025-06-20
categories:
- 系统设计
tags:
- 算法
- 数据结构
- LeetCode
- 面试
toc: true
---

<!-- toc -->

## 数据结构

### 1. 数组与矩阵 (核心模式归类)

#### A. 原地修改 / 数组作为哈希表 (实现 $O(1)$ 额外空间)
  - [41. 缺失的第一个正数](https://leetcode.cn/problems/first-missing-positive)【模式：原地置换；将 `nums[i]` 归位到 `nums[i]-1`】
  - [448. 找到所有数组中消失的数字](https://leetcode.cn/problems/find-all-numbers-disappeared-in-an-array)【模式：正负号标记；用负号表示“该索引对应的数值已出现”】
  - [73. 矩阵置零](https://leetcode.cn/problems/set-matrix-zeroes)【模式：边缘标记；利用首行首列记录内部状态】
  - [289. 生命游戏](https://leetcode.cn/problems/game-of-life)【模式：复合状态编码；利用位运算同时记录 (现状态, 次状态)】
  - [48. 旋转图像](https://leetcode.cn/problems/rotate-image)【模式：几何变换；核心：顺时针 90° = 转置 + 水平翻转；逆时针 90° = 转置 + 垂直翻转】

#### B. 矩阵遍历与坐标变换
  - [54. 螺旋矩阵](https://leetcode.cn/problems/spiral-matrix)【模式：四边界收缩；$(u, d, l, r)$ 指针随遍历向内挤压】
  - [36. 有效的数独](https://leetcode.cn/problems/valid-sudoku)【模式：一维化索引；利用 `(r/3)*3 + c/3` 映射九宫格】
  - [189. 轮转数组](https://leetcode.cn/problems/rotate-array)【模式：三次翻转；通过 `reverse` 实现 $O(1)$ 空间位移】
  - [31. 下一个排列](https://leetcode.cn/problems/next-permutation)【模式：标准算法；找 pivot -> 找更大数 -> 交换并反转】

#### C. 双指针、贪心与接雨水 (处理单调性或边界)
  - [11. 盛最多水的容器](https://leetcode.cn/problems/container-with-most-water)【模式：对撞指针；每次移动较短边以求更大容积】
  - [42. 接雨水](https://leetcode.cn/problems/trapping-rain-water/)【模式：双指针/单调栈；核心是“木桶原理”，高度由短板决定】
  - [407. 接雨水 II](https://leetcode.cn/problems/trapping-rain-water-ii)【模式：BFS + 优先队列；从外向内收缩 3D 边界】
  - [283. 移动零](https://leetcode.cn/problems/move-zeroes)【模式：快慢指针；一个负责遍历，一个负责存放非零值】
  - [228. 汇总区间](https://leetcode.cn/problems/summary-ranges)【模式：分组循环 / 双指针；核心：通过 `nums[j+1] != nums[j]+1` 识别连续区间断点】
  - [134. 加油站](https://leetcode.cn/problems/gas-station)【模式：贪心；记录总收益与局部余量判断起点】
  - [135. 分发糖果](https://leetcode.cn/problems/candy/)【模式：双向遍历；确保同时满足左右邻居约束】

#### D. 前缀和与子数组 (处理区间和/积)
  - [560. 和为 K 的子数组](https://leetcode.cn/problems/subarray-sum-equals-k)【模式：前缀和 + 哈希表；转化为 `count(PrefixSum - K)`】
  - [238. 除了自身以外数组的乘积](https://leetcode.cn/problems/product-of-array-except-self)【模式：前后缀分解；分别存储左积和右积】
  - [53. 最大子数组和](https://leetcode.cn/problems/maximum-subarray/)【模式：Kadane 算法；动态规划基础题】
  - [152. 乘积最大子数组](https://leetcode.cn/problems/maximum-product-subarray)【模式：维护双极值；同时保存 max 和 min 应对负负得正】

#### E. 区间处理 (排序 + 贪心)
  - [56. 合并区间](https://leetcode.cn/problems/merge-intervals)【模式：区间合并；核心：按起点排序，维护 `[L, R]`，利用 `l <= cur_right` 动态扩展右边界】
  - [57. 插入区间](https://leetcode.cn/problems/insert-interval)【模式：分类讨论；核心：将区间分为“左侧不重叠”、“中间重叠合并”、“右侧不重叠”三部分处理】
  - [452. 用最少数量的箭引爆气球](https://leetcode.cn/problems/minimum-number-of-arrows-to-burst-balloons)【模式：区间交集；核心：按终点排序，贪心选择重叠区域的边缘】
  - [435. 无重叠区间](https://leetcode.cn/problems/non-overlapping-intervals)【模式：贪心留空；核心：按终点排序，尽量保留先结束的区间，以给后续留出更多空间】
  - [646. 最长数对链](https://leetcode.cn/problems/maximum-length-of-pair-chain)【模式：贪心；核心：按第二个数排序，贪心选择结束最早的区间，同 435 题】
  - [253. 会议室 II](https://leetcode.cn/problems/meeting-rooms-ii)【模式：上下车/差分思想；核心：将起点看作 +1，终点看作 -1，求最大并发数；或利用小顶堆维护当前结束时间】
  - [228. 汇总区间](https://leetcode.cn/problems/summary-ranges)【模式：分组循环 / 双指针；核心：识别连续数字序列的断点】
  - [2580. 统计将重叠区间合并成组的方案数](https://leetcode.cn/problems/count-ways-to-group-overlapping-ranges)【模式：区间合并 + 组合数学；核心：合并后得到 m 个独立连通块，结果为 $2^m$】

#### F. 查找、排序与二分
  - [4. 寻找两个正序数组的中位数](https://leetcode.cn/problems/median-of-two-sorted-arrays)【模式：二分划分；寻找切割点平衡左右数量】
  - [268. 丢失的数字](https://leetcode.cn/problems/missing-number)【模式：异或运算；利用成对抵消性质】
  - [2028. 找出缺失的观测数据](https://leetcode.cn/problems/find-missing-observations)【模式：余数分配；平均分配的思想】
  - [88. 合并两个有序数组](https://leetcode.cn/problems/merge-sorted-array)【原地 $O(1)$ 空间合并】
  - [136/137/260. 只出现一次的数字](https://leetcode.cn/problems/single-number)【模式：位运算异或；核心在于消除成对出现的位】

#### G. 哈希计数与频率统计 (利用数组或 Map 记录状态)
  - [1. 两数之和](https://leetcode.cn/problems/two-sum)【模式：在线哈希查找；核心：在一次遍历中同时进行“查找”与“存入”，实现 $O(n)$ 时间复杂度】
  - [128. 最长连续序列](https://leetcode.cn/problems/longest-consecutive-sequence)【模式：哈希集合 + 智能起点；核心：利用 `unordered_set` 实现 $O(1)$ 查找，仅从序列起点 `(x-1 不存在)` 开始计数，确保 $O(n)$ 复杂度】
  - [217. 存在重复元素](https://leetcode.cn/problems/contains-duplicate)【模式：哈希集合；核心：利用 `unordered_set` 实现 $O(n)$ 频率检测，最基础的去重思想】
  - [219. 存在重复元素 II](https://leetcode.cn/problems/contains-duplicate-ii)【模式：固定窗口哈希；核心：维护大小为 k 的 `unordered_set`】
  - [220. 存在重复元素 III](https://leetcode.cn/problems/contains-duplicate-iii)【模式：滑动窗口 + 有序集合；核心：利用 `std::set::lower_bound` 寻找满足范围条件的元素】
  - [202. 快乐数](https://leetcode.cn/problems/happy-number)【模式：循环检测；核心：利用 `unordered_set` 记录历史值或使用“快慢指针”在 $O(1)$ 空间内检测无限循环】
  - [383. 赎金信](https://leetcode.cn/problems/ransom-note)【模式：字符计数；利用 `int[26]` 数组实现 $O(n)$ 时间 $O(1)$ 空间的高性能频率校验】
  - [242. 有效的字母异位词](https://leetcode.cn/problems/valid-anagram)【模式：频率对比；核心：利用 `int[26]` 计数，通过“先加后减”配合“负数早期退出”实现 $O(n)$ 校验】
  - [49. 字母异位词分组](https://leetcode.cn/problems/group-anagrams)【模式：等类规约；核心：利用“排序后的字符串”或“字符频次”作为 Map 的 Key 进行归一化分类】
  - [387. 字符串中的第一个唯一字符](https://leetcode.cn/problems/first-unique-character-in-a-string)【模式：两次遍历；先统计频次，再找第一个频次为 1 的索引】
  - [205. 同构字符串](https://leetcode.cn/problems/isomorphic-strings)【模式：索引映射；通过 `mapS[s[i]] == mapT[t[i]]` 校验字符映射的一致性】
  - [290. 单词规律](https://leetcode.cn/problems/word-pattern)【模式：双向哈希；核心：利用双 Map 或 Map+Set 建立 char 与 string 的双射关系，注意利用 `stringstream` 处理单词拆分】
  - [266. 判断一个字符串是否是回文排列](https://leetcode.cn/problems/palindrome-permutation)【模式：奇偶计数；回文排列最多只能有一个字符出现奇数次】
  - [409. 最长回文串](https://leetcode.cn/problems/longest-palindrome)【模式：贪心构造；统计成对出现的字符，最后可选加一个奇数项作为中心】

### 2. 栈与单调栈 (核心模式归类)

#### A. 基础栈应用与模拟 (处理嵌套、撤销与状态存取)
  - [20. 有效的括号](https://leetcode.cn/problems/valid-parentheses)【模式：括号匹配；核心：利用栈的 LIFO 特性处理嵌套关系】
  - [150. 逆波兰表达式求值](https://leetcode.cn/problems/evaluate-reverse-polish-notation)【模式：后缀表达式计算；核心：遇到运算符弹出两数计算，注意减/除顺序】
  - [224. 基本计算器](https://leetcode.cn/problems/basic-calculator)【模式：符号栈模拟；核心：利用栈维护当前括号层级的“全局正负号”，实现 $O(n)$ 一次遍历展开括号】
  - [155. 最小栈](https://leetcode.cn/problems/min-stack)【模式：双栈/辅助栈；核心：同步维护一个“当前的最小值”栈】
  - [232. 用栈实现队列](https://leetcode.cn/problems/implement-queue-using-stacks)【模式：双栈翻转；核心：利用入栈和出栈两个容器实现 FIFO】
  - [394. 字符串解码](https://leetcode.cn/problems/decode-string)【模式：多栈状态存取；核心：分别用栈存储当前的倍数 `cnt` 和已拼出的 `string`】
  - [71. 简化路径](https://leetcode.cn/problems/simplify-path)【模式：路径模拟；核心：遇到 `..` 执行出栈，配合 `stringstream` 拆分单词】

#### B. 单调栈基础 (在线性时间内寻找左右最近的极值)
  - [496. 下一个更大元素 I](https://leetcode-cn.com/problems/next-greater-element-i/)【模式：单调递减栈；核心：在栈中保留尚未找到“下一个更大数”的元素】
  - [503. 下一个更大元素 II](https://leetcode-cn.com/problems/next-greater-element-ii/)【模式：循环数组；核心：通过遍历两遍数组模拟环形结构】
  - [739. 每日温度](https://leetcode.cn/problems/daily-temperatures)【模式：距离计算；核心：栈中存储索引，用于计算下标差值】
  - [901. 股票价格跨度](https://leetcode.cn/problems/online-stock-span)【模式：在线单调栈；核心：将历史跨度累加，实现 $O(1)$ 的平均查询】

#### C. 单调栈进阶 (处理区间面积与贡献度计算)
  - [84. 柱状图中最大的矩形](https://leetcode.cn/problems/largest-rectangle-in-histogram)【模式：左右扩展边界；核心：利用单调栈一次性确定每个柱子的左、右边界，求最大矩形面积】
  - [907. 子数组的最小值之和](https://leetcode.cn/problems/sum-of-subarray-minimums)【模式：贡献度法；核心：计算每个元素作为最小值的区间覆盖范围 $(i-L)*(R-i)$】
  - [2866. 美丽塔 II](https://leetcode.cn/problems/beautiful-towers-ii)【模式：前后缀单调栈；核心：分别计算左侧和右侧的单调递增和，最后枚举顶点取 Max】
  - [768. 最多能完成排序的块 II](https://leetcode.cn/problems/max-chunks-to-make-sorted-ii/)【模式：单调栈维护块极值；核心：栈中每个元素代表一个“块”的最大值，重叠则合并】

#### D. 栈与贪心/其他
  - [769. 最多能完成排序的块](https://leetcode.cn/problems/max-chunks-to-make-sorted)【模式：贪心模拟；核心：维护当前最大值，若 `max == index` 则可分块】
  - [2216. 美化数组的最少删除数](https://leetcode.cn/problems/minimum-deletions-to-make-array-beautifu)【模式：相邻元素冲突；核心：贪心删除或栈模拟相邻不同】
  - [2818. 操作使得分最大](https://leetcode.cn/problems/apply-operations-to-maximize-score)【模式：组合应用；核心：质因数分解 + 贡献度计数 + 单调栈】

### 3. 队列、双端队列与堆 (用于滑动窗口极值或动态 TopK)
  - [295. 数据流的中位数](https://leetcode.cn/problems/find-median-from-data-stream)【最大堆 + 最小堆】
  - [239. 滑动窗口最大值](https://leetcode.cn/problems/sliding-window-maximum) 【双端单调递减队列】
  - [剑指 Offer 59 - II. 队列的最大值](https://leetcode.cn/problems/dui-lie-de-zui-da-zhi-lcof)【队列 + 双端队列 + 单调队列】
  - [862. 和至少为 K 的最短子数组](https://leetcode.cn/problems/shortest-subarray-with-sum-at-least-k)
  - [2810. 故障键盘](https://leetcode.cn/problems/faulty-keyboard/)【deque 模拟】
  - [最大平均通过率](https://leetcode-cn.com/problems/maximum-average-pass-ratio/)
  - [单线程 CPU](https://leetcode-cn.com/problems/single-threaded-cpu/)
  - [1823. 找出游戏的获胜者](https://leetcode.cn/problems/find-the-winner-of-the-circular-game)【队列模拟】

### 4. 链表 (核心模式归类)

#### A. 基础操作与反转 (双指针、递归与 Dummy Node)
  - [2. 两数相加](https://leetcode-cn.com/problems/add-two-numbers/)【模式：模拟加法；注意进位处理与 Dummy Node 的使用】
  - [206. 反转链表](https://leetcode.cn/problems/reverse-linked-list)【模式：迭代反转；核心：利用 `pre`, `cur`, `next` 三指针完成原地调向】
  - [92. 反转链表 II](https://leetcode-cn.com/problems/reverse-linked-list-ii/)【模式：局部反转；核心：头插法实现一次遍历反转】
  - [25. K 个一组翻转链表](https://leetcode.cn/problems/reverse-nodes-in-k-group)【模式：分段翻转；核心：先求长度确定组数，利用“头插法”在 $O(n)$ 时间 $O(1)$ 空间内完成翻转】
  - [82. 删除排序链表中的重复元素 II](https://leetcode.cn/problems/remove-duplicates-from-sorted-list-ii)【模式：三指针去重；核心：利用 Dummy 节点，彻底跳过重复项】
  - [1721. 交换链表中的节点](https://leetcode.cn/problems/swapping-nodes-in-a-linked-list/)【模式：快慢指针；核心：找倒数第 k 个节点与正数第 k 个节点进行交换】
  - [160. 相交链表](https://leetcode.cn/problems/intersection-of-two-linked-lists)【模式：双指针同步；核心：`a+c+b = b+c+a`，消除长度差实现首遇】

#### B. 快慢指针与环形检测
  - [141. 环形链表](https://leetcode.cn/problems/linked-list-cycle)【模式：快慢指针；核心：利用步长差 $(2-1=1)$，在 $O(n)$ 时间 $O(1)$ 空间内检测链表是否有环】
  - [142. 环形链表 II](https://leetcode.cn/problems/linked-list-cycle-ii)【模式：双指针追赶；核心：相遇后将一指针归零，同步慢走寻找环入口】
  - [19. 删除链表的倒数第 N 个结点](https://leetcode.cn/problems/remove-nth-node-from-end-of-list)【模式：快慢指针；核心：利用 $n$ 步位移差定位倒数第 $n+1$ 个节点（前驱节点）】
  - [61. 旋转链表](https://leetcode.cn/problems/rotate-list)【模式：成环解环；核心：先连成环再在 $n-(k\%n)$ 处断开，简化指针操作】
  - [876. 链表的中间结点](https://leetcode.cn/problems/middle-of-the-linked-list)【模式：快慢指针；核心：`fast` 走两步 `slow` 走一步，`fast` 到头时 `slow` 在中点】
  - [287. 寻找重复数](https://leetcode.cn/problems/find-the-duplicate-number)【模式：映射找环；将数组索引视为链表指针，转化为环入口问题】

#### C. 合并、排序与分隔
  - [21. 合并两个有序链表](https://leetcode.cn/problems/merge-two-sorted-lists)【模式：双指针归并；合并有序链表的基础】
  - [23. 合并 K 个升序链表](https://leetcode.cn/problems/merge-k-sorted-lists)【模式：分治归并；核心：利用分治思想将 K 个链表两两合并，实现 $O(N \log k)$ 复杂度】
  - [148. 链表归并排序](https://leetcode.cn/problems/sort-list)【模式：归并排序；核心：快慢指针找中点 + 递归合并】
  - [147. 对链表进行插入排序](https://leetcode.cn/problems/insertion-sort-list)【模式：插入排序；维护已排序部分，将新节点插入合适位置】
  - [86. 分隔链表](https://leetcode.cn/problems/partition-list)【模式：双链表拆分；将节点按大小分拨到两个哑节点链表，最后首尾相连】

#### D. 复杂链表与采样
  - [138. 随机链表的复制](https://leetcode.cn/problems/copy-list-with-random-pointer)【模式：原地克隆；核心：`A->A'->B->B'` 插入法，实现 $O(1)$ 空间拷贝随机指针】
  - [382. 链表随机节点](https://leetcode.cn/problems/linked-list-random-node)【模式：水塘抽样；核心：从未知长度流中等概率采样，确保概率为 $1/i$】
  - [146. LRU 缓存](https://leetcode-cn.com/problems/lru-cache/)【模式：哈希表 + 双向链表；实现 $O(1)$ 的访问与淘汰】
  - [460. LFU 缓存](https://leetcode-cn.com/problems/lfu-cache/)【模式：双哈希表 + 频次链表；实现 $O(1)$ 的频率敏感淘汰】

### 5. 二叉树与树形结构 (核心模式归类)

#### A. 遍历、属性与结构基础 (递归与迭代)
  - [144/94/145. 前/中/后序遍历](https://leetcode.cn/problems/binary-tree-inorder-traversal)【模式：栈迭代；核心：利用显式栈模拟递归过程，掌握中序“左孩子一路入栈”模板】
  - [102. 二叉树的层序遍历](https://leetcode.cn/problems/binary-tree-level-order-traversal/)【模式：BFS；核心：利用队列按层处理，`size` 控制当前层边界】
  - [103. 二叉树的锯齿形层序遍历](https://leetcode.cn/problems/binary-tree-zigzag-level-order-traversal)【模式：BFS + Deque；核心：根据层数奇偶性决定从队头或队尾插入，规避显式翻转开销】
  - [637. 二叉树的层平均值](https://leetcode.cn/problems/average-of-levels-in-binary-tree)【模式：BFS；核心：按层求和，注意利用 `long long` 防止累加溢出】
  - [199. 二叉树的右视图](https://leetcode.cn/problems/binary-tree-right-side-view)【模式：BFS / DFS；核心：BFS 记录层末节点，或 DFS 优先访问右子树并记录首次到达深度】
  - [226. 翻转二叉树](https://leetcode.cn/problems/invert-binary-tree)【模式：递归/层序；核心：交换左右子节点，自底向上或自顶向下均可】
  - [101. 对称二叉树](https://leetcode.cn/problems/symmetric-tree)【模式：分治；核心：判断 `L->left vs R->right` 且 `L->right vs R->left`】
  - [104. 二叉树的最大深度](https://leetcode.cn/problems/maximum-depth-of-binary-tree)【模式：DFS/回溯；核心：`max(left, right) + 1`】
  - [222. 完全二叉树的节点个数](https://leetcode.cn/problems/count-complete-tree-nodes)【模式：分治 + 完全二叉树性质；核心：利用左右子树高度差判断满二叉树，实现 $O(\log^2 n)$ 极速计数】

#### B. 路径、祖先与贡献度计算 (DFS 进阶)
  - [236. 二叉树的最近公共祖先](https://leetcode.cn/problems/lowest-common-ancestor-of-a-binary-tree)【模式：DFS 递归；核心：利用返回值上传状态（自底向上聚合），实现无副作用的纯函数设计】
  - [1644. LCA II](https://leetcode.cn/problems/lowest-common-ancestor-of-a-binary-tree-ii)【模式：计数 DFS；核心：需完整遍历以确认 p, q 是否均存在】
  - [112/113. 路径总和](https://leetcode.cn/problems/path-sum-ii)【模式：回溯；核心：维护当前路径和，到达叶子节点时校验】
  - [129. 求根节点到叶节点数字之和](https://leetcode.cn/problems/sum-root-to-leaf-numbers)【模式：DFS 路径累加；核心：利用 `x * 10 + node->val` 传递路径状态，推荐“结果上传”的纯函数写法】
  - [124. 二叉树中的最大路径和](https://leetcode.cn/problems/binary-tree-maximum-path-sum)【模式：树形 DP / 贡献度；核心：区分“节点作为转折点”与“节点作为贡献者”两种状态，利用 `max(0, gain)` 过滤负贡献】
  - [114. 二叉树展开为链表](https://leetcode.cn/flatten-binary-tree-to-linked-list)【模式：递归/前驱节点；核心：将左子树插入右侧，注意清空左指针以防成环】
  - [2385. 感染二叉树需要的总时间](https://leetcode.cn/problems/amount-of-time-for-binary-tree-to-be-infected)【模式：图论转化；核心：建图后 BFS 或 DFS 求最长路径】
  - [572. 另一棵树的子树](https://leetcode.cn/problems/subtree-of-another-tree)【模式：双重递归；核心：对每个节点调用 `isSameTree`】

#### C. 构造、变换与序列化
  - [105. 从前序与中序遍历序列构造二叉树](https://leetcode.cn/problems/construct-binary-tree-from-preorder-and-inorder-traversal)【模式：分治；核心：前序定根，中序分左右；优化：哈希表预处理索引实现 $O(n)$】
  - [106. 从中序与后序遍历序列构造二叉树](https://leetcode.cn/problems/construct-binary-tree-from-inorder-and-postorder-traversal)【模式：分治；核心：后序定根，中序分左右】
  - [117. 填充每个节点的下一个右侧节点指针 II](https://leetcode.cn/problems/populating-next-right-pointers-in-each-node-ii)【模式：层序遍历 / 链表模拟；核心：利用已建立的 `next` 指针作为“下一层”的驱动，实现 $O(1)$ 空间复杂度】

#### D. 二叉搜索树 (BST 专项)
  - [108. 将有序数组转换为二叉搜索树](https://leetcode.cn/problems/convert-sorted-array-to-binary-search-tree)【模式：分治/二分构造；核心：选取中间节点作为根，确保左右子树高度平衡】
  - [230. 二叉搜索树中第 K 小的元素](https://leetcode.cn/problems/kth-smallest-element-in-a-bst)【模式：中序遍历；核心：利用 BST 中序递增特性，通过迭代法实现“早期停止”以优化性能；进阶：频繁查询可维护子树 size 实现 $O(h)$ 检索】
  - [530. 二叉搜索树的最小绝对差](https://leetcode.cn/problems/minimum-absolute-difference-in-bst)【模式：中序遍历；核心：利用 BST 中序递增特性；进阶：Morris 遍历可实现 $O(1)$ 空间复杂度】
  - [98. 验证二叉搜索树](https://leetcode.cn/problems/validate-binary-search-tree)【模式：中序遍历 / 分治；陷阱：需保证子树所有节点均在全局上下界内，而非仅满足局部父子关系】
  - [173. 二叉搜索树迭代器](https://leetcode.cn/problems/binary-search-tree-iterator)【模式：栈模拟中序遍历；核心：利用显式栈实现懒加载，确保 $O(h)$ 空间与均摊 $O(1)$ 时间】
  - [LCR 193. 二叉搜索树的最近公共祖先](https://leetcode.cn/problems/er-cha-sou-suo-shu-de-zui-jin-gong-gong-zu-xian-lcof)【模式：数值比较；核心：利用 `root->val` 与 `p, q` 的大小关系快速剪枝】
  - [2673. 使二叉树所有路径值相等的最小代价](https://leetcode.cn/problems/make-costs-of-paths-equal-in-a-binary-tree)【模式：贪心/树形 DP；核心：自底向上平衡左右子树代价】

### 6. 树与图论 (Tree & Graph - 核心模式归类)

#### A. 树的基础与进阶 (Tree)
  - [3373. 连接两棵树后最大目标节点数目 II](https://leetcode.cn/problems/maximize-the-number-of-target-nodes-after-connecting-trees-ii)
  - [3786. 树组的交互代价总和](https://leetcode.cn/problems/total-sum-of-interaction-cost-in-tree-groups)【边贡献法，auto lamda dfs写法，避免使用function】

#### B. 网格搜索与连通性 (DFS/BFS)
  - [200. 岛屿数量](https://leetcode.cn/problems/number-of-islands)【模式：DFS/BFS；核心：原地修改标记（沉岛）实现 $O(1)$ 空间】
  - [305. 岛屿数量 II](https://leetcode.cn/problems/number-of-islands-ii)【模式：并查集 (Union-Find)；核心：动态维护连通分量，将"陆地化"转化为"集合合并"】
  - [130. 被围绕的区域](https://leetcode.cn/problems/surrounded-regions)【模式：逆向思维；从边界 `'O'` 开始标记，未被标记的内部 `'O'` 均需填充】
  - [133. 克隆图](https://leetcode.cn/problems/clone-graph)【模式：哈希表 + DFS/BFS；核心：利用 Map 存储 `[原节点 -> 新节点]` 防止死循环】
  - [399. 除法求值](https://leetcode.cn/problems/evaluate-division)【模式：带权图搜索；将变量视为节点，比值视为边权，通过 DFS 或并查集求解】

#### C. 拓扑排序 (有向无环图 DAG)
  - [207. 课程表](https://leetcode.cn/problems/course-schedule)【模式：入度统计；判断图中是否存在环】
  - [210. 课程表 II](https://leetcode.cn/problems/course-schedule-ii)【模式：BFS 拓扑序列；核心：记录入度为 0 的节点弹出顺序】
  - [2603. 收集树中金币](https://leetcode.cn/problems/collect-coins-in-a-tree)【模式：剥层法；通过度数反复删除叶子节点】

#### D. 广度优先搜索进阶 (最短路径/步数)
  - [909. 蛇梯棋](https://leetcode.cn/problems/snakes-and-ladders)【模式：BFS；核心：一维坐标与二维矩阵的映射转换】
  - [433. 最小基因变化](https://leetcode.cn/problems/minimum-genetic-mutation)【模式：单向/双向 BFS；寻找状态空间的最短路径】
  - [127. 单词接龙](https://leetcode.cn/problems/word-ladder)【模式：双向 BFS；核心：利用中间态（如 `h*t`）优化状态转移搜索】

#### E. 最短路径算法 (Dijkstra/Floyd/Bellman)
  - [743. 网络延迟时间](https://leetcode.cn/problems/network-delay-time)【Dijkstra 模板】
  - [2642. 设计可以求最短路径的图类](https://leetcode.cn/problems/design-graph-with-shortest-path-calculator)【Dijkstra 封装】
  - [1334. 阈值距离内邻居最少的城市](https://leetcode.cn/problems/find-the-city-with-the-smallest-number-of-neighbors-at-a-threshold-distance)【Floyd 多源最短路】
  - [2045. 到达目的地的第二短时间](https://leetcode.cn/problems/second-minimum-time-to-reach-destination)
  - [1976. 到达目的地的方案数](https://leetcode.cn/problems/number-of-ways-to-arrive-at-destination)
  - [2065. 最大化一张图中的路径价值](https://leetcode.cn/problems/maximum-path-quality-of-a-graph/description/)
  - [2959.关闭分部的可行集合数目](https://leetcode.cn/problems/number-of-possible-sets-of-closing-branche)
  - [3112. 访问消失节点的最少时间](https://leetcode.cn/problems/minimum-time-to-visit-disappearing-nodes)
  - [最短路径和K短问题汇总](https://github.com/wxquare/online-algorithm-coding/blob/master/other/graph_short_path.cc)

#### F. 高级图论与连通性 (并查集/基环树/网络流)
  - [2127. 参加会议的最多员工数](https://leetcode.cn/problems/maximum-employees-to-be-invited-to-a-meeting)【基环树内向树】
  - [2360. 图中的最长环](https://leetcode.cn/problems/longest-cycle-in-a-graph)
  - [2876. 有向图访问计数](https://leetcode.cn/problems/count-visited-nodes-in-a-directed-graph)
  - [1584. 连接所有点的最小费用](https://leetcode.cn/problems/min-cost-to-connect-all-points)【最小生成树 Kruskal】
  - [785. 判断二分图](https://leetcode.cn/problems/is-graph-bipartite)【染色法】
  - [7rLGCR. 守卫城堡](https://leetcode.cn/problems/7rLGCR/)【网络流/最小割】
  - [1489. 找到最小生成树里的关键边和伪关键边](https://leetcode.cn/problems/find-critical-and-pseudo-critical-edges-in-minimum-spanning-tree)
  - [2685. 统计完全连通分量的数量](https://leetcode.cn/problems/count-the-number-of-complete-components)
  - [查找集群内的关键连接](https://leetcode.cn/problems/critical-connections-in-a-network)
  - [6942. 树中可以形成回文的路径数](https://leetcode.cn/problems/count-paths-that-can-form-a-palindrome-in-a-tree)
  - [并查集 (Union Find) 模板与应用](https://github.com/wxquare/online-algorithm-coding/blob/master/other/union_find.cc)

### 7. 平衡二叉搜索树 (std::map/set)

### 8. 前缀和、后缀和与区间求和
  - [1423. 可获得的最大点数](https://leetcode.cn/problems/maximum-points-you-can-obtain-from-cards)
  - [2132. 用邮票贴满网格图](https://leetcode.cn/problems/stamping-the-grid/description/)(二维前缀和)
  - [1671. 得到山形数组的最少删除次数](https://leetcode.cn/problems/minimum-number-of-removals-to-make-mountain-array)
  - [3212. 统计 X 和 Y 频数相等的子矩阵数量](https://leetcode.cn/problems/count-submatrices-with-equal-frequency-of-x-and-y/description/)(二维)
  - [3152. 特殊数组 II](https://leetcode.cn/problems/special-array-ii)【prefix sum array】
  - [3127. 构造相同颜色的正方形](https://leetcode.cn/problems/make-a-square-with-the-same-color)【二维区间和】
  - [238. 除了自身以外数组的乘积](https://leetcode.cn/problems/product-of-array-except-self)【前缀乘积 后缀乘积】
  - [238. 除了自身以外数组的乘积](https://leetcode.cn/problems/product-of-array-except-self)【两次遍历优化前后缀存储空间】

### 9. 差分 (连续区间同时加上或者减去一个数，数组还原)
  - [1109. Corporate Flight Bookings](https://leetcode.cn/problems/corporate-flight-bookings)
  - [1094. 拼车](https://leetcode.cn/problems/car-pooling/)
  - [2132. 用邮票贴满网格图](https://leetcode.cn/problems/stamping-the-grid/description/)（二维差分）
  - [995. K 连续位的最小翻转次数](https://leetcode.cn/problems/minimum-number-of-k-consecutive-bit-flips)

### 10. 树状数组 (Binary Indexed Tree)
  - [307. Range Sum Query - Mutable](https://leetcode.cn/problems/range-sum-query-mutable)

### 11. 线段树 (Segment Tree - 区间修改与聚合查询)
  - 手撕线段树，支持区间update和区间聚合值（最大，最小、和）等
  - [218. 天际线问题,区间更新，单点查询](https://leetcode.cn/problems/the-skyline-problem) 【区间更新（lazy)、区间范围大且稀疏（动态开点，查询单点的值】
  - [699. 掉落的方块](https://leetcode.cn/problems/falling-squares)【查询区间最大值，区间更新（lazy），数据范围很大，需要动态开点，】
  - [2276. 统计区间中的整数数目](https://leetcode.cn/problems/count-integers-in-intervals)【区间更新(lazy),区间范围很大（动态开点）线段树】
  - [2916. 子数组不同元素数目的平方和 II](https://leetcode.cn/problems/subarrays-distinct-element-sum-of-squares-ii)【线段树】
  - [3479. 将水果装入篮子 III](https://leetcode.cn/problems/fruits-into-baskets-iii/description/)【线段树，无序数组找第一个大于某个树的index,单点更新，区间最大值】

### 12. 字典树 (Trie)
  - [208. 实现 Trie (前缀树)](https://leetcode.cn/problems/implement-trie-prefix-tree/)【模式：多叉树；用于高效字符串检索与前缀匹配】
  - [211. 添加与搜索单词](https://leetcode.cn/problems/design-add-and-search-words-data-structure)【模式：Trie + DFS；核心：利用递归处理 `.` 通配符的模糊匹配】
  - [212. 单词搜索 II](https://leetcode.cn/problems/word-search-ii)【模式：Trie + 回溯；核心：将词典建模为 Trie，在网格回溯中同步移动 Trie 指针，实现多模式高效匹配】
  - [676. 设计神奇字典](https://leetcode.cn/problems/implement-magic-dictionary)【模式：Trie + DFS；支持模糊匹配】
  - [421. 数组中两个数的最大异或值](https://leetcode.cn/problems/maximum-xor-of-two-numbers-in-an-array)【模式：0-1 Trie；利用二进制前缀树实现 $O(n \log C)$ 的异或极值查找】
  - [1707. 与数组中元素的最大异或值](https://leetcode.cn/problems/maximum-xor-with-an-element-from-array)【模式：离线查询 + 0-1 Trie】
  - [1803. 统计异或值在范围内的数对有多少](https://leetcode-cn.com/problems/count-pairs-with-xor-in-a-range/)【模式：0-1 Trie + 计数；类似于数位 DP 的统计思想】

### 13. 字符串匹配 (KMP / AC 自动机)
  - [28. 找出字符串中第一个匹配项的下标](https://leetcode.cn/problems/find-the-index-of-the-first-occurrence-in-a-string)【KMP】
  - [3006. 找出数组中的美丽下标 I](https://leetcode.cn/problems/find-beautiful-indices-in-the-given-array-i)
  - [最长合法子字符串的长度 ac 自动机](https://leetcode.cn173370/)

### 14. 并查集 (Union Find)
  - [并查集](https://github.com/wxquare/online-algorithm-coding/blob/master/other/union_find.cc)
  - [547. 省份的数量](https://leetcode-cn.com/problems/number-of-provinces/)
  - [684.冗余连接](https://leetcode.cn/problems/redundant-connection)
  - [721. 账户合并](https://leetcode.cn/problems/accounts-merge)
  - [924. 尽量减少恶意软件的传播](https://leetcode.cn/problems/minimize-malware-spread)
  - [1998. 数组的最大公因数排序](https://leetcode.cn/problems/gcd-sort-of-an-array)
  - [2709. 最大公约数遍历](https://leetcode.cn/problems/greatest-common-divisor-traversal)

### 15. 数据结构设计与实现 (Consolidated)

#### A. 基础结构实现
  - [155. 最小栈](https://leetcode.cn/problems/min-stack)【模式：双栈辅助；实现 $O(1)$ 获取最小值】
  - [232. 用栈实现队列](https://leetcode.cn/problems/implement-queue-using-stacks)【模式：双栈翻转；实现 FIFO 逻辑】
  - [剑指 Offer 59 - II. 队列的最大值](https://leetcode.cn/problems/dui-lie-de-zui-da-zhi-lcof)【模式：单调队列；实现 $O(1)$ 获取队列最大值】
  - [622. 设计循环队列](https://leetcode.cn/problems/design-circular-queue)【模式：数组模拟；注意首尾指针的取模处理】
  - [641. 设计循环双端队列](https://leetcode.cn/problems/design-circular-deque)【模式：数组模拟；支持两端插入与删除】

#### B. 缓存与高级哈希
  - [146. LRU 缓存](https://leetcode-cn.com/problems/lru-cache/)【模式：哈希表 + 双向链表；实现 $O(1)$ 的访问与淘汰】
  - [460. LFU 缓存](https://leetcode-cn.com/problems/lfu-cache/)【模式：双哈希表 + 频次链表；实现 $O(1)$ 的频率敏感淘汰】
  - [705. 设计哈希集合](https://leetcode-cn.com/problems/design-hashset/)【模式：链地址法；处理哈希冲突的基础实现】
  - [706. 设计哈希映射](https://leetcode-cn.com/problems/design-hashmap/)【模式：链地址法；KV 存储的基础实现】
  - [380. O(1) 时间插入、删除和获取随机元素](https://leetcode.cn/problems/insert-delete-getrandom-o1)【模式：哈希表 + 动态数组；利用数组末尾交换实现 $O(1)$ 删除】

#### C. 树与图的高级结构
  - [208. 实现 Trie (前缀树)](https://leetcode.cn/problems/implement-trie-prefix-tree/)【模式：多叉树；用于高效字符串检索与前缀匹配】
  - [211. 添加与搜索单词](https://leetcode.cn/problems/design-add-and-search-words-data-structure)【模式：Trie + DFS；核心：利用递归处理 `.` 通配符的模糊匹配】
  - [676. 设计神奇字典](https://leetcode.cn/problems/implement-magic-dictionary)【模式：Trie + DFS；支持模糊匹配】
  - [并查集 (Union Find)](https://github.com/wxquare/online-algorithm-coding/blob/master/other/union_find.cc)【模式：路径压缩 + 按秩合并；处理连通性问题】
  - [手撕线段树 (Segment Tree)](https://leetcode.cn/problems/range-sum-query-mutable)【模式：二叉树递归；支持区间修改与 $O(\log n)$ 聚合查询】
  - [树状数组 (Fenwick Tree)](https://leetcode.cn/problems/range-sum-query-mutable)【模式：位运算 (lowbit)；极简实现的区间和维护】
  - [跳表 (SkipList)](https://leetcode.cn/problems/design-skiplist/)【模式：多级索引链表；Redis 核心数据结构，替代平衡树】
  - [2642. 设计可以求最短路径的图类](https://leetcode.cn/problems/design-graph-with-shortest-path-calculator)【模式：图论封装；动态增边并维护 Dijkstra 最短路】

### 16. 离线query
  - [2747. 统计没有收到请求的服务器数目](https://leetcode.cn/problems/count-zero-request-servers/)
  - [1851. 包含每个查询的最小区间](https://leetcode.cn/problems/minimum-interval-to-include-each-query)
  - [2940. 找到 Alice 和 Bob 可以相遇的建筑](https://leetcode.cn/problems/find-building-where-alice-and-bob-can-meet)

### 17. 优先队列与堆 (Priority Queue & Heap)

#### A. 基础堆应用 (Top K / 动态极值 / 中位数)
  - [215. 数组中的第K个最大元素](https://leetcode.cn/problems/kth-largest-element-in-an-array)【模式：快速选择 / 小顶堆；维护大小为 K 的堆】
  - [295. 数据流的中位数](https://leetcode.cn/problems/find-median-from-data-stream)【模式：对顶堆；最大堆维护左半部，最小堆维护右半部】
  - [264. 丑数 II](https://leetcode.cn/problems/ugly-number-ii/)【模式：多路归并 / 小顶堆；每次取出最小丑数生成新丑数】
  - [面试题 17.09. 第 k 个数](https://leetcode.cn/problems/get-kth-magic-number-lcci/)【模式：优先队列；同丑数 II】
  - [2512. 奖励最顶尖的 K 名学生](https://leetcode.cn/problems/reward-top-k-students)【模式：自定义排序 / 堆；Top K 问题】
  - [23. 合并 K 个升序链表](https://leetcode.cn/problems/merge-k-sorted-lists)【模式：最小堆；维护 K 个链表头的最小值】

#### B. 反悔贪心 (Regret Greedy - 核心模式)
  - [630. 课程表 III](https://leetcode.cn/problems/course-schedule-iii)【模式：大顶堆维护耗时；遇到冲突时“反悔”替换掉耗时最长的课程】
  - [LCP 30. 魔塔游戏](https://leetcode.cn/problems/p0NxJO/)【模式：小顶堆维护负值；血量不足时“反悔”将之前扣血最多的移到最后】
  - [871. 最低加油次数](https://leetcode.cn/problems/minimum-number-of-refueling-stops)【模式：大顶堆维护油量；油不够时“反悔”在之前经过的油量最大的站加油】
  - [502. IPO](https://leetcode.cn/problems/ipo)【模式：双堆；按资本排序 + 大顶堆选利润最大的项目】

#### C. 最短路径与图搜索 (Dijkstra 及其变体)
  - [743. 网络延迟时间](https://leetcode.cn/problems/network-delay-time)【模式：Dijkstra 模板；最小堆维护 (dist, node)】
  - [407. 接雨水 II](https://leetcode.cn/problems/trapping-rain-water-ii)【模式：优先队列 BFS；从外向内收缩，维护当前边界的最小高度】
  - [2642. 设计可以求最短路径的图类](https://leetcode.cn/problems/design-graph-with-shortest-path-calculator)【模式：Dijkstra 封装】
  - [1976. 到达目的地的方案数](https://leetcode.cn/problems/number-of-ways-to-arrive-at-destination)【模式：Dijkstra + DP 计数】
  - [2045. 到达目的地的第二短时间](https://leetcode.cn/problems/second-minimum-time-to-reach-destination)【模式：BFS/Dijkstra 变体；求次短路】
  - [3112. 访问消失节点的最少时间](https://leetcode.cn/problems/minimum-time-to-visit-disappearing-nodes)【模式：Dijkstra 带时间限制】

#### D. 区间与会议室 (扫描线 / 堆优化)
  - [253. 会议室 II](https://leetcode.cn/problems/meeting-rooms-ii)【模式：小顶堆；堆顶存储最早结束的会议时间，判断是否需开新房】

## 基本算法
### 1. 滑动窗口/双指针 【left,right】
  - [392. 判断子序列](https://leetcode.cn/problems/is-subsequence)【双指针；进阶：预处理 T 建立“序列自动机” (f[m][26])，将匹配复杂度从 O(m) 降至 O(n)】
  - [167. 两数之和 II - 输入有序数组](https://leetcode.cn/problems/two-sum-ii-input-array-is-sorted)【双指针夹逼；核心：利用有序性保证单向移动，减少搜索空间】
  - [151. 反转字符串中的单词](https://leetcode.cn/problems/reverse-words-in-a-string)【双指针 + 局部翻转；核心：原地 $O(1)$ 空间实现内存重排】
  - [3. 无重复字符的最长子串](https://leetcode.cn/problems/longest-substring-without-repeating-characters)【不定长滑动窗口；核心：哈希表/数组记录索引，左边界实现“跨越式”跳转】
  - [15. 三数之和](https://leetcode.cn/problems/3sum/description)【排序 + 双指针；核心：i 去重、j&k 匹配后去重、nums[i]>0 提前终止】

  - [239. 滑动窗口最大值](https://leetcode.cn/problems/sliding-window-maximum)【单调队列；核心：维护窗口内“候选最大值”，保持队列单调递减】
  - [862. 和至少为 K 的最短子数组](https://leetcode.cn/problems/shortest-subarray-with-sum-at-least-k)【前缀和 + 单调队列；核心：解决包含负数的最短区间和问题】
  - [30. 串联所有单词的子串](https://leetcode.cn/problems/substring-with-concatenation-of-all-words)【分组滑动窗口 (起点偏移 0~L-1) + 词频统计；核心：化排列为计数，化随机步长为固定 L 步长】
  - [76. 最小覆盖子串](https://leetcode.cn/problems/minimum-window-substring)【滑动窗口 + 数组映射哈希；核心：有效计数器 cnt + 左边界极致收缩】
  - [2938. 区分黑球与白球](https://leetcode.cn/problems/separate-black-and-white-balls)
  - [统计最大元素出现至少 K 次的子数组](https://leetcode.cn/problems/count-subarrays-where-max-element-appears-at-least-k-times)
  - [1793.好子数组的最大分数](https://leetcode-cn.com/problems/maximum-score-of-a-good-subarray/)
  - [1052. 爱生气的书店老板](https://leetcode.cn/problems/grumpy-bookstore-owner)
  - [2970. 统计移除递增子数组的数目 I](https://leetcode.cn/problems/count-the-number-of-incremovable-subarrays-i)
  - [42. 接雨水](https://leetcode.cn/problems/trapping-rain-water)【双指针接雨水，木桶效应】
  - [2024. 考试的最大困扰度](https://leetcode.cn/problems/maximize-the-confusion-of-an-exam)(【不定长滑动窗口个】))
  - [统计重新排列后包含另一个字符串的子字符串数目 II](https://leetcode.cn/contest/weekly-contest-416/problems/count-substrings-that-can-be-rearranged-to-contain-a-string-ii)(【不定长滑动窗口统计 满足需求的子串的数量】)



### 2. 二分搜索：从"查找"到"答案空间"的跨越 (Binary Search)
- **一维二分 (1D Binary Search)**
  - [704. 二分查找](https://leetcode.cn/problems/binary-search/) 【模式：标准模板；左闭右闭 `[l, r]` 或左闭右开 `[l, r)`】
  - [34. 在排序数组中查找元素的第一个和最后一个位置](https://leetcode.cn/problems/find-first-and-last-position-of-element-in-sorted-array/) 【模式：边界查找；寻找第一个 `>= target` 的位置】
  - [33. 搜索旋转排序数组](https://leetcode.cn/problems/search-in-rotated-sorted-array/) 【模式：分类讨论；利用局部单调性缩小区间】
  - [81. 搜索旋转排序数组 II](https://leetcode.cn/problems/search-in-rotated-sorted-array-ii/) 【模式：去重处理；含重复元素时，通过 `l++, r--` 恢复单调性，最坏 $O(n)$】
  - [153. 寻找旋转排序数组中的最小值](https://leetcode.cn/problems/find-minimum-in-rotated-sorted-array/) 【模式：旋转点查找；比较 `mid` 与 `right` 确定最小值所在半区】
  - [162. 寻找峰值](https://leetcode.cn/problems/find-peak-element) 【模式：局部单调性二分；核心：比较 `mid` 与 `mid+1` 确定爬坡方向，在无序数组中实现 $O(\log n)$ 查找】

- **二维矩阵二分 (2D Matrix Search)**
  - [74. 搜索二维矩阵](https://leetcode.cn/problems/search-a-2d-matrix) 【模式：一维化二分；整个矩阵严格单调，映射为一维处理】
  - [240. 搜索二维矩阵 II](https://leetcode.cn/problems/search-a-2d-matrix-ii/) 【模式：Z字形查找；从右上或左下开始，利用行列独立单调性 $O(m+n)$】
  - [1901. 寻找峰值 II](https://leetcode.cn/problems/find-a-peak-element-ii) 【模式：行二分 + 行内最大值；核心：162 题的一维爬坡思想在二维矩阵的扩展，利用行最大值引导搜索方向】
  - [378. 有序矩阵中第 K 小的元素](https://leetcode.cn/problems/kth-smallest-element-in-a-sorted-matrix) 【模式：二分答案 + Z字形计数；在数值空间二分，结合矩阵单调性计数】

- **二分答案 (Binary Search on Answer)**
  - [69. x 的平方根 ](https://leetcode.cn/problems/sqrtx)【模式：二分答案；在 `[0, x]` 范围内寻找 `k^2 <= x` 的最大整数】
  - [2560. 打家劫舍 IV](https://leetcode.cn/problems/house-robber-iv)【贪心+二分答案；模式：最大值最小化】
  - [1482. 制作 m 束花所需的最少天数](https://leetcode-cn.com/problems/minimum-number-of-days-to-make-m-bouquets/)
  - [3007. 价值和小于等于 K 的最大数字](https://leetcode.cn/problems/maximum-number-that-sum-of-the-prices-is-less-than-or-equal-to-k)
  - [3134. 找出唯一性数组的中位数](https://leetcode.cn/problems/find-the-median-of-the-uniqueness-array)
  - [有界数组中指定下标处的最大值](https://leetcode-cn.com/problems/maximum-value-at-a-given-index-in-a-bounded-array/)

- **进阶划分与技巧 (Advanced Partitioning & Tactics)**
  - [4. 寻找两个正序数组的中位数](https://leetcode.cn/problems/median-of-two-sorted-arrays)【模式：二分划分；核心：寻找切割点平衡左右数量，$O(\log(\min(m,n)))$】
  - [两个有序数组中的第 K 小元素](https://leetcode.cn/problems/median-of-two-sorted-arrays/)【模式：排除法二分；核心：每次排除 $k/2$ 个元素，是 4 题的通用变体】
  - [11. 盛最多水的容器](https://leetcode.cn/problems/container-with-most-water)【双指针搜索】
  - [179. 最大数](https://leetcode.cn/problems/largest-number)【自定义排序，贪心】
  - [300. 最长递增子序列](https://leetcode.cn/problems/longest-increasing-subsequence) 【二分+贪心】

### 3. 排序算法/top k/select kth
  - [面试题 17.09. 第 k 个数](https://leetcode.cn/problems/get-kth-magic-number-lcci/) 三指针，优先队列
  - [2512.奖励最顶尖的 K 名学生](https://leetcode.cn/problems/reward-top-k-students)
  - [912. 排序数组，快速排序和归并排序](https://leetcode.cn/problems/sort-an-array)【三段式快速排序less,more】
  - [1329. 将矩阵按对角线排序](https://leetcode.cn/problems/sort-the-matrix-diagonally)
  - [179. 最大数](https://leetcode.cn/problems/largest-number)【自定义逆序排序 + 贪心】

### 4. 枚举
  - [2735. 收集巧克力](https://leetcode.cn/problems/collecting-chocolates)
  - [2555. 两个线段获得的最多奖品](https://leetcode.cn/problems/maximize-win-from-two-segments)【维护左边，枚举右边】
  - [3404. 统计特殊子序列的数目](https://leetcode.cn/problems/count-special-subsequences)

### 5. 模拟/分组/循环 (group/cycle arrray/模拟/易错)
  - [66. 加一](https://leetcode.cn/problems/plus-one)【模式：数组模拟；处理进位，注意 99...9 变为 100...0 的特殊情况】
  - [415. 字符串相加](https://leetcode.cn/problems/add-strings)
  - [1823. 找出游戏的获胜者](https://leetcode.cn/problems/find-the-winner-of-the-circular-game)【队列模拟】
  - [65. 有效数字](https://leetcode.cn/problems/valid-number)【一个字符串包含+-.e判断是否是一个有效的数字】
  - [1103. 分糖果 II](https://leetcode.cn/problems/distribute-candies-to-people)【循环队列分糖果】
  - [按照题目要求，数组会被分割成若干组，且每一组的判断/处理逻辑是一样的](https://leetcode.cn/problems/longest-even-odd-subarray-with-threshold/solutions/2528771/jiao-ni-yi-ci-xing-ba-dai-ma-xie-dui-on-zuspx/)
  - [466. 统计重复个数](https://leetcode.cn/problems/count-the-repetitions)
  - [2808. 使循环数组所有元素相等的最少秒数](https://leetcode.cn/problems/minimum-seconds-to-equalize-a-circular-array)
  - [2332. 坐上公交的最晚时间](https://leetcode.cn/problems/the-latest-time-to-catch-a-bus)【双指针-易错模拟题】
  - [2760. 最长奇偶子数组](https://leetcode.cn/problems/longest-even-odd-subarray-with-threshold)

### 6. 贪心算法 (Greedy Algorithm - 核心模式归类)

#### A. 基础贪心与排序 (利用排序消除维度影响)
  - [455. 分发饼干](https://leetcode.cn/problems/assign-cookies)【模式：双指针+排序；最基础的贪心匹配】
  - [3111. 覆盖所有点的最少矩形数目](https://leetcode.cn/problems/minimum-rectangles-to-cover-points)【模式：排序+固定步长；核心：按 X 轴排序后一次遍历覆盖】
  - [2576. 求出最多标记下标](https://leetcode.cn/problems/find-the-maximum-number-of-marked-indices)【模式：排序+双指针匹配；核心：将数组分为两半进行配对】
  - [179. 最大数](https://leetcode.cn/problems/largest-number)【模式：自定义排序；核心：通过 `a+b > b+a` 确定全局最优序】
  - [910. 最小差值 II](https://leetcode.cn/problems/smallest-range-ii)【模式：排序+分界线枚举；核心：只有在排序数组的某个点切换 +k 和 -k 才能获得最优解】
  - [3789. 采购的最小花费](https://leetcode.cn/problems/minimum-cost-to-acquire-required-items)【模式：组合定价；贪心选择最优组合】

#### B. 反悔贪心 (结合优先队列动态调整)
  - [630. 课程表 III](https://leetcode.cn/problems/course-schedule-iii)【模式：反悔贪心；核心：先按截止时间排序，若当前无法加入则替换掉之前耗时最长的课程】
  - [502. IPO](https://leetcode.cn/problems/ipo)【模式：双堆/排序+大顶堆；核心：动态选择当前资金下利润最大的项目】
  - [LCP 30. 魔塔游戏](https://leetcode.cn/problems/p0NxJO/)【模式：反悔贪心；核心：血量不足时将之前扣血最多的房间移到最后】
  - [871. 最低加油次数](https://leetcode.cn/problems/minimum-number-of-refueling-stops)【模式：反悔贪心；核心：油不够时从经过的加油站中选油最多的加】

#### C. 贡献度法与数学贪心
  - [2897. 对数组执行操作使平方和最大](https://leetcode.cn/problems/apply-operations-on-array-to-maximize-sum-of-squares)【模式：位运算贪心；核心：通过位交换让大数更大，利用平方和性质】
  - [2171. 拿出最少数目的魔法豆](https://leetcode.cn/problems/removing-minimum-number-of-magic-beans)【模式：枚举+贡献度；核心：排序后枚举最终相等的豆子数量，计算总损失】
  - [2731. 移动机器人](https://leetcode.cn/problems/movement-of-robots)【模式：碰撞无关性+贡献度；核心：碰撞视为穿透，转化为所有点对距离之和】
  - [2009. 使数组连续的最少操作数](https://leetcode.cn/problems/minimum-number-of-operations-to-make-array-continuous)【模式：去重排序+滑动窗口；转化为求窗口内最多的保留元素】

#### D. 区间处理 (排序 + 边界维护)
  - [56. 合并区间](https://leetcode.cn/problems/merge-intervals)【模式：排序+维护右边界；区间贪心基础】
  - [452. 用最少数量的箭引爆气球](https://leetcode.cn/problems/minimum-number-of-arrows-to-burst-balloons)【模式：按右端点排序；核心：尽量在重叠区域射箭】
  - [435. 无重叠区间](https://leetcode.cn/problems/non-overlapping-intervals)【模式：按右端点排序；核心：留出更多空间给后续区间】
  - [769. 最多能完成排序的块](https://leetcode.cn/problems/max-chunks-to-make-sorted)【模式：前缀最大值；核心：当 `max(arr[0...i]) == i` 时可分块】

#### E. 邻居约束与多遍遍历
  - [135. 分发糖果](https://leetcode.cn/problems/candy)【模式：双向遍历；核心：左一遍满足左邻居，右一遍满足右邻居，取 Max】
  - [1702. 修改后的最大二进制字符串](https://leetcode.cn/problems/maximum-binary-string-after-change)【模式：构造贪心；核心：统计 0 的个数和第一个 0 的位置，直接构造结果】
  - [402. 移掉 K 位数字](https://leetcode.cn/problems/remove-k-digits)【模式：单调栈+贪心；核心：尽量让高位数字更小】

#### F. 贪心 + 二分/DP 结合
  - [2560. 打家劫舍 IV](https://leetcode.cn/problems/house-robber-iv)【模式：二分答案 + 贪心 Check；核心：二分最小金额，用贪心判断是否能选出 k 间房】
  - [300. 最长递增子序列](https://leetcode.cn/problems/longest-increasing-subsequence)【模式：贪心 + 二分；核心：维护一个上升最慢的序列 `d[i]`】
  - [55. 跳跃游戏](https://leetcode.cn/problems/jump-game)【模式：维护最远可达点；贪心基础】
  - [45. 跳跃游戏 II](https://leetcode.cn/problems/jump-game-ii)【模式：维护当前步数的最远边界；BFS 思想的贪心应用】

### 7. 分治 (Divide and Conquer)
> **核心逻辑**：
> 1. **分解 (Divide)**：将原问题拆分为规模较小、相互独立的子问题（如左右子树、数组半区）。
> 2. **解决 (Conquer)**：递归解决子问题，直到触及边界。
> 3. **合并 (Combine)**：将子问题的解合并为原问题的解（如归并排序的 `merge` 或 LCA 的状态上传）。

  - [50. Pow(x, n)](https://leetcode.cn/problems/powx-n) 【模式：快速幂；将大指数拆分为一半计算，实现 $O(\log n)$】
  - [23. 合并 K 个升序链表](https://leetcode.cn/problems/merge-k-sorted-lists) 【模式：分治合并；将大任务拆分为小任务合并，优化逐一合并的线性复杂度】
  - [148. 排序链表](https://leetcode.cn/problems/sort-list) 【模式：归并排序；核心：快慢指针找中点 + 递归拆分 + 有序链表合并；注意：断开中点连接以防止死循环】
  - [108. 将有序数组转换为二叉搜索树](https://leetcode.cn/problems/convert-sorted-array-to-binary-search-tree) 【模式：二分构造；选取中点作为根，递归处理左右子区间】
  - [105. 从前序与中序遍历序列构造二叉树](https://leetcode.cn/problems/construct-binary-tree-from-preorder-and-inorder-traversal) 【模式：结构重建；利用前序定根，中序划分左右子树区间】
  - [222. 完全二叉树的节点个数](https://leetcode.cn/problems/count-complete-tree-nodes) 【模式：属性统计；利用完全二叉树性质，通过高度差判断满二叉树并递归】
  - [236. 二叉树的最近公共祖先](https://leetcode.cn/problems/lowest-common-ancestor-of-a-binary-tree) 【模式：状态上传；递归寻找左右子树中的目标，根据返回值决策】
  - [101. 对称二叉树](https://leetcode.cn/problems/symmetric-tree) 【模式：镜像分治；判断 `L->left vs R->right` 且 `L->right vs R->left`】

## 数学 (Mathematics - 核心模式归类)

### 1. 数论基础 (Number Theory)
#### A. 质数、约数与筛选法
  - [204. 计数质数](https://leetcode.cn/problems/count-primes/)【模式：埃氏筛/线性筛；高效统计 $[1, n]$ 内的质数数量】
  - [判断质数与筛选模板](https://github.com/wxquare/online-algorithm-coding/blob/master/other/primes.cc)【基础模板：$O(\sqrt{n})$ 判断与 $O(n \log \log n)$ 筛选】
  - [172. 阶乘后的零](https://leetcode.cn/problems/factorial-trailing-zeroes)【模式：因子统计；核心：统计 n! 中因子 5 的个数】
  - [263. 丑数](https://leetcode.cn/problems/ugly-number)【模式：质因数分解；判断因子是否仅包含 2, 3, 5】
  - [264. 丑数 II](https://leetcode.cn/problems/ugly-number-ii/)【模式：三指针/小顶堆；按序生成仅含特定质因子的数字】
  - [2709. 最大公约数遍历](https://leetcode.cn/problems/greatest-common-divisor-traversal)【模式：质因数分解 + 并查集；通过公约数建立连通性】

#### B. 最大公约数 (GCD) 与 最小公倍数 (LCM)
  - [GCD/LCM 模板](https://github.com/wxquare/online-algorithm-coding/blob/master/other/primes.cc)【辗转相除法：`__gcd(a, b)`】
  - [1998. 数组的最大公因数排序](https://leetcode.cn/problems/gcd-sort-of-an-array)【模式：GCD + 并查集；判断是否可以通过交换实现排序】

#### C. 数字处理与投票算法
  - [9. 回文数](https://leetcode.cn/problems/palindrome-number)【模式：数学反转；核心：反转一半数字与前半部分比较，避免溢出】
  - [169. 多数元素](https://leetcode.cn/problems/majority-element)【模式：Boyer-Moore 摩尔投票法；$O(n)$ 时间 $O(1)$ 空间找众数】
  - [229. 多数元素 II](https://leetcode.cn/problems/majority-element-ii/)【模式：进阶摩尔投票；统计出现次数超过 $n/3$ 的元素】
  - [400. 第 N 位数字](https://leetcode.cn/problems/nth-digit)【模式：数学模拟；按位数区间（个位、十位...）定位数字】
  - [343. 整数拆分](https://leetcode.cn/problems/integer-break)【模式：数学推导；核心：尽可能拆分成 3 以获得最大乘积】

### 2. 组合数学与概率 (Combinatorics & Probability)
#### A. 排列组合与大数取模
  - [排列组合模板](https://leetcode.cn/problems/count-k-subsequences-of-a-string-with-maximum-beauty)【核心公式：$C(n, k) = \frac{n!}{k!(n-k)!}$】
  - [31. 下一个排列](https://leetcode.cn/problems/next-permutation/)【模式：字典序算法；找 pivot -> 找更大数 -> 交换并反转】
  - [3405. 统计恰好有 K 个相等相邻元素的数组数目](https://leetcode.cn/problems/count-the-number-of-arrays-with-k-matching-adjacent-elements)【模式：组合数 $C(n-1, k)$】
  - [3154. 到达第 K 级台阶的方案数](https://leetcode.cn/problems/find-number-of-ways-to-reach-the-k-th-stair/)【模式：组合数学 + 容斥/动态规划】

#### B. 模运算与乘法逆元
  - **模运算性质**：`(a / b) % c = (a * b^(c-2)) % c` (当 c 为质数时，费马小定理)】
  - [3463. 判断操作后字符串中的数字是否相等 II](https://leetcode.cn/problems/check-if-digits-are-equal-in-string-after-operations-ii)【模式：Lucas 定理/组合数学取模；处理非质数模数】
  - [3790. 最小全 1 倍数](https://leetcode.cn/problems/smallest-all-ones-multiple)【模式：同余定理；`(a * 10 + 1) % k` 的迭代处理】

#### C. 随机采样 (Sampling)
  - [398. 随机数索引](https://leetcode.cn/problems/random-pick-index)【模式：水塘抽样 (Reservoir Sampling)；未知长度流的等概率采样】
  - [384. 打乱数组](https://leetcode.cn/problems/shuffle-an-array)【模式：Fisher-Yates 洗牌算法；实现公平随机排列】
  - [470. 用 Rand7() 实现 Rand10()](https://leetcode.cn/problems/implement-rand10-using-rand7)【模式：拒绝抽样 (Rejection Sampling)】

### 3. 位运算 (Bit Manipulation)
#### A. 基础技巧与 Lowbit
  - **核心性质**：`n & (n-1)` 消除最低位 1；`n & -n` 获取最低位 1 (lowbit)】
  - [191. 位 1 的个数](https://leetcode.cn/problems/number-of-1-bits)【模式：`__builtin_popcount` 或 `n & (n-1)` 迭代】
  - [190. 颠倒二进制位](https://leetcode.cn/problems/reverse-bits)【模式：位操作；逐位反转 `ans = (ans << 1) | (n & 1)` 或 分治法】
  - [231. 2 的幂](https://leetcode.cn/problems/power-of-two)【模式：`n > 0 && (n & (n-1)) == 0`】
  - [201. 数字范围按位与](https://leetcode.cn/problems/bitwise-and-of-numbers-range)【模式：公共前缀；寻找 `left` 和 `right` 的二进制公共前缀】

#### B. 异或 (XOR) 专项
  - [136. 只出现一次的数字](https://leetcode.cn/problems/single-number)【模式：异或抵消；利用 $a \oplus a = 0$】
  - [137. 只出现一次的数字 II](https://leetcode.cn/problems/single-number-ii)【模式：数字电路设计 / 状态机；利用 `ones` 和 `twos` 记录位状态，或统计位 1 个数模 3】
  - [260. 只出现一次的数字 III](https://leetcode.cn/problems/single-number-iii)【模式：异或分组；利用 lowbit 将两个不同数字分到不同组】
  - [421. 数组中两个数的最大异或值](https://leetcode.cn/problems/maximum-xor-of-two-numbers-in-an-array)【模式：0-1 Trie / 贪心位构造】

### 4. 快速幂与几何 (Fast Power & Geometry)
  - [50. Pow(x, n)](https://leetcode.cn/problems/powx-n)【模式：快速幂；$O(\log n)$ 复杂度】
  - [149. 直线上最多的点数](https://leetcode.cn/problems/max-points-on-a-line)【模式：几何 + 哈希表；枚举起点，计算斜率归类】
  - [3102. 最小化曼哈顿距离](https://leetcode.cn/problems/minimize-manhattan-distances)【模式：坐标转换；曼哈顿距离与切比雪夫距离的转换 $O(n)$】

## 搜索问题核心分类与总结 (Search Strategies)

### 1. BFS：状态空间的最短路径 (Breadth-First Search)
- **基础网格与层级遍历**
  - [200. 岛屿数量](https://leetcode.cn/problems/number-of-islands)【模式：连通分量统计】
  - [909. 蛇梯棋](https://leetcode.cn/problems/snakes-and-ladders)【模式：状态转移模拟】
  - [45. 跳跃游戏 II](https://leetcode.cn/problems/jump-game-ii)
- **多源 BFS (Multi-source BFS)**
  - [2812. 找出最安全路径](https://leetcode.cn/problems/find-the-safest-path-in-a-grid) 【模式：预处理所有起点距离】
  - [2258. 逃离火灾](https://leetcode.cn/problems/escape-the-spreading-fire/)
  - [2477. 到达首都的最少油耗](https://leetcode.cn/problems/minimum-fuel-cost-to-report-to-the-capital)
- **双向 BFS (Bi-directional BFS)**
  - [127. 单词接龙](https://leetcode.cn/problems/word-ladder) 【模式：极大减小搜索树规模】
  - [433. 最小基因变化](https://leetcode.cn/problems/minimum-genetic-mutation)
- **优先队列 BFS (Dijkstra 思想)**
  - [407. 接雨水 II](https://leetcode.cn/problems/trapping-rain-water-ii/description/) 【模式：3D 边界收缩】

### 2. DFS 与回溯：全量枚举与约束满足 (DFS & Backtracking)

#### 💡 回溯四阶梯与去重口诀
    | 阶梯 | 核心场景 | 去重/控制逻辑 | 关键代码 |
    | :--- | :--- | :--- | :--- |
    | **1. 基础回溯** | 简单组合 (17题) | 递归深度控制索引 | `dfs(i + 1)` |
    | **2. 组合去重** | 选 k 个数 (77题) | `start` 索引控制单向搜索 | `for (int i = start; ...)` |
    | **3. 状态压缩** | 全排列 (46题) | `Bitmask` 替代 `visited` 数组 | `if (!(mask & (1 << j)))` |
    | **4. 排列去重** | 有重全排列 (47题) | **排序 + 相邻状态校验** | `if (j > 0 && nums[j] == nums[j-1] && !used[j-1])` |
    | **5. 复杂约束** | 棋盘/皇后 (52题) | **空间换时间 (哈希标记)** | `if (!cols[c] && !diag[r+c])` |
    | **6. 余额控制** | 括号生成 (22题) | **动态维护待匹配余额** | `if (remain > 0) dfs(..., remain-1)` |
    | **7. 矩阵回溯** | 单词搜索 (79题) | **原地标记 + 字符还原** | `board[r][c] = '#'; dfs(); board[r][c] = tmp;` |

> **去重口诀**：
> - **组合**靠 `start`：不回头看，一路向右。
> - **排列**靠 `used`：全员参与，位掩码标记。
> - **重复**靠**排序**：前人未用，后人莫入（`!used[i-1]`）。
> - **棋盘**靠标记：列号、和、差，三位一体定乾坤。
> - **括号**看余额：左括号不超标，右括号不透支。
> - **矩阵**靠沉岛：先占位再递归，事后记得还原。

#### 组合、排列与路径
  - [17. 电话号码的字母组合](https://leetcode.cn/problems/letter-combinations-of-a-phone-number) 【模式：基础回溯；核心：递归深度控制数字索引，for 循环遍历字母映射】
  - [77. 组合](https://leetcode.cn/problems/combine) 【模式：组合回溯；核心：【口诀】组合靠 `start`：不回头看，一路向右】
  - [39. 组合总和](https://leetcode.cn/problems/combination-sum) 【模式：重复选组合；核心：【原理】传递当前索引 `i` 而非 `i+1` 实现元素可重复选取】
  - [46. 全排列](https://leetcode.cn/problems/permutations) 【模式：排列回溯；核心：【口诀】排列靠 `used`：全员参与，位掩码标记】
  - [47. 全排列 II](https://leetcode.cn/problems/permutations-ii) 【模式：有重排列；核心：【原理】重复靠排序：前人未用，后人莫入（`!used[i-1]`）】
  - [52. N 皇后 II](https://leetcode.cn/problems/n-queens-ii) 【模式：棋盘回溯；核心：【口诀】棋盘靠标记：列号、和、差，三位一体定乾坤】
  - [22. 括号生成](https://leetcode.cn/problems/generate-parentheses) 【模式：配对回溯；核心：【口诀】括号看余额：左括号不超标，右括号不透支】
  - [79. 单词搜索](https://leetcode.cn/problems/word-search) 【模式：矩阵回溯；核心：【口诀】矩阵靠沉岛：先占位再递归，事后记得还原】
  - [LCR 086. 分割回文串](https://leetcode.cn/problems/M99OJA)

- **树/图中的深度搜索**
  - [112. 路径总和](https://leetcode.cn/problems/path-sum)
  - [113. 路径总和 II](https://leetcode.cn/problems/path-sum-ii)
  - [437. 路径总和 III](https://leetcode.cn/problems/path-sum-iii/)
  - [133. 克隆图](https://leetcode.cn/problems/clone-graph) 【模式：哈希表防止死循环】
  - [2101. 引爆最多的炸弹](https://leetcode.cn/problems/detonate-the-maximum-bombs)

- **复杂约束与剪枝**
  - [51. N 皇后](https://leetcode.cn/problems/n-queens)【回溯】
  - [928. 尽量减少恶意软件的传播 II](https://leetcode.cn/problems/minimize-malware-spread-ii)
  - [2850. 将石头分散到网格图的最少移动次数](https://leetcode.cn/problems/minimum-moves-to-spread-stones-over-grid)
  - [2698. 求一个整数的惩罚数](https://leetcode.cn/problems/find-the-punishment-number-of-an-integer)
  - [3211. 生成不含相邻零的二进制字符串](https://leetcode.cn/problems/generate-binary-strings-without-adjacent-zeros)

### 3. 逆向思维与启发式搜索 (Advanced Search)
- **逆向搜索**
  - [130. 被围绕的区域](https://leetcode.cn/problems/surrounded-regions)【模式：从边界向内部反向标记】
- **启发式搜索 (Heuristic Search)**
  - [八数码问题](https://github.com/wxquare/online-algorithm-coding/blob/master/other/8-puzzle.cc) 【模式：A* 算法】

## DP 问题 (Dynamic Programming - 核心模式归类)

### DP 类问题处理五部曲总结

| 步骤 | 核心任务 (Key Action) | 你的代码体现 (Example) | 空间优化思路 (Space Optimization) |
| :--- | :--- | :--- | :--- |
| **1. 状态定义** | 明确 `dp` 数组各维度的物理含义（是长度、最值还是布尔值？对应什么区间？） | `dp[i][j]` 表示到达坐标 `(i,j)` 的最小路径和 | **维度压缩**：若当前状态只依赖前一状态，可将二维数组降为一维（或常数个变量）。 |
| **2. 转移方程** | 逻辑推导过程，包括不同条件下的决策 | `dp[i][j] = min(dp[i-1][j], dp[i][j-1]) + grid[i][j]` | **原地修改**：如果输入数组（如 `grid`）后续不再使用，可以直接在原数组上操作实现 $O(1)$ 额外空间。 |
| **3. 初始边界** | 算法开始的基石（如单字符情况、空串情况），确定无需推导的“种子”值 | 初始化 `dp[0][0]`，并单独处理首行 `dp[0][i]` 和首列 `dp[i][0]` | **虚拟边界**：有时可以多申请一行/一列（如 `dp[m+1][n+1]`）并填入占位值，从而统一循环内的逻辑。 |
| **4. 计算顺序** | 确定循环的方向（Top-down vs Bottom-up），由状态依赖关系决定 | 使用双重 `for` 循环，从左到右、从上到下遍历 | **倒序遍历**：在 0/1 背包等问题中，通过倒序遍历一维 DP 数组，可以防止当前层的计算污染待使用的旧数据。 |
| **5. 最终结果** | 确定答案在 `dp` 表中的存储位置 | 返回 `dp[m-1][n-1]` | **状态追踪**：如果不仅要结果还要路径，通常需要额外的 `parent` 数组记录来源，空间优化此时会受限。 |
| **6. 复杂度分析** | 分析时间与空间开销 | 时间 $O(M \times N)$，空间 $O(M \times N)$ | **时空权衡**：有时为了降低时间复杂度（如利用前缀和优化转移），可能会增加空间复杂度。 |

### 1. 基础线性 DP (1D/2D 填表)
> *最基础的递推，dp[i] 只依赖于前面几个状态*

#### A. 斐波那契/爬楼梯模型
  - [509. 斐波那契数](https://leetcode.cn/problems/fibonacci-number)
  - [70. 爬楼梯](https://leetcode.cn/problems/climbing-stairs)
  - [1137. 第 N 个泰波那契数](https://leetcode.cn/problems/n-th-tribonacci-number)
  - [746. 使用最小花费爬楼梯](https://leetcode.cn/problems/min-cost-climbing-stairs)
  - [91. 解码方法](https://leetcode.cn/problems/decode-ways)【模式：爬楼梯变体；判断单字符和双字符是否有效，`dp[i] += dp[i-1] + dp[i-2]`】

#### B. 网格路径模型 (Grid)
  - [62. 不同路径](https://leetcode.cn/problems/unique-paths)【模式：基础网格 DP；`dp[i][j] = dp[i-1][j] + dp[i][j-1]` 或组合数】
  - [63. 不同路径 II](https://leetcode.cn/problems/unique-paths-ii)【模式：障碍物处理；若 `grid[i][j] == 1` 则 `dp[i][j] = 0`】
  - [64. 最小路径和](https://leetcode.cn/problems/minimum-path-sum)【模式：原地修改；`grid[i][j] += min(左, 上)`】
  - [120. 三角形最小路径和](https://leetcode.cn/problems/triangle)【模式：自底向上 DP；`dp[j] = min(dp[j], dp[j+1]) + val`，空间优化至 O(N)】
  - [221. 最大正方形](https://leetcode.cn/problems/maximal-square)【模式：木桶短板原理；`dp[i][j] = min(左, 上, 左上) + 1`】
  - [1277. 统计全为 1 的正方形子矩阵](https://leetcode.cn/problems/count-square-submatrices-with-all-ones)【模式：最大正方形变体；`dp[i][j]` 既是最大边长，也是以该点为右下角的正方形个数】
  - [3148. 矩阵中的最大得分](https://leetcode.cn/problems/maximum-difference-score-in-a-grid)【模式：二维前缀最小值；维护矩形区域内的最小值 `min_val`】

#### C. 简单一维推导
  - [53. 最大子数组和](https://leetcode.cn/problems/maximum-subarray)【模式：Kadane 算法；`dp[i] = max(nums[i], dp[i-1] + nums[i])`】
  - [918. 环形子数组的最大和](https://leetcode.cn/problems/maximum-sum-circular-subarray)【模式：Kadane 变体；`max(最大子数组和, 总和 - 最小子数组和)`】
  - [152. 乘积最大子数组](https://leetcode.cn/problems/maximum-product-subarray)【模式：双状态 DP；同时维护 `max_prod` 和 `min_prod` 以应对负数】
  - [91. 解码方法](https://leetcode.cn/problems/decode-ways)【模式：爬楼梯变体；判断单字符和双字符是否有效，`dp[i] += dp[i-1] + dp[i-2]`】
  - [198. 打家劫舍](https://leetcode.cn/problems/house-robber)【模式：线性 DP；`dp[i] = max(dp[i-1], dp[i-2] + nums[i])`】
  - [213. 打家劫舍 II](https://leetcode.cn/problems/house-robber-ii)【模式：环形 DP；拆分为 `[0, n-2]` 和 `[1, n-1]` 两次线性 DP】


### 2. 状态机 DP (State Machine)
> *核心在于定义“持有”、“冷冻”、“卖出”等有限状态，画状态转移图*

#### A. 股票系列
  - [121. 买卖股票的最佳时机](https://leetcode.cn/problems/best-time-to-buy-and-sell-stock)【模式：一次交易；维护 `min_price`，计算 `price - min_price`】
  - [122. 买卖股票的最佳时机 II](https://leetcode.cn/problems/best-time-to-buy-and-sell-stock-ii)【模式：无限次交易；贪心收集所有正收益 `max(0, p[i]-p[i-1])`】
  - [123. 买卖股票的最佳时机 III](https://leetcode.cn/problems/best-time-to-buy-and-sell-stock-iii)【模式：两次交易；维护 `buy1, sell1, buy2, sell2` 四个状态】
  - [188. 买卖股票的最佳时机 IV](https://leetcode.cn/problems/best-time-to-buy-and-sell-stock-iv)【模式：K 次交易；维护 `buy[k]` 和 `sell[k]` 数组，若 `k > n/2` 退化为无限次】
  - [309. 买卖股票的最佳时机含冷冻期](https://leetcode.cn/problems/best-time-to-buy-and-sell-stock-with-cooldown)【模式：状态机；持有、不持有(处于冷冻期)、不持有(非冷冻期)】
  - [714. 买卖股票的最佳时机含手续费](https://leetcode.cn/problems/best-time-to-buy-and-sell-stock-with-transaction-fee)【模式：状态机；`sell` 转移时减去 `fee`】

#### B. 其他状态机
  - [552. 学生出勤记录 II](https://leetcode.cn/problems/student-attendance-record-ii)
  - [3098. 求出所有子序列的能量和](https://leetcode.cn/problems/find-the-sum-of-subsequence-powers)

#### C. 打家劫舍系列汇总 (House Robber)
  - [198. 打家劫舍](https://leetcode.cn/problems/house-robber)【模式：线性 DP；`dp[i] = max(dp[i-1], dp[i-2] + nums[i])`】
  - [213. 打家劫舍 II](https://leetcode.cn/problems/house-robber-ii)【模式：环形 DP；拆分为 `[0, n-2]` 和 `[1, n-1]` 两次线性 DP】
  - [337. 打家劫舍 III](https://leetcode.cn/problems/house-robber-iii)【模式：树形 DP；每个节点返回 `{偷, 不偷}` 两个状态】
  - [2560. 打家劫舍 IV](https://leetcode.cn/problems/house-robber-iv)【模式：二分答案 + 贪心；最大值最小化问题】

### 3. 序列 DP (双串/单串)
> *处理字符串或数组子序列问题，核心是 LCS/LIS 模型*

#### A. 单串 LIS 模型 ($O(n^2)$ 或 $O(n \log n)$)
  - [300. 最长递增子序列](https://leetcode.cn/problems/longest-increasing-subsequence)【模式：DP / 贪心 + 二分；核心：维护 `tails` 数组实现 $O(n \log n)$】
  - [673. 最长递增子序列的个数](https://leetcode.cn/problems/number-of-longest-increasing-subsequence)【模式：双 DP 数组；核心：同步维护 `len[i]` 和 `cnt[i]`，注意相等长度时的累加】
  - [354. 俄罗斯套娃信封问题](https://leetcode.cn/problems/russian-doll-envelopes)【模式：二维 LIS；核心：按 w 升序排序，w 相同按 h 降序排序，转化为对 h 求 LIS】
  - [646. 最长数对链](https://leetcode.cn/problems/maximum-length-of-pair-chain)【模式：贪心；核心：按第二个数排序，贪心选择结束最早的区间】

#### B. 双串 LCS 模型 (二维表,m + 1,n + 1 处理空串的情况)
  - [1143. 最长公共子序列](https://leetcode.cn/problems/longest-common-subsequence)【模式：双串 DP；`dp[i][j] = s1[i]==s2[j] ? dp[i-1][j-1]+1 : max(左, 上)`；**注意：DP 数组大小为 (M+1)*(N+1) 处理空串**】
  - [72. 编辑距离](https://leetcode.cn/problems/edit-distance)【模式：增删改三选一；`dp[i][j] = min(插入, 删除, 替换) + 1`；**注意：DP 数组大小为 (M+1)*(N+1) 处理空串**】
  - [97. 交错字符串](https://leetcode.cn/problems/interleaving-string)【模式：双串 DP；`dp[i][j]` 表示 `s1[0..i]` 和 `s2[0..j]` 能否交错组成 `s3[0..i+j]`；**注意：DP 数组大小为 (M+1)*(N+1) 处理空串**】
  - [583. 两个字符串的删除操作](https://leetcode.cn/problems/delete-operation-for-two-strings)【模式：LCS 变体；结果为 `m + n - 2 * LCS`；**注意：DP 数组大小为 (M+1)*(N+1) 处理空串**】
  - [1035. 不相交的线](https://leetcode.cn/problems/uncrossed-lines)【模式：LCS 本质；完全等同于最长公共子序列；**注意：DP 数组大小为 (M+1)*(N+1) 处理空串**】
  - [115. 不同的子序列](https://leetcode.cn/problems/distinct-subsequences)【模式：计数 DP；`s[i]==t[j]` 时可选匹配或不匹配，`dp[i][j] = dp[i-1][j-1] + dp[i-1][j]`；**注意：DP 数组大小为 (M+1)*(N+1) 处理空串**】

#### C. 回文串模型
  - [5. 最长回文子串](https://leetcode.cn/problems/longest-palindromic-substring)【模式：中心扩展法；枚举中心向两边扩散，空间 $O(1)$】
  - [516. 最长回文子序列](https://leetcode.cn/problems/longest-palindromic-subsequence)【模式：区间 DP；`dp[i][j]` 表示 `s[i...j]` 的最长回文子序列长度】
  - [647. 回文子串](https://leetcode.cn/problems/palindromic-substrings)【模式：中心扩展法；统计回文中心扩展过程中的有效回文数】
  - [132. 分割回文串 II](https://leetcode.cn/problems/palindrome-partitioning-ii)【模式：两次 DP；先预处理回文表，再求最小分割数】


### 4. 划分型 DP (Partition)
> *将数组/字符串切分为 k 段，求最优解*

  - [343. 整数拆分](https://leetcode.cn/problems/integer-break)
  - [279. 完全平方数](https://leetcode.cn/problems/perfect-squares)
  - [410. 分割数组的最大值](https://leetcode.cn/problems/split-array-largest-sum)
  - [887. 鸡蛋掉落](https://leetcode.cn/problems/super-egg-drop)
  - [1043. 分隔数组以得到最大和](https://leetcode.cn/problems/partition-array-for-maximum-sum)

### 5. 背包 DP (Knapsack)
> *组合优化问题，关注容量与价值*

#### A. 0/1 背包
  - [416. 分割等和子集](https://leetcode.cn/problems/partition-equal-subset-sum)
  - [494. 目标和](https://leetcode.cn/problems/target-sum)
  - [474. 一和零](https://leetcode.cn/problems/ones-and-zeroes)

#### B. 完全背包
  - [322. 零钱兑换](https://leetcode.cn/problems/coin-change)【模式：完全背包；求凑成总金额的最少硬币数】
  - [518. 零钱兑换 II](https://leetcode.cn/problems/coin-change-ii)【模式：完全背包；求凑成总金额的组合数】
  - [279. 完全平方数](https://leetcode.cn/problems/perfect-squares)
  - [139. 单词拆分](https://leetcode.cn/problems/word-break)【模式：完全背包 / 线性 DP；判断字符串能否由字典构成】
  - [1449. 数位成本和为目标值的最大数字](https://leetcode.cn/problems/form-largest-integer-with-digits-that-add-up-to-target)

#### C. 多重/分组背包
  - [1155. 掷骰子等于目标和的方法数](https://leetcode.cn/problems/number-of-dice-rolls-with-target-sum)

### 6. 区间 DP (Interval)
> *从小区间合并到大区间，枚举分割点 k*

  - [312. 戳气球](https://leetcode.cn/problems/burst-balloons)
  - [877. 石子游戏](https://leetcode.cn/problems/stone-game)
  - [1140. 石子游戏 II](https://leetcode.cn/problems/stone-game-ii)
  - [375. 猜数字大小 II](https://leetcode.cn/problems/guess-number-higher-or-lower-ii)

### 7. 树形 DP (Tree DP)
> *自底向上汇总信息，或换根 DP*

#### A. 子树贡献/直径
  - [543. 二叉树的直径](https://leetcode.cn/problems/diameter-of-binary-tree)
  - [124. 二叉树中的最大路径和](https://leetcode.cn/problems/binary-tree-maximum-path-sum)
  - [337. 打家劫舍 III](https://leetcode.cn/problems/house-robber-iii)
  - [2246. 相邻字符不同的最长路径](https://leetcode.cn/problems/longest-path-with-different-adjacent-characters)

#### B. 换根 DP
  - [834. 树中距离之和](https://leetcode.cn/problems/sum-of-distances-in-tree)
  - [2581. 统计可能的树根数目](https://leetcode.cn/problems/count-number-of-possible-root-nodes)
  - [310. 最小高度树](https://leetcode.cn/problems/minimum-height-trees)

### 8. 状压 DP (Bitmask)
> *数据范围 n < 20，用二进制表示集合*

  - [464. 我能赢吗](https://leetcode.cn/problems/can-i-win)
  - [526. 优美的排列](https://leetcode.cn/problems/beautiful-arrangement)
  - [847. 访问所有节点的最短路径](https://leetcode.cn/problems/shortest-path-visiting-all-nodes)
  - [698. 划分为k个相等的子集](https://leetcode.cn/problems/partition-to-k-equal-sum-subsets)
  - [2741. 特别的排列](https://leetcode.cn/problems/special-permutations)

### 9. 数位 DP (Digit DP)
> *按位填数，通常配合记忆化搜索*

  - [233. 数字 1 的个数](https://leetcode.cn/problems/number-of-digit-one)
  - [902. 最大为 N 的数字组合](https://leetcode.cn/problems/numbers-at-most-n-given-digit-set)
  - [1012. 至少有 1 位重复的数字](https://leetcode.cn/problems/numbers-with-repeated-digits)
  - [2376. 统计特殊整数](https://leetcode.cn/problems/count-special-integers)
  - [600. 二进制不含连续1的非负整数](https://leetcode.cn/problems/non-negative-integers-without-consecutive-ones)

### 10. 其他/高级 DP
  - [10. 正则表达式匹配](https://leetcode.cn/problems/regular-expression-matching)
  - [44. 通配符匹配](https://leetcode.cn/problems/wildcard-matching)
  - [2713. 矩阵中严格递增的单元格数](https://leetcode.cn/problems/maximum-strictly-increasing-cells-in-a-matrix)
  - [2836. 在传球游戏中最大化函数值](https://leetcode.cn/problems/maximize-value-of-function-in-a-ball-pfassing-game) (倍增)




## 工程小技巧 (Engineering Tips)

### 1. 数组快速清零
在 C++ 中，局部变量（栈上分配）默认包含随机垃圾值。使用以下语法可实现极致高效的清零：
```cpp
int arr[9][9] = {0}; // 显式初始化第一个元素为 0，其余元素自动补零
int arr[9][9] = {};  // C++11 简写，全员清零
```
- **注意**：如果不加 `{0}`，数组内容将不可预测，这是初学者最常见的 Bug 来源。
- **性能**：编译器通常会将其优化为内联的 `memset` 或专门的 CPU 指令，比手动 `for` 循环快得多。

### 2. 字符串单词拆分 (stringstream)
在处理以空格分隔的字符串（如“hello world”）时，手动控制指针解析既繁琐又容易出错（需考虑首尾空格、多空格等情况）。
```cpp
#include <sstream>
string s = "  hello   world  ";
stringstream ss(s);
string word;
while (ss >> word) {
    // 自动跳过所有空格，依次提取出 "hello" 和 "world"
}
```

### 3. 搜索策略选择指南
在算法竞赛或面试中，面对复杂的搜索问题，快速判断技术路线是节省时间的关键：
- **求最短路径 / 最小步数**：首选 **BFS**（利用其层级遍历的天然最短性）。
- **求所有方案 / 排列组合**：首选 **DFS + 回溯**（全量枚举状态空间）。
- **在单调 / 有序空间找最优值**：首选 **二分答案**（将最优化问题转化为判定问题 `check(mid)`）。
- **状态空间爆炸**：优先考虑 **双向 BFS**（极大减小搜索树规模）或 **记忆化搜索**（DFS + Memo，避免重复计算）。

## C++ 字符处理函数速查

| 函数名 | 检查内容 | 说明 |
|--------|----------|------|
| `isdigit(c)` | 是否为数字 (0-9) | 数字字符 |
| `isalpha(c)` | 是否为字母 (a-z, A-Z) | 纯字符/字母 |
| `isalnum(c)` | 是否为字母或数字 | 字母数字混合 |
| `tolower(c)` | 转换为小写 | 字符转换 |
| `toupper(c)` | 转换为大写 | 字符转换 |

## 参考
- [leetcode](https://leetcode.cn/)
- [灵茶山艾府](https://leetcode.cn/u/endlesscheng/)
- [宫水三叶](https://github.com/SharingSource/LogicStack-LeetCode)
- [代码随想录](https://github.com/youngyangyang04/leetcode-master)
- https://oi-wiki.org/
- https://zerotrac.github.io/leetcode_problem_rating/
- https://cp-algorithms.com/
- https://leetcode.cn/circle/discuss/tXLS3i/
- https://leetcode.Studyplan/selected-coding-interview/

