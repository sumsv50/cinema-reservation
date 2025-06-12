-- KEYS[1] = Redis hash key (cinema:{cinemaID}:seats)
-- ARGV[1] = Manhattan distance
-- ARGV[2..] = seat list: "row:col"

local min_dist = tonumber(ARGV[1])
local requested_seats = {}

for i = 2, #ARGV do
    local coord = ARGV[i]
    table.insert(requested_seats, coord)
end

-- Check if any seat is already taken
for _, seat in ipairs(requested_seats) do
    if redis.call("HEXISTS", KEYS[1], seat) == 1 then
        return {err="[SEATS_RESERVED] Seat already reserved: " .. seat}
    end
end

-- Check social distancing
for _, seat1 in ipairs(requested_seats) do
    local row1, col1 = seat1:match("^(%d+):(%d+)$")
    row1 = tonumber(row1)
    col1 = tonumber(col1)

    local keys = redis.call("HKEYS", KEYS[1])
    for _, seat2 in ipairs(keys) do
        local row2, col2 = seat2:match("^(%d+):(%d+)$")
        row2 = tonumber(row2)
        col2 = tonumber(col2)
        local dist = math.abs(row1 - row2) + math.abs(col1 - col2)
        if dist < min_dist then
            return {err="[MIN_DISTANCE_VIOLATION]Social distancing violated near: " .. seat2}
        end
    end
end

-- All checks passed, reserve the seats
for _, seat in ipairs(requested_seats) do
    redis.call("HSET", KEYS[1], seat, "1")
end

return "OK"
