package systems

import (
	"log"
	"math"

	"beautifulmess/pkg/core"
	"beautifulmess/pkg/level"
	"beautifulmess/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
	lua "github.com/yuin/gopher-lua"
)

// InitLua initializes the shared Lua state and binds API functions
func InitLua(w *world.World) {
	L := w.LState

	// Helper to get ID from first arg
	getID := func(L *lua.LState) core.Entity {
		return core.Entity(L.CheckNumber(1))
	}

	// Expose physics manipulation to allow scripts to drive entities
	L.SetGlobal("apply_force", L.NewFunction(func(L *lua.LState) int {
		id := getID(L)
		if phys, ok := w.Physics[id]; ok {
			phys.Acceleration.X += float64(L.CheckNumber(2))
			phys.Acceleration.Y += float64(L.CheckNumber(3))
		}
		return 0
	}))

	L.SetGlobal("set_max_speed", L.NewFunction(func(L *lua.LState) int {
		id := getID(L)
		if phys, ok := w.Physics[id]; ok {
			phys.MaxSpeed = float64(L.CheckNumber(2))
		}
		return 0
	}))

	L.SetGlobal("rotate", L.NewFunction(func(L *lua.LState) int {
		id := getID(L)
		delta := float64(L.CheckNumber(2))
		if trans, ok := w.Transforms[id]; ok {
			trans.Rotation += delta
		}
		return 0
	}))

	// Expose perception data
	L.SetGlobal("get_self", L.NewFunction(func(L *lua.LState) int {
		id := getID(L)
		trans := w.Transforms[id]
		phys := w.Physics[id]
		if trans != nil && phys != nil {
			L.Push(lua.LNumber(trans.Position.X))
			L.Push(lua.LNumber(trans.Position.Y))
			L.Push(lua.LNumber(phys.Velocity.X))
			L.Push(lua.LNumber(phys.Velocity.Y))
			return 4
		}
		return 0
	}))

	L.SetGlobal("get_vec_to", L.NewFunction(func(L *lua.LState) int {
		id := getID(L)
		tx, ty := float64(L.CheckNumber(2)), float64(L.CheckNumber(3))
		
		trans := w.Transforms[id]
		if trans == nil {
			return 0
		}
		
		delta := core.VecToWrapped(trans.Position, core.Vector2{X: tx, Y: ty})
		dx, dy := delta.X, delta.Y

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
		id := getID(L)
		ai := w.AIs[id]
		if ai == nil {
			return 0
		}
		
		targetID := core.Entity(ai.TargetID)
		if trans, ok := w.Transforms[targetID]; ok {
			L.Push(lua.LNumber(trans.Position.X))
			L.Push(lua.LNumber(trans.Position.Y))
			return 2
		}
		return 0
	}))

	L.SetGlobal("get_input_dir", L.NewFunction(func(L *lua.LState) int {
		// Does not need ID, global input
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

	L.SetGlobal("play_sound", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		// Basic rate limiting could go here if needed, but for now we trust the script
		w.Audio.Play(name)
		return 0
	}))

	// Load scripts as modules/tables
	// We will load them into global tables named after their filename (minus extension)
	scripts := []string{"runner.lua", "spectre.lua"}
	for _, script := range scripts {
		if err := L.DoFile(script); err != nil {
			log.Printf("Failed to load script %s: %v", script, err)
		}
	}
}

func SystemAI(w *world.World, lvl *level.Level) {
	L := w.LState

	for e, ai := range w.AIs {
		if ai == nil || ai.ScriptName == "" { continue }

		// Pre-calculating perception data reduces the burden on the Lua VM and ensures consistent behavior
		bestDist := 99999.0
		var bestWellPos core.Vector2
		foundWell := false
		
		pos, ok := w.Transforms[e]
		if !ok { continue }
		
		for wellID, well := range w.GravityWells {
			if well == nil { continue }
			wellTrans, ok := w.Transforms[wellID]
			if !ok { continue }
			
			// Perceived shortest path calculation respects the toroidal nature of the universe
			delta := core.VecToWrapped(pos.Position, wellTrans.Position)
			d := math.Sqrt(delta.X*delta.X + delta.Y*delta.Y)
			if d < bestDist {
				bestDist, bestWellPos, foundWell = d, wellTrans.Position, true
			}
		}

		wellX, wellY := 0.0, 0.0
		if foundWell {
			wellX, wellY = bestWellPos.X, bestWellPos.Y
		}

		// Delegating decision-making to hot-reloadable scripts enables rapid gameplay balancing
		tableName := getScriptName(ai.ScriptName)
		
		tbl := L.GetGlobal(tableName)
		if tbl.Type() == lua.LTTable {
			fn := L.GetField(tbl, "update_state")
			if fn.Type() == lua.LTFunction {
				L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, 
					lua.LNumber(e),
					lua.LNumber(lvl.Memory.Position.X),
					lua.LNumber(lvl.Memory.Position.Y),
					lua.LNumber(core.MemoryRadius),
					lua.LNumber(wellX),
					lua.LNumber(wellY),
				)
			}
		}
	}
}

func getScriptName(name string) string {
	if len(name) > 4 && name[len(name)-4:] == ".lua" {
		return name[:len(name)-4]
	}
	return name
}

