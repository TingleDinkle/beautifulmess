package systems

import (
	"fmt"
	"log"
	"math"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/level"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
	lua "github.com/yuin/gopher-lua"
)

func SystemAI(w *world.World, e core.Entity, lvl *level.Level) {
	ai := w.AIs[e]
	if ai == nil || ai.LState == nil {
		return
	}
	L := ai.LState

	// Feed the AI the current world state
	L.SetGlobal("mem_x", lua.LNumber(lvl.Memory.Position.X))
	L.SetGlobal("mem_y", lua.LNumber(lvl.Memory.Position.Y))
	L.SetGlobal("mem_radius", lua.LNumber(core.MemoryRadius))

	// Help AI find the nearest gravity well
	bestDist := 99999.0
	var bestWell level.GravityWell
	pos := w.Transforms[e].Position
	for _, well := range lvl.Wells {
		// Use wrapped distance logic so AI doesn't get confused by edges
		dx := well.Position.X - pos.X
		dy := well.Position.Y - pos.Y
		if dx > core.ScreenWidth/2 {
			dx -= core.ScreenWidth
		}
		if dx < -core.ScreenWidth/2 {
			dx += core.ScreenWidth
		}
		if dy > core.ScreenHeight/2 {
			dy -= core.ScreenHeight
		}
		if dy < -core.ScreenHeight/2 {
			dy += core.ScreenHeight
		}

		d := math.Sqrt(dx*dx + dy*dy)
		if d < bestDist {
			bestDist = d
			bestWell = well
		}
	}
	L.SetGlobal("well_x", lua.LNumber(bestWell.Position.X))
	L.SetGlobal("well_y", lua.LNumber(bestWell.Position.Y))

	// Let Lua take the wheel
	if err := L.CallByParam(lua.P{Fn: L.GetGlobal("update_state"), NRet: 0, Protect: true}); err != nil {
		fmt.Printf("Lua Error (Brain Freeze): %v\n", err)
	}
}

// BindLua connects Go functions to Lua scripts
func BindLua(w *world.World, e core.Entity, targetID core.Entity) {
	L := lua.NewState()
	w.AIs[e].LState = L

	// Physics Hooks
	L.SetGlobal("apply_force", L.NewFunction(func(L *lua.LState) int {
		w.Physics[e].Acceleration.X += float64(L.CheckNumber(1))
		w.Physics[e].Acceleration.Y += float64(L.CheckNumber(2))
		return 0
	}))

	L.SetGlobal("set_max_speed", L.NewFunction(func(L *lua.LState) int {
		w.Physics[e].MaxSpeed = float64(L.CheckNumber(1))
		return 0
	}))

	// Sensor Hooks
	L.SetGlobal("get_self", L.NewFunction(func(L *lua.LState) int {
		p := w.Transforms[e].Position
		v := w.Physics[e].Velocity
		L.Push(lua.LNumber(p.X))
		L.Push(lua.LNumber(p.Y))
		L.Push(lua.LNumber(v.X))
		L.Push(lua.LNumber(v.Y))
		return 4
	}))

	L.SetGlobal("get_vec_to", L.NewFunction(func(L *lua.LState) int {
		tx, ty := float64(L.CheckNumber(1)), float64(L.CheckNumber(2))
		mx, my := w.Transforms[e].Position.X, w.Transforms[e].Position.Y
		dx, dy := tx-mx, ty-my
		if dx > core.ScreenWidth/2 {
			dx -= core.ScreenWidth
		}
		if dx < -core.ScreenWidth/2 {
			dx += core.ScreenWidth
		}
		if dy > core.ScreenHeight/2 {
			dy -= core.ScreenHeight
		}
		if dy < -core.ScreenHeight/2 {
			dy += core.ScreenHeight
		}

		d := math.Sqrt(dx*dx + dy*dy)
		if d < 0.01 {
			d = 0.01
		}
		L.Push(lua.LNumber(dx / d))
		L.Push(lua.LNumber(dy / d))
		L.Push(lua.LNumber(d))
		return 3
	}))

	L.SetGlobal("get_target", L.NewFunction(func(L *lua.LState) int {
		p := w.Transforms[targetID].Position
		L.Push(lua.LNumber(p.X))
		L.Push(lua.LNumber(p.Y))
		return 2
	}))

	L.SetGlobal("get_input_dir", L.NewFunction(func(L *lua.LState) int {
		dx, dy := 0.0, 0.0
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			dx = -1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			dx = 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
			dy = -1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
			dy = 1
		}
		L.Push(lua.LNumber(dx))
		L.Push(lua.LNumber(dy))
		return 2
	}))

	L.SetGlobal("cast_ray", L.NewFunction(func(L *lua.LState) int {
		// In the void, raycasts see infinity.
		dist := float64(L.CheckNumber(2))
		L.Push(lua.LNumber(dist))
		return 1
	}))

	if err := L.DoFile(w.AIs[e].ScriptPath); err != nil {
		log.Printf("Lua Script Error: %v", err)
	}
}
