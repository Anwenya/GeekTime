local key = KEYS[1]
local cntKey = key..":cnt"
-- 用户输入的验证码
local expectedCode = ARGV[1]
-- 该验证码的可验证次数
-- 不存在时tonumber会返回nil
local cnt = tonumber(redis.call("get", cntKey))
-- 取出验证码
local code = redis.call("get", key)

-- 验证次数耗尽了 或者 已经过期了
if cnt == nil or cnt <= 0 then
    return -1
end

-- 验证码相等 设置可验证次数为0
if code == expectedCode then
    redis.call("set", cntKey, 0)
    return 0
-- 不相等 用户输错了
else
    redis.call("decr", cntKey)
    return -2
end