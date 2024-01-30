-- 令牌桶限流
-- 能解滑动窗口中 单位时间内用光次数的情况

-- 参数 KEYS[1] 限流对象:具体的redis中的key 如ip:127.0.0.1
-- 参数 ARGV[1] 桶中最大令牌数
-- 参数 ARGV[2] 每秒生成的令牌数, 比如每秒生成10个就是10
-- 参数 ARGV[3] 执行该命令时的时间

local max_token = tonumber(ARGV[1])
local token_rate = tonumber(ARGV[2])
local current_time = tonumber(ARGV[3])

-- 用于存储 此次请求-上一次成功的请求=过去的时间
local past_time = 0
-- 每毫秒生产令牌速率 转换为毫秒控制粒度更细
local ratePerMills = token_rate/1000

-- 先查询之前是否有请求成功的记录
-- 获取哈希数据结构中 "last_time" 和 "stored_token_nums"
local info = redis.pcall("HMGET", KEYS[1], "last_time", "stored_token_nums")
-- 最后一次通过限流的时间
local last_time = info[1]
-- 剩余的令牌数量
local stored_token_nums = tonumber(info[2])

if stored_token_nums == nil then
    -- 第一次请求或者键已经过期
    -- 令牌恢复至最大数量
    stored_token_nums = max_token
    -- 记录请求时间
    last_time = current_time
else
    -- 处于流量中
    -- 经过了多少时间
    past_time = current_time - last_time

    if past_time <= 0 then
        -- 高并发下每个服务的时间可能不一致
        -- 强制变成0 此处可能会出现少量误差
        past_time = 0
    end
    -- 两次请求期间内应该生成多少个token
    -- 向下取整 多余的认为还没生成完
    local generated_nums = math.floor(past_time * ratePerMills)
    -- 合并所有的令牌后不能超过设定的最大令牌数
    stored_token_nums = math.min((stored_token_nums + generated_nums), max_token)
end

-- 返回值
local returnVal = "true"

-- 通过限流
if stored_token_nums > 0 then
    returnVal = "false"
    -- 减少令牌
    stored_token_nums = stored_token_nums - 1
    -- 必须要在获得令牌后才能重新记录时间
    -- 举例:当每隔2ms请求一次时 只要第一次没有获取到token 那么后续会无法生产token 永远只过去了2ms
    last_time = last_time + past_time

    -- 更新缓存
    redis.call("HMSET", KEYS[1], "last_time", last_time, "stored_token_nums", stored_token_nums)
    -- 设置过期时间
    -- 令牌桶满额的时间 = 空缺的令牌数 * 生成一枚令牌所需要的毫秒数
    redis.call("PEXPIRE", KEYS[1], math.ceil((1/ratePerMills) * (max_token - stored_token_nums)))
end

return returnVal