spectre = {}

-- Utility-based state machine constants
local STATE_CRUISE = 0
local STATE_SPRINT = 1
local STATE_JINK   = 2
local STATE_RECOVER = 3

local current_state = STATE_CRUISE
local state_timer = 0
local stamina = 100.0
local max_stamina = 100.0
local jink_dir = 1

function spectre.update_state(id, mem_x, mem_y, mem_radius, well_x, well_y)
    local _, _, my_vx, my_vy = get_self(id)
    local opp_x, opp_y = get_target(id)
    local to_opp_x, to_opp_y, dist = get_vec_to(id, opp_x, opp_y)
    
    state_timer = state_timer - 1
    
    -- Stamina regeneration prevents infinite sprinting and encourages tactical retreats
    if current_state ~= STATE_SPRINT then stamina = math.min(max_stamina, stamina + 0.5) end

    -- Threat-response logic triggers evasion when the runner enters the spectre's personal space
    if dist < 120 and current_state == STATE_CRUISE then
        if stamina > 30 then
            -- Active counter-force maneuvers prevent the player from easily maintaining contact
            current_state = STATE_JINK; state_timer = 15; jink_dir = (math.random()<0.5) and 1 or -1
            play_sound("spectre_dash")
        else
            current_state = STATE_RECOVER; state_timer = 40
        end
    end
    
    -- State transitions are timer-based to ensure rhythmic movement cycles
    if state_timer <= 0 then
        if current_state == STATE_SPRINT then current_state = STATE_RECOVER; state_timer = 30
        elseif current_state == STATE_JINK then current_state = STATE_SPRINT; state_timer = 40
        elseif current_state == STATE_RECOVER then current_state = STATE_CRUISE end
    end
    
    -- Resistance forces near memory nodes simulate the narrative 'struggle' against re-assimilation
    local to_mem_x, to_mem_y, mem_dist = get_vec_to(id, mem_x, mem_y)
    if mem_dist < mem_radius then
        -- Fleeing the singularity represents the spectre's vestigial survival instinct
        apply_force(id, -to_mem_x * 2.5, -to_mem_y * 2.5)
        apply_force(id, (math.random()-0.5)*4, (math.random()-0.5)*4)
        return
    end

    local fx, fy = 0, 0
    
    if current_state == STATE_CRUISE then
        set_max_speed(id, 4.0)
        fx, fy = -to_opp_x * 0.5, -to_opp_y * 0.5
        
        -- Singularity avoidance simulates active engine compensation against gravitational pull
        local to_well_x, to_well_y, well_dist = get_vec_to(id, well_x, well_y)
        if well_dist < 400 then
             -- Applying a counter-force away from the well increases the effort required to trap the entity
             fx, fy = fx - (to_well_x * 1.2), fy - (to_well_y * 1.2)
        end
        
    elseif current_state == STATE_SPRINT then
        set_max_speed(id, 9.0)
        stamina = stamina - 2.0
        fx, fy = -to_opp_x * 2.0, -to_opp_y * 2.0
        if stamina <= 0 then current_state = STATE_RECOVER; state_timer = 60 end
        
    elseif current_state == STATE_JINK then
        set_max_speed(id, 12.0)
        stamina = stamina - 1.0
        -- Perpendicular vectors create lateral movement to break target locks
        fx, fy = -to_opp_y * jink_dir * 3.0, to_opp_x * jink_dir * 3.0
        
    elseif current_state == STATE_RECOVER then
        set_max_speed(id, 3.0)
        fx, fy = -to_opp_x * 0.8, -to_opp_y * 0.8
    end
    
    -- Velocity damping prevents infinite drifting in the void
    if math.abs(fx) < 0.1 and math.abs(fy) < 0.1 then
        fx, fy = my_vx * 0.1, my_vy * 0.1
    end

    apply_force(id, fx, fy)
end