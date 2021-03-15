---
title: 熟悉常用的数据结构和算法
categories: 
- C/C++
---



## design pattern
- [单例模式的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/singleton.cpp)
- [工厂模式的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/factory.cpp)
- [观察者模式的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/observer.cpp)
- [生产者消费者模式的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/producer_consumer.cpp)

## 大数据算法总结
- [教你如何迅速秒杀掉99%的海量数据处理面试题](https://juejin.cn/post/6844903519640616967)
- 海量日志数据，提取出某日访问百度次数最多的那个IP
- 寻找热门查询，300万个查询字符串中统计最热门的10个查询
- 有一个1G大小的一个文件，里面每一行是一个词，词的大小不超过16字节，内存限制大小是1M。返回频数最高的100个词
- 海量数据分布在100台电脑中，想个办法高效统计出这批数据的TOP10
- 有10个文件，每个文件1G，每个文件的每一行存放的都是用户的query，每个文件的query都可能重复。要求你按照query的频度排序
- 5亿个int找它们的中位数

其它：
- [C++ MyString类的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/mystring.cpp)
- [C++ MyVector类的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/myvector.cpp)
- [二分查找递归和非递归的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/binary_search.cpp)
- [排序算法的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/sort.cpp)
- [字符串和数字的相互转换atoi和itoa函数的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/itoa_atoi.cpp)
- [memcpy/memset/strcpy/strncpy函数的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/str_function.cpp)
- 大数加减乘除
- KMP字符串匹配算法]
- [快排查找第k大的数](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/select_k.cpp)
- [全排列组合的实现](https://www.cnblogs.com/wxquare/p/4719228.html)

- [C++ 不可继承类的实现](https://www.cnblogs.com/wxquare/p/7280025.html)
- [C++ 读写锁的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/read_write_locker.cpp) 
- 哈希算法、冲突解决和简单实现
- 平衡二叉树、红黑树、B树、Trie树
- bitmamp布隆滤波器
- 暴力：dfs和bfs
- 哈希一致性算法
- 逻辑问题
- 100亿个数选top5，小根堆
- 唯一订单号问题，并发量高的话怎么解决
- 跳跃表，为什么使用跳跃表而不使用红黑树
- hash表设计要注意什么问题
- LRU的实现
- 布隆过滤器,bloom filter 使用场景，优缺点.与 hash set的比较
- 经典算法
- 单例模式. https://www.liwenzhou.com/posts/Go/singleton_in_go/
- 求平方根（根号n）的两种算法——二分法
- 雪花算法SnowFlake
- 洗牌算法
- 蓄水池抽样，从m个数中抽取n个数
- 经典算法
- 单例模式. https://www.liwenzhou.com/posts/Go/singleton_in_go/
- 求平方根（根号n）的两种算法——二分法
- 雪花算法SnowFlake
- 洗牌算法
- 蓄水池抽样，从m个数中抽取n个数

## 大数据算法
1. 十道海量数据处理面试题与十个方法大总结



## Array && String && Double Pointer
1. [两数之和](https://leetcode-cn.com/problems/two-sum/)
2. [三数之和](https://leetcode-cn.com/problems/3sum/)
3. [215. 数组中的第K个最大元素](https://leetcode-cn.com/problems/kth-largest-element-in-an-array/)
4. [18. 四数之和](https://leetcode-cn.com/problems/4sum/)
5. [189. 旋转数组](https://leetcode-cn.com/problems/rotate-array/),注意是否能使用额外的内存
6. [830. Positions of Large Groups](https://leetcode-cn.com/problems/positions-of-large-groups/)
7. [228. 汇总区间](https://leetcode-cn.com/problems/summary-ranges/)
8. [3. 无重复字符的最长子串](https://leetcode-cn.com/problems/longest-substring-without-repeating-characters/)
9. [1438. 绝对差不超过限制的最长连续子数组](https://leetcode-cn.com/problems/longest-continuous-subarray-with-absolute-diff-less-than-or-equal-to-limit/)
10. [1052. 爱生气的书店老板](https://leetcode-cn.com/problems/grumpy-bookstore-owner/)
11. [1438. 绝对差不超过限制的最长连续子数组](https://leetcode-cn.com/problems/longest-continuous-subarray-with-absolute-diff-less-than-or-equal-to-limit/)
12. [1004. 最大连续1的个数 III](https://leetcode-cn.com/problems/max-consecutive-ones-iii/)
13. [至少有 K 个重复字符的最长子串](https://leetcode-cn.com/problems/longest-substring-with-at-least-k-repeating-characters/)
14. [54 螺旋矩阵](https://leetcode-cn.com/problems/spiral-matrix/)
15. 

## stack,单调栈
1. [503. 下一个更大元素 II](https://leetcode-cn.com/problems/next-greater-element-ii/)
2. [496. 下一个更大元素 I](https://leetcode-cn.com/problems/next-greater-element-i/)
3. [556. 下一个更大元素 III](https://leetcode-cn.com/problems/next-greater-element-iii/)
4. [739. 每日温度](https://leetcode-cn.com/problems/daily-temperatures/)
5. [基本计算器](https://leetcode-cn.com/problems/basic-calculator/)
6. [基本计算器2](https://leetcode-cn.com/problems/basic-calculator-ii/)
7. 

## Graph && Union Find && DFS && BFS && Topological-sort
1. [547. 省份的数量](https://leetcode-cn.com/problems/number-of-provinces/)
2. [399. 除法求值](https://leetcode-cn.com/problems/evaluate-division/)(graph)
3. [ws面试：有一个矩形格子框，每个框都有一个字母，需要你找到路径，使得这条路径上的字母都不重复，请问这个最长的路径是多长?]
4. [执行交换操作后的最小汉明距离](https://leetcode-cn.com/problems/minimize-hamming-distance-after-swap-operations/)
5. [1202. 交换字符串中的元素](https://leetcode-cn.com/problems/smallest-string-with-swaps/)
6. [684. 冗余连接](https://leetcode-cn.com/problems/redundant-connection/)
7. [207. 课程表](https://leetcode-cn.com/problems/course-schedule/)
8. [210. 课程表 II](https://leetcode-cn.com/problems/course-schedule-ii/)
9. [721. 账户合并](https://leetcode-cn.com/problems/accounts-merge/)
并查集时间复杂读分析(https://leetcode-cn.com/problems/number-of-provinces/solution/jie-zhe-ge-wen-ti-ke-pu-yi-xia-bing-cha-0unne/)
10. [相似字符串组](https://leetcode-cn.com/problems/similar-string-groups/)

## Dynamic Programming && DFS
1. [300. 最长递增子序列](https://leetcode-cn.com/problems/longest-increasing-subsequence/)
2. [poj，ws面试,最长的递增递减子序列长度](https://my.oschina.net/Alexanderzhou/blog/205171)
3. [121. 买卖股票的最佳时机](https://leetcode-cn.com/problems/best-time-to-buy-and-sell-stock/)，最多买卖一次
4. [122. 买卖股票的最佳时机](https://leetcode-cn.com/problems/best-time-to-buy-and-sell-stock-ii/)，可以买卖任意次数
5. [123. 买卖股票的最佳时机 III](https://leetcode-cn.com/problems/best-time-to-buy-and-sell-stock-iii/)，最多可以买卖两次
6. [188. 买卖股票的最佳时机 IV](https://leetcode-cn.com/problems/best-time-to-buy-and-sell-stock-iv/)，最多可以买卖K次
7. [309. 最佳买卖股票时机含冷冻期](https://leetcode-cn.com/problems/best-time-to-buy-and-sell-stock-with-cooldown/)
8. [714. 买卖股票的最佳时机含手续费](https://leetcode-cn.com/problems/best-time-to-buy-and-sell-stock-with-transaction-fee/)
9. [ 完成所有工作的最短时间](https://leetcode-cn.com/problems/find-minimum-time-to-finish-all-jobs/)，状态压缩DP或者dfs加剪枝
11. [回文字符串分割IV](https://leetcode-cn.com/problems/palindrome-partitioning-iv)
12. [354. 俄罗斯套娃信封问题](https://leetcode-cn.com/problems/russian-doll-envelopes/)

## Linked List && Tree
1. [交换链表中的第K个节点和倒数第K个节点](https://leetcode-cn.com/problems/swapping-nodes-in-a-linked-list/)
3. [lowest common ancestor](https://leetcode-cn.com/problems/lowest-common-ancestor-of-a-binary-tree/)
4. [二叉树的前、中、后序遍历,非递归的实现](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/binary_tree.cpp)（12.21）
5. [二叉搜索树的查找、插入和删除](https://github.com/wxquare/programming/blob/master/oj/datastruct-algorithm/binary_search_tree.cpp)
5. [链表反转](https://leetcode-cn.com/problems/reverse-linked-list/)
6. [面试题 02.05. 链表求和](https://leetcode-cn.com/problems/sum-lists-lcci/)
7. [2. 两数相加](https://leetcode-cn.com/problems/add-two-numbers/)

## Design DataStruct
1. [LRU](https://leetcode-cn.com/problems/lru-cache/)
2. [LFU](https://leetcode-cn.com/problems/lfu-cache/)
3. [HashSet](https://leetcode-cn.com/problems/design-hashset/)
4. [HashMap](https://leetcode-cn.com/problems/design-hashmap/)
