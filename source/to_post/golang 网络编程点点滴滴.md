---
title: golang 网络编程点点滴滴 
---

## 1、echo服务器和客户端
    
    package main
    
    import (
    	"bufio"
    	"fmt"
    	"net"
    	"os"
    )
    
    func main() {
    	l, err := net.Listen("tcp", "127.0.0.1:8080")
    	if err != nil {
    		fmt.Println("Error listening...")
    		os.Exit(1)
    	}
    	defer l.Close()
    	for {
    		conn, err := l.Accept()
    		if err != nil {
    			fmt.Println("Error Accepting...")
    			os.Exit(2)
    		}
    		fmt.Printf("Receive message %s -> %s\n", conn.RemoteAddr(), conn.LocalAddr())
    		go handleRequest(conn)
    	}
    }
    
    func handleRequest(conn net.Conn) {
    	reader := bufio.NewReader(conn)
    	defer conn.Close()
    	for {
    		message, err := reader.ReadString('\n')
    		if err != nil {
    			os.Exit(3)
    		}
    		fmt.Println(string(message))
    		conn.Write([]byte(message))
    	}
    }

-

    package main
    
    import (
    	"fmt"
    	"net"
    	"os"
    	"time"
    	// "strconv"
    )
    func main() {
    	conn, err := net.Dial("tcp", "127.0.0.1:8080")
    	if err != nil {
    		fmt.Println("Error Dial....")
    		os.Exit(1)
    	}
    	defer conn.Close()
    
    	fmt.Printf("Connectiong to %s\n", conn.RemoteAddr())
    
    	for {
    		_, err = conn.Write([]byte("hello " + "\r\n"))
    		if err != nil {
    			fmt.Println("Error to send message because of", err.Error())
    			os.Exit(2)
    		}
    
    		buf := make([]byte, 1024)
    		reqLen, err := conn.Read(buf)
    		if err != nil {
    			fmt.Println("Error to read message because of ", err.Error())
    			return
    		}
    		fmt.Println(string(buf[:reqLen-1]))
    		time.Sleep(1 * time.Second)
    	}
    
    }


## 2、设置链接的超时处理
	conn.SetDeadline(time.Now().Add(time.Duration(1) * time.Second))
