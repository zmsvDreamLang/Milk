-- 测试统计库函数

-- 简单调用方式测试
print("简单调用方式测试:")

-- 测试基本功能
print("基本功能测试:")
print("均值:", statistic.mean(1, 2, 3, 4, 5))
print("中位数:", statistic.median(1, 3, 2, 5, 4))
print("众数:", statistic.mode(1, 2, 2, 3, 3, 3, 4, 4, 5))
print("标准差:", statistic.stddev(1, 2, 3, 4, 5))
print("方差:", statistic.variance(1, 2, 3, 4, 5))

-- 使用表调用方式测试
print("\n使用表调用方式测试:")
print("均值:", statistic.mean({1, 2, 3, 4, 5}))
print("中位数:", statistic.median({1, 3, 2, 5, 4}))
print("众数:", statistic.mode({1, 2, 2, 3, 3, 3, 4, 4, 5}))
print("标准差:", statistic.stddev({1, 2, 3, 4, 5}))
print("方差:", statistic.variance({1, 2, 3, 4, 5}))

-- 测试边界情况
print("\n边界情况测试:")
print("空集合均值:", statistic.mean({}))
print("单值中位数:", statistic.median({5}))
print("全相同值众数:", statistic.mode({3, 3, 3, 3}))
print("零方差:", statistic.variance({1, 1, 1, 1}))

-- 测试大数据集
print("\n大数据集测试:")
local large_dataset = {}
for i = 1, 1000 do
    large_dataset[i] = i
end
print("1到1000的均值:", statistic.mean(large_dataset))
print("1到1000的中位数:", statistic.median(large_dataset))

-- 面向对象调用方式测试
print("\n面向对象调用方式测试:")

-- 使用索引添加数据
local stats1 = statistic.new()
for i = 1, 5 do
    stats1[i] = i
end

print("使用索引添加数据:")
print("均值:", stats1:mean())
print("中位数:", stats1:median())
print("众数:", stats1:mode())
print("标准差:", stats1:stddev())
print("方差:", stats1:variance())

-- 使用add方法添加数据
local stats2 = statistic.new()
for i = 1, 5 do
    stats2:add(i)
end

print("\n使用add方法添加数据:")
print("均值:", stats2:mean())
print("中位数:", stats2:median())
print("众数:", stats2:mode())
print("标准差:", stats2:stddev())
print("方差:", stats2:variance())

-- 测试混合数据类型
print("\n混合数据类型测试:")
local mixed_stats = statistic.new()
mixed_stats:add(1)
mixed_stats:add(2.5)
mixed_stats:add(3.14)
mixed_stats:add(4)
mixed_stats:add(5.9)

print("均值:", mixed_stats:mean())
print("中位数:", mixed_stats:median())
print("标准差:", mixed_stats:stddev())

-- 测试异常情况处理
print("\n异常情况处理测试:")
print("无效输入 (字符串):", pcall(statistic.mean, "not a number"))
print("无效输入 (布尔值):", pcall(statistic.median, true, false))
print("无效输入 (nil):", pcall(statistic.mode, nil))

-- 性能测试
print("\n性能测试:")

-- 生成随机数据的辅助函数
local function generate_random_data(length)
    local data = {}
    for i = 1, length do
        data[i] = math.random(1, 100)
    end
    return data
end

local function run_test(name, func, args, iterations)
    local start_time = os.clock()
    for i = 1, iterations do
        func(table.unpack(args))
    end
    print(string.format("计算 %s %d 次所需时间: %.4f 秒", name, iterations, os.clock() - start_time))
end

-- 固定数据测试
local fixed_data = {1, 2, 3, 4, 5}
run_test("均值 (固定数据)", statistic.mean, fixed_data, 10000000)
run_test("中位数 (固定数据)", statistic.median, fixed_data, 10000000)
run_test("众数 (固定数据)", statistic.mode, fixed_data, 10000000)
run_test("标准差 (固定数据)", statistic.stddev, fixed_data, 10000000)
run_test("方差 (固定数据)", statistic.variance, fixed_data, 10000000)

-- 动态数据测试
local iterations = 1000000
for length = 10, 50, 10 do
    print(string.format("\n测试随机数据长度为 %d", length))
    local random_data = generate_random_data(length)
    run_test("均值 (随机数据)", statistic.mean, random_data, iterations)
    run_test("中位数 (随机数据)", statistic.median, random_data, iterations)
    run_test("众数 (随机数据)", statistic.mode, random_data, iterations)
    run_test("标准差 (随机数据)", statistic.stddev, random_data, iterations)
    run_test("方差 (随机数据)", statistic.variance, random_data, iterations)
end

-- 面向对象调用测试
local stats = statistic.new()
for i = 1, 5 do
    stats:add(i)
end

print("\n测试面向对象调用")
run_test("均值 (面向对象)", function() return stats:mean() end, {}, 10000000)
run_test("中位数 (面向对象)", function() return stats:median() end, {}, 10000000)
run_test("众数 (面向对象)", function() return stats:mode() end, {}, 10000000)
run_test("标准差 (面向对象)", function() return stats:stddev() end, {}, 10000000)
run_test("方差 (面向对象)", function() return stats:variance() end, {}, 10000000)
