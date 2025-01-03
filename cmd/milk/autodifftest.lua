-- 无需 require，calculus 库已经被自动加载

-- 定义一个函数 f(x) = sin(x^2)
function f_autodiff(x)
    return calculus.sin(calculus.pow(x, 2))
end

function f(x)
    return math.sin(x^2) 
end

-- 原有的计算部分
x = math.pi / 4
value, derivative = calculus.autodiff(f_autodiff, x)
print(string.format("f(%.4f) = %.6f", x, value))
print(string.format("f'(%.4f) = %.6f", x, derivative))

numerical_derivative = calculus.derivative(f, x)
print(string.format("Numerical f'(%.4f) = %.6f", x, numerical_derivative))

a, b = 0, 1
integral = calculus.integral(f, a, b)
print(string.format("Integral of f from %.1f to %.1f = %.6f", a, b, integral))

-- Lua原生实现的数值微分
function lua_numerical_derivative(f, x, h)
    h = h or 1e-8
    return (f(x + h) - f(x - h)) / (2 * h)
end

-- Milk自动微分库
milk_autodiff = {}

function milk_autodiff.new(value, derivative)
    return {value = value, derivative = derivative or 1}
end

function milk_autodiff.sin(d)
    return milk_autodiff.new(
        math.sin(d.value),
        d.derivative * math.cos(d.value)
    )
end

function milk_autodiff.pow(d, n)
    return milk_autodiff.new(
        d.value ^ n,
        n * d.value ^ (n - 1) * d.derivative
    )
end

function milk_autodiff.autodiff(f, x)
    local result = f(milk_autodiff.new(x, 1))
    return result.value, result.derivative
end

-- 使用Milk自动微分库的函数
function f_milk_autodiff(x)
    return milk_autodiff.sin(milk_autodiff.pow(x, 2))
end

-- 性能测试单元
function performance_test()
    local iterations = 10000000
    local start_time

    start_time = os.clock()
    for i = 1, iterations do
        local x = i / 1000.0
        calculus.derivative(f, x)
    end
    local numerical_derivative_time = os.clock() - start_time

    start_time = os.clock()
    for i = 1, iterations do
        local x = i / 1000.0
        calculus.autodiff(f_autodiff, x)
    end
    local auto_derivative_time = os.clock() - start_time

    start_time = os.clock()
    for i = 1, iterations do
        local x = i / 1000.0
        lua_numerical_derivative(f, x)
    end
    local lua_numerical_derivative_time = os.clock() - start_time

    start_time = os.clock()
    for i = 1, iterations do
        local x = i / 1000.0
        milk_autodiff.autodiff(f_milk_autodiff, x)
    end
    local milk_auto_derivative_time = os.clock() - start_time

    print(string.format("Iterations: %d", iterations))
    print(string.format("Calculus numerical derivative performance: %.6f seconds", numerical_derivative_time))
    print(string.format("Calculus auto derivative performance: %.6f seconds", auto_derivative_time))
    print(string.format("Milk numerical derivative performance: %.6f seconds", lua_numerical_derivative_time))
    print(string.format("Milk auto derivative performance: %.6f seconds", milk_auto_derivative_time))
end

print("\nPerformance Test:")
performance_test()
