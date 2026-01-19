# A Beautiful Mess: Reunion

An engine for generating catharsis at 60 FPS.

## Controls

*   **Movement:** Arrow Keys or WASD
*   **Blitz Dash (Overdrive):** Hold `Shift` or `C`
*   **Shooting:** Automatic (Interval based on level)
*   **Pause:** `P` or `Esc`
*   **Memory Fragment Navigation:** `A/D` or Left/Right (while in popup)
*   **Advance/Recover:** `Space` (while in popup)

## Mechanics

1.  **Reunion:** To complete a level, you must trap the **Spectre** near a **Gravity Well** while being in close proximity to her yourself.
2.  **Gravity Susceptibility:** The Spectre actively avoids singularities. Shoot her with bullets to increase her mass and drag her toward the center.
3.  **Ricochets:** Bullets bounce off all walls. Use the environment to land complex shots.
4.  **Shatter:** Bricks and data-structures can be destroyed by your projectiles.
5.  **Toroidal Void:** The world wraps. Leaving the left side brings you back on the right.

## Development

Requires Go 1.21+ and Ebitengine dependencies.

```powershell
go build .
./beautifulmess.exe
```

Or use the development runner:
```powershell
./run.ps1
```