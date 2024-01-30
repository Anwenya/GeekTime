-- 滑动窗口限流实现
-- 使用有序集合记录每个ip所有请求的访问时间 并设置过期时间为窗口大小
-- 验证规则:
-- 当前请求访问时间 - 窗口大小 = 需要计数的时间段
-- 计算剩余的元素数量 和 阈值 进行比较

-- 参数 KEYS[1] 限流对象:具体的redis中的key 如ip:127.0.0.1
-- 参数 ARGV[1] 窗口大小:也就是时间范围 如5秒
-- 参数 ARGV[2] 执行该命令时的时间
-- 参数 ARGV[3] 窗口的起始时间:用于排除当前窗口外的记录

local key = KEYS[1]
local window = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local min = now - window

-- 调用 Redis 的 ZREMRANGEBYSCORE 命令，
-- 从有序集合 key 中移除分数介于 -inf （负无穷）和 min 之间的所有元素。
-- 这一步相当于将有序集合中小于等于 min 分数的元素全部删除。
redis.call("ZREMRANGEBYSCORE", key, "-inf", min)

-- 计算该key所有的元素数量
local cnt = redis.call("ZCOUNT", key, "-inf", "+inf")

-- 大于阈值则执行限流操作
if cnt >= threshold then
    -- 执行限流
    return "true"
else
    -- 没有超过阈值则把当前记录
    -- 添加一个元素把 score 和 member 都设置成 now
    redis.call("ZADD", key, now, now)
    -- 重新设置过期时间
    redis.call("PEXPIRE", key, window)
    return "false"
end