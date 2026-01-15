-- spectre.lua (The Infinite Dancer)
-- Optimized for Toroidal Space

local STATE_CRUISE = 0
local STATE_SPRINT = 1
local STATE_JINK   = 2
local STATE_RECOVER = 3

local current_state = STATE_CRUISE
local state_timer = 0
local stamina = 100.0
local max_stamina = 100.0
local jink_dir = 1

function update_state()
    local my_x, my_y, my_vx, my_vy = get_self()
    local opp_x, opp_y = get_target()
    
    -- TOROIDAL AWARENESS
    -- get_vec_to now returns the shortest path across the screen wrap
    local to_opp_x, to_opp_y, dist = get_vec_to(opp_x, opp_y)
    
    state_timer = state_timer - 1
    if current_state ~= STATE_SPRINT then stamina = stamina + 0.5 end
    if stamina > max_stamina then stamina = max_stamina end
    
    -- Logic: Fleeing
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
    
    -- TRAP LOGIC
    local to_mem_x, to_mem_y, mem_dist = get_vec_to(mem_x, mem_y)
    if mem_dist < mem_radius then
        local pull = 1.5
        apply_force(to_mem_x * pull, to_mem_y * pull)
        apply_force((math.random()-0.5)*2, (math.random()-0.5)*2)
        return
    end

    local fx, fy = 0, 0
    
    if current_state == STATE_CRUISE then
        set_max_speed(4.0)
        fx = -to_opp_x * 0.5
        fy = -to_opp_y * 0.5
        
        -- Gravity Well Surfing
        local to_well_x, to_well_y, well_dist = get_vec_to(well_x, well_y)
        if well_dist < 300 and well_dist > 50 then
             -- Spiral IN slightly to gain speed
             fx = fx + (to_well_x * 0.4)
             fy = fy + (to_well_y * 0.4)
        end
        
    elseif current_state == STATE_SPRINT then
        set_max_speed(9.0)
        stamina = stamina - 2.0
        fx = -to_opp_x * 2.0
        fy = -to_opp_y * 2.0
        if stamina <= 0 then current_state = STATE_RECOVER; state_timer = 60 end
        
    elseif current_state == STATE_JINK then
        set_max_speed(12.0)
        stamina = stamina - 1.0
        local perp_x = -to_opp_y * jink_dir
        local perp_y = to_opp_x * jink_dir
        fx = perp_x * 3.0
        fy = perp_y * 3.0
        
    elseif current_state == STATE_RECOVER then
        set_max_speed(3.0)
        fx = -to_opp_x * 0.8
        fy = -to_opp_y * 0.8
    end
    
    -- INFINITE SPACE DRIFT
    -- Since there are no walls, we just maintain current momentum if forces are low
    -- to prevent jittering.
    if math.abs(fx) < 0.1 and math.abs(fy) < 0.1 then
        fx = my_vx * 0.1
        fy = my_vy * 0.1
    end

    apply_force(fx, fy)
end