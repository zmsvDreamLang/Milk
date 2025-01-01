print("测试 calculuslib:")

-- 简单调用测试
print("\n简单调用测试:")

-- 测试求导
local function f1(x) return x^2 end
print("f(x) = x^2 的导数在 x = 2 处的值:")
print(calculus.derivative(f1, 2))

-- 测试积分
local function f2(x) return x^3 end
print("f(x) = x^3 在区间 [0, 1] 上的定积分:")
print(calculus.integral(f2, 0, 1))

-- 边界条件测试
print("\n边界条件测试:")

-- 测试求导的边界条件
print("f(x) = x^2 的导数在 x = 0 处的值:")
print(calculus.derivative(f1, 0))

print("f(x) = x^2 的导数在 x = 1e6 处的值:")
print(calculus.derivative(f1, 1e6))

-- 测试积分的边界条件
print("f(x) = x^3 在区间 [0, 0] 上的定积分:")
print(calculus.integral(f2, 0, 0))

print("f(x) = x^3 在区间 [0, 1e5] 上的定积分:")
print(calculus.integral(f2, 0, 1e5))

-- 性能测试
print("\n性能测试:")

-- 测试求导的性能
local start_time = os.clock()
for i = 1, 1000 do
    calculus.derivative(f1, i)
end
local end_time = os.clock()
print("求导性能测试耗时 (1000 次调用):", end_time - start_time, "秒")

-- 测试积分的性能
start_time = os.clock()
for i = 1, 1000 do
    calculus.integral(f2, 0, i)
end
end_time = os.clock()
print("积分性能测试耗时 (1000 次调用):", end_time - start_time, "秒")
