-- runner.lua (High Performance)

function update_state()
    local idx, idy = get_input_dir()
    
    -- "Snap" physics: High acceleration, High friction
    local accel = 1.5
    
    if idx ~= 0 or idy ~= 0 then
        apply_force(idx * accel, idy * accel)
    else
        -- Active braking (logic handled by friction in Go, but we can boost it)
    end
    
    -- Dash (Shift key equivalent logic, random glitch for now)
    if math.random() < 0.02 then
        apply_force(idx * 10, idy * 10)
    end
end