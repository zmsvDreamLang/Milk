-- 测试基本算术运算和打印
ffi.godef([[
package main

func main() {
    x = 10
    y = 5
    println("x =", x)
    println("y =", y)
    println("x + y =", x + y)
    println("x - y =", x - y)
    println("x * y =", x * y)
    println("x / y =", x / y)
}
]])
local res = ffi.goexec()
print("算术运算测试结果:\n" .. res)

-- 测试循环和条件语句
ffi.godef([[
package main

func main() {
    for i := 1; i <= 100; i++ {
        if i % 2 == 0 {
            println(i, "是偶数")
        } else {
            println(i, "是奇数")
        }
    }
}
]])
res = ffi.goexec()
print("循环和条件语句测试结果:\n" .. res)

-- 测试自定义函数
ffi.godef([[
package main

func add(a, b int) int {
    return a + b
}

func main() {
    result := add(10, 20)
    println("10 + 20 =", result)
}
]])

res = ffi.goexec()

print("自定义函数测试结果:\n" .. res)

-- 测试切片操作
ffi.godef([[
package main

func main() {
    numbers := []int{1, 2, 3, 4, 5}
    println("切片长度:", len(numbers))
    println("第三个元素:", numbers[2])
    numbers = append(numbers, 6)
    println("添加元素后的长度:", len(numbers))
}
]])
res = ffi.goexec()
print("切片操作测试结果:\n" .. res)

-- 测试函数定义和调用
ffi.godef([[
package main

func add(a, b int) int {
    return a + b
}

func main() {
    result := add(10, 20)
    println("10 + 20 =", result)
}
]])
res = ffi.goexec()
print("函数定义和调用测试结果:\n" .. res)

-- 测试结构体
ffi.godef([[
package main

type Person struct {
    Name string
    Age  int
}

func main() {
    p := Person{Name: "Alice", Age: 30}
    println("姓名:", p.Name)
    println("年龄:", p.Age)
}
]])
res = ffi.goexec()
print("结构体测试结果:\n" .. res)