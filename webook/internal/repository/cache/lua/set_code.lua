-- 验证码发送间隔是1分钟
-- 有效期是10分钟

-- 验证码的key
local key = KEYS[1]
-- 验证码的可验证次数的key
local cntKey = key..":cnt"
-- 准备存储的验证码
local val = ARGV[1]

-- 查询过期时间
-- key不存在返回-2
-- key没有设置过期时间返回-1
local ttl = tonumber(redis.call("ttl", key))

-- 存在异常的key
if ttl == -1 then
    return -2
-- 可以发送
elseif ttl == -2 or ttl < 540 then
    -- 设置可验证次数和过期时间
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire", cntKey, 600)
    return 0
-- 发送太频繁 间隔小于1分钟
else
    return -1
end