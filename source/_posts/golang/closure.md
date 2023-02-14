---
title: Go Closure 使用场景介绍
categories:
- Golang
---

## What's Go closure?
    In Go, a closure is a function that has access to variables from its outer (enclosing) function's scope. The closure "closes over" the variables, meaning that it retains access to them even after the outer function has returned. This makes closures a powerful tool for encapsulating data and functionality and for creating reusable code.


### Encapsulating State

```Go
package main

import "fmt"

func counter() func() int {
    i := 0
    return func() int {
        i++
        return i
    }
}

func main() {
    c := counter()

    fmt.Println(c()) // Output: 1
    fmt.Println(c()) // Output: 2
    fmt.Println(c()) // Output: 3
}

```

### Implementing Callbacks
```Go
package main

import "fmt"

func forEach(numbers []int, callback func(int)) {
    for _, n := range numbers {
        callback(n)
    }
}

func main() {
    numbers := []int{1, 2, 3, 4, 5}

    // Define a callback function to apply to each element of the numbers slice.
    callback := func(n int) {
        fmt.Println(n * 2)
    }

    // Use the forEach function to apply the callback function to each element of the numbers slice.
    forEach(numbers, callback)
}

```

### Fibonacci

```Go
package main

import "fmt"

func memoize(f func(int) int) func(int) int {
    cache := make(map[int]int)
    return func(n int) int {
        if val, ok := cache[n]; ok {
            return val
        }
        result := f(n)
        cache[n] = result
        return result
    }
}

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
    fib := memoize(fibonacci)
    for i := 0; i < 10; i++ {
        fmt.Println(fib(i))
    }
}
```


### Factorial
```Go
package main

import "fmt"

func main() {
    factorial := func(n int) int {
        if n <= 1 {
            return 1
        }
        return n * factorial(n-1)
    }

    fmt.Println(factorial(5)) // Output: 120
}

```


### Event Handling
```Go
package main

import (
	"fmt"
	"time"
)

type Button struct {
	onClick func()
}

func NewButton() *Button {
	return &Button{}
}

func (b *Button) SetOnClick(f func()) {
	b.onClick = f
}

func (b *Button) Click() {
	if b.onClick != nil {
		b.onClick()
	}
}

func main() {
	button := NewButton()
	button.SetOnClick(func() {
		fmt.Println("Button Clicked!")
	})

	go func() {
		for {
			button.Click()
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Scanln()
}

```
