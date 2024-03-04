-- 具体业务
local key = KEYS[1]
-- 区分 阅读数 点赞数 收藏数
local cntKey = ARGV[1]

local delta = tonumber(ARGV[2])

local exist=redis.call("EXISTS", key)

-- If key does not exist,
-- a new key holding a hash is created.
-- If field does not exist
-- the value is set to 0 before the operation is performed.
if exist == 1 then
    redis.call("HINCRBY", key, cntKey, delta)
    return 1
else
    return 0
end