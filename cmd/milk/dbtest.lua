-- 连接到数据库
-- database.open(数据库名, 端口, 用户名, 密码)
local db, err = database.open("mydb", 3306, "username", "password")
if err then
    print("连接数据库失败:", err)
    return
end

-- 检查表是否存在并删除
-- database.query(数据库连接, SQL查询语句)
local check_table_query = "SHOW TABLES LIKE 'users'"
local results = database.query(db, check_table_query)
if results and #results > 0 then
    print("表存在，正在删除")
    -- database.exec(数据库连接, SQL执行语句)
    local drop_table_query = "DROP TABLE IF EXISTS users"
    local success, result = database.exec(db, drop_table_query)
    if not success then
        print("删除表时发生错误:", result)
        return
    end
    print("表删除操作完成")
else
    print("表不存在")
end

-- 创建表
local create_table_query = [[
    CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(50) UNIQUE,
        age INT
    )
]]
-- database.exec(数据库连接, SQL执行语句)
local success, result = database.exec(db, create_table_query)
if not success then
    print("创建表失败:", result)
    return
else
    print("表创建成功")
end

-- 插入数据
-- database.exec(数据库连接, SQL执行语句)
local insert_query = "INSERT INTO users (name, age) VALUES ('Alice', 30)"
local success, result = database.exec(db, insert_query)
if not success then
    print("插入失败:", result)
else
    print("插入成功，最后插入的ID:", result)
end

-- 尝试插入重复的主键
-- database.exec(数据库连接, SQL执行语句)
local duplicate_insert_query = "INSERT INTO users (id, name, age) VALUES (1, 'Bob', 25)"
success, result = database.exec(db, duplicate_insert_query)
if not success then
    print("插入重复主键失败:", result)
else
    print("插入重复主键成功（意外结果）")
end

-- 尝试插入重复的唯一键（name）
-- database.exec(数据库连接, SQL执行语句)
local duplicate_name_query = "INSERT INTO users (name, age) VALUES ('Alice', 35)"
success, result = database.exec(db, duplicate_name_query)
if not success then
    print("插入重复名字失败:", result)
else
    print("插入重复名字成功（意外结果）")
end

-- 更新不存在的记录
-- database.exec(数据库连接, SQL执行语句)
local update_nonexistent_query = "UPDATE users SET age = 40 WHERE name = 'NonExistent'"
success, rows_affected = database.exec(db, update_nonexistent_query)
if success then
    print("更新不存在的记录，受影响的行数:", rows_affected)
else
    print("更新不存在的记录失败:", rows_affected)
end

-- 删除不存在的记录
-- database.exec(数据库连接, SQL执行语句)
local delete_nonexistent_query = "DELETE FROM users WHERE name = 'NonExistent'"
success, rows_affected = database.exec(db, delete_nonexistent_query)
if success then
    print("删除不存在的记录，受影响的行数:", rows_affected)
else
    print("删除不存在的记录失败:", rows_affected)
end

-- 执行复杂查询
-- database.query(数据库连接, SQL查询语句)
local complex_query = [[
    SELECT name, age,
           CASE
               WHEN age < 18 THEN 'Minor'
               WHEN age BETWEEN 18 AND 65 THEN 'Adult'
               ELSE 'Senior'
           END AS age_group
    FROM users
    WHERE age > 20
    ORDER BY age DESC
    LIMIT 5
]]
results = database.query(db, complex_query)
if results then
    if #results > 0 then
        for i, row in ipairs(results) do
            print(string.format("用户 %d: 名字=%s, 年龄=%s, 年龄组=%s", 
                                i, row.name, row.age, row.age_group))
        end
    else
        print("复杂查询没有返回数据")
    end
else
    print("复杂查询执行失败")
end

-- 测试新的快捷方法

-- 使用 createTable 方法创建新表
-- database.createTable(数据库连接, 表名, 列定义表)
print("\n测试 createTable 方法:")
local columns = {
    id = "INT AUTO_INCREMENT PRIMARY KEY",
    name = "VARCHAR(50) UNIQUE",
    email = "VARCHAR(100)",
    age = "INT"
}
success = database.createTable(db, "employees", columns)
if success then
    print("employees 表创建成功")
else
    print("创建 employees 表失败")
end

-- 使用 insert 方法插入数据
-- database.insert(数据库连接, 表名, 数据表)
print("\n测试 insert 方法:")
local employee_data = {
    name = "John Doe",
    email = "john@example.com",
    age = 35
}
success, last_insert_id = database.insert(db, "employees", employee_data)
if success then
    print("插入成功，最后插入的ID:", last_insert_id)
else
    print("插入失败")
end

-- 使用 update 方法更新数据
-- database.update(数据库连接, 表名, 更新数据表, 条件)
print("\n测试 update 方法:")
local update_data = {
    age = 36
}
success, rows_affected = database.update(db, "employees", update_data, "name = 'John Doe'")
if success then
    print("更新成功，受影响的行数:", rows_affected)
else
    print("更新失败")
end

-- 使用 delete 方法删除数据
-- database.delete(数据库连接, 表名, 条件)
print("\n测试 delete 方法:")
success, rows_affected = database.delete(db, "employees", "name = 'John Doe'")
if success then
    print("删除成功，受影响的行数:", rows_affected)
else
    print("删除失败")
end

-- 关闭数据库连接
-- database.close(数据库连接)
local success, err = database.close(db)
if not success then
    print("关闭数据库连接失败:", err)
else
    print("数据库连接已关闭")
end

-- 测试面向对象调用和简单调用

print("\n测试面向对象调用和简单调用:")

-- 重新连接数据库
local db, err = database.open("mydb", 3306, "username", "password")
if err then
    print("重新连接数据库失败:", err)
    return
end

-- 测试 query 方法（面向对象调用）
print("\n测试 query 方法（面向对象调用）:")
local results = db:query("SELECT * FROM users LIMIT 1")
if results and #results > 0 then
    print("查询成功，第一个用户:", results[1].name)
else
    print("查询失败或没有数据")
end

-- 测试 query 方法（简单调用）
print("\n测试 query 方法（简单调用）:")
results = database.query(db, "SELECT * FROM users LIMIT 1")
if results and #results > 0 then
    print("查询成功，第一个用户:", results[1].name)
else
    print("查询失败或没有数据")
end

-- 测试 exec 方法（面向对象调用）
print("\n测试 exec 方法（面向对象调用）:")
local success, result = db:exec("INSERT INTO users (name, age) VALUES ('Bob', 25)")
if success then
    print("插入成功，最后插入的ID:", result)
else
    print("插入失败:", result)
end

-- 测试 exec 方法（简单调用）
print("\n测试 exec 方法（简单调用）:")
success, result = database.exec(db, "INSERT INTO users (name, age) VALUES ('Charlie', 40)")
if success then
    print("插入成功，最后插入的ID:", result)
else
    print("插入失败:", result)
end

-- 测试 createTable 方法（面向对象调用）
print("\n测试 createTable 方法（面向对象调用）:")
local columns = {
    id = "INT AUTO_INCREMENT PRIMARY KEY",
    title = "VARCHAR(100) UNIQUE",
    content = "TEXT"
}
success = db:createTable("posts", columns)
if success then
    print("posts 表创建成功")
else
    print("创建 posts 表失败")
end

-- 测试 createTable 方法（简单调用）
print("\n测试 createTable 方法（简单调用）:")
columns = {
    id = "INT AUTO_INCREMENT PRIMARY KEY",
    name = "VARCHAR(50) UNIQUE",
    description = "TEXT"
}
success = database.createTable(db, "categories", columns)
if success then
    print("categories 表创建成功")
else
    print("创建 categories 表失败")
end

-- 测试 insert 方法（面向对象调用）
print("\n测试 insert 方法（面向对象调用）:")
local post_data = {
    title = "First Post",
    content = "This is the content of the first post."
}
success, last_insert_id = db:insert("posts", post_data)
if success then
    print("插入成功，最后插入的ID:", last_insert_id)
else
    print("插入失败")
end

-- 测试 insert 方法（简单调用）
print("\n测试 insert 方法（简单调用）:")
local category_data = {
    name = "Technology",
    description = "Posts about technology"
}
success, last_insert_id = database.insert(db, "categories", category_data)
if success then
    print("插入成功，最后插入的ID:", last_insert_id)
else
    print("插入失败")
end

-- 测试 update 方法（面向对象调用）
print("\n测试 update 方法（面向对象调用）:")
local update_data = {
    age = 26
}
success, rows_affected = db:update("users", update_data, "name = 'Bob'")
if success then
    print("更新成功，受影响的行数:", rows_affected)
else
    print("更新失败")
end

-- 测试 update 方法（简单调用）
print("\n测试 update 方法（简单调用）:")
update_data = {
    age = 41
}
success, rows_affected = database.update(db, "users", update_data, "name = 'Charlie'")
if success then
    print("更新成功，受影响的行数:", rows_affected)
else
    print("更新失败")
end

-- 测试 delete 方法（面向对象调用）
print("\n测试 delete 方法（面向对象调用）:")
success, rows_affected = db:delete("users", "name = 'Bob'")
if success then
    print("删除成功，受影响的行数:", rows_affected)
else
    print("删除失败")
end

-- 测试 delete 方法（简单调用）
print("\n测试 delete 方法（简单调用）:")
success, rows_affected = database.delete(db, "users", "name = 'Charlie'")
if success then
    print("删除成功，受影响的行数:", rows_affected)
else
    print("删除失败")
end

-- 关闭数据库连接（面向对象调用）
print("\n关闭数据库连接（面向对象调用）:")
success, err = db:close()
if not success then
    print("关闭数据库连接失败:", err)
else
    print("数据库连接已关闭")
end

-- 重新连接数据库以测试简单调用的关闭方法
db, err = database.open("mydb", 3306, "username", "password")
if err then
    print("重新连接数据库失败:", err)
    return
end

-- 关闭数据库连接（简单调用）
print("\n关闭数据库连接（简单调用）:")
success, err = database.close(db)
if not success then
    print("关闭数据库连接失败:", err)
else
    print("数据库连接已关闭")
end
