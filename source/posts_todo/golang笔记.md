###1. int 和 string 的变换
    #####string到int
    int,err:=strconv.Atoi(string)
    #####string到int64
    int64, err := strconv.ParseInt(string, 10, 64)
    #####int到string
    string:=strconv.Itoa(int)
    #####int64到string
    string:=strconv.FormatInt(int64,10)
###2. 创建对象
	obj1 := stockPosition{"abc",1.0,2.1}
	obj2 := &stockPosition{"bcd",1.0,3.1}
	//new
	obj3 := new(stockPosition)
	//make
	slice1 := make([]int,5)
	map1 := make(map[int]string,5)
	chan1 := make(chan int,5)

###3. golang闭包

###4. defer

###5. "_"

###6. golang API json ,struct结构中标签(Tag)的使用