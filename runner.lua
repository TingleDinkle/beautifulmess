runner = {}

function runner.update_state(id, mem_x, mem_y, mem_radius, well_x, well_y)
    local idx, idy = get_input_dir() -- Input is global, no ID needed
    
    -- Use high acceleration/friction for tight controls ("Snap" physics)
    local accel = 1.5
    
    if idx ~= 0 or idy ~= 0 then
        apply_force(id, idx * accel, idy * accel)
    else
        -- Active braking (logic handled by friction in Go, but we can boost it)
    end
    
    -- Randomly trigger dash to simulate glitchy behavior
    if math.random() < 0.02 then
        apply_force(id, idx * 10, idy * 10)
    end
end