spectre = {}

-- Logic assumes wrapped coordinates

local STATE_CRUISE = 0
local STATE_SPRINT = 1
local STATE_JINK   = 2
local STATE_RECOVER = 3

-- State needs to be per-entity if we have multiple spectres. 
-- For now, with one spectre, this local state is "okay" but ideally should be stored in Go or a Lua table keyed by ID.
-- Let's keep it simple for this refactor as there is only one spectre.
local current_state = STATE_CRUISE
local state_timer = 0
local stamina = 100.0
local max_stamina = 100.0
local jink_dir = 1

function spectre.update_state(id, mem_x, mem_y, mem_radius, well_x, well_y)
    local my_x, my_y, my_vx, my_vy = get_self(id)
    local opp_x, opp_y = get_target(id)
    
    -- get_vec_to returns the shortest path across the screen wrap
    local to_opp_x, to_opp_y, dist = get_vec_to(id, opp_x, opp_y)
    
    state_timer = state_timer - 1
    if current_state ~= STATE_SPRINT then stamina = stamina + 0.5 end
    if stamina > max_stamina then stamina = max_stamina end
    
    if dist < 150 and current_state == STATE_CRUISE then
        if stamina > 30 then
            if math.random() < 0.4 then
                current_state = STATE_JINK; state_timer = 20; jink_dir = (math.random()<0.5) and 1 or -1
            else
                current_state = STATE_SPRINT; state_timer = 60
            end
        else
            current_state = STATE_RECOVER; state_timer = 40
        end
    end
    
    if state_timer <= 0 then
        if current_state == STATE_SPRINT then current_state = STATE_RECOVER; state_timer = 30
        elseif current_state == STATE_JINK then current_state = STATE_SPRINT; state_timer = 40
        elseif current_state == STATE_RECOVER then current_state = STATE_CRUISE end
    end
    
    -- Apply erratic force when near the memory node to simulate struggle
    local to_mem_x, to_mem_y, mem_dist = get_vec_to(id, mem_x, mem_y)
    if mem_dist < mem_radius then
        local pull = 1.5
        apply_force(id, to_mem_x * pull, to_mem_y * pull)
        apply_force(id, (math.random()-0.5)*2, (math.random()-0.5)*2)
        return
    end

    local fx, fy = 0, 0
    
    if current_state == STATE_CRUISE then
        set_max_speed(id, 4.0)
        fx = -to_opp_x * 0.5
        fy = -to_opp_y * 0.5
        
        -- Utilize gravity assist for speed boost when near wells
        local to_well_x, to_well_y, well_dist = get_vec_to(id, well_x, well_y)
        if well_dist < 300 and well_dist > 50 then
             -- Spiral IN slightly to gain speed
             fx = fx + (to_well_x * 0.4)
             fy = fy + (to_well_y * 0.4)
        end
        
    elseif current_state == STATE_SPRINT then
        set_max_speed(id, 9.0)
        stamina = stamina - 2.0
        fx = -to_opp_x * 2.0
        fy = -to_opp_y * 2.0
        if stamina <= 0 then current_state = STATE_RECOVER; state_timer = 60 end
        
    elseif current_state == STATE_JINK then
        set_max_speed(id, 12.0)
        stamina = stamina - 1.0
        local perp_x = -to_opp_y * jink_dir
        local perp_y = to_opp_x * jink_dir
        fx = perp_x * 3.0
        fy = perp_y * 3.0
        
    elseif current_state == STATE_RECOVER then
        set_max_speed(id, 3.0)
        fx = -to_opp_x * 0.8
        fy = -to_opp_y * 0.8
    end
    
    -- Dampen velocity when idle to prevent floating away forever
    if math.abs(fx) < 0.1 and math.abs(fy) < 0.1 then
        fx = my_vx * 0.1
        fy = my_vy * 0.1
    end

    apply_force(id, fx, fy)
end