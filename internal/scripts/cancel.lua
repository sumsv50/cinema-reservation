-- cancel.lua
-- Cancel reserved seats by removing them from the Redis set

-- KEYS[1] = key where reserved seats are stored (SET)
-- ARGV = list of seat coordinates like "1,2"

for i = 1, #ARGV do
  redis.call("HDEL", KEYS[1], ARGV[i])
end

return "OK"
