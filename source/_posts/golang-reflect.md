---
title: golang 反射
---



　　听说过java的反射（reflect）概念，但不是很了解。golang也提供反射机制，简单的来说，反射能在运行期获取接口对象的类型、数据和方法。golang的反射机制依赖于接口，因为接口对象保存了自身类型和实际对象的对象的类型和数据。用好反射需要理解：**实际对象、接口对象、反射类型Type和反射值Value类型**。reflect包提供下面两个入口函数，将任何传入的对象转换为接口类型，从而获取反射类型（Type）和反射值（Value）：
  
	func TypeOf( i interface{}) Type
	func ValueOf(i interface{}) Value



