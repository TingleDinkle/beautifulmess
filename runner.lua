runner = {}

function runner.update_state(id, mem_x, mem_y, mem_radius, well_x, well_y)
    local idx, idy = get_input_dir() 
    
    -- Tank Controls: Left/Right rotates
    local ang_vel = 0.1
    
    if idx < 0 then
        rotate(id, -ang_vel)
    elseif idx > 0 then
        rotate(id, ang_vel)
    end
    
    -- Physics handles forward movement automatically.
end