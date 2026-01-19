# THE ARCHITECTURE OF A BEAUTIFUL MESS (V2: REUNION)

**To Future Me (and anyone else lost in the void):**

This project has evolved. It is no longer just a tragedy engine; it is a machine for finding common ground. We have merged the "Rigid Physics" of the void with the "Astro-Style" kinetic energy of arcade combat.

Here is the updated philosophy. Respect the polish.

---

## 1. The Core Philosophy: "Precision Logic, Explosive Emotion"

We still use a strict separation of concerns, but the boundaries are sharper.

*   **Go (The Laws of Physics):** The Universe. Now optimized at the hardware level. It handles $O(N)$ linear updates via Active Lists and $O(1)$ collisions via a Spatial Hash Grid. It is unyielding and blazing fast.
*   **Lua (The Ghosts):** The AI. The `Spectre` now actively resists the void. She doesn't just orbit; she fights to stay free until you increase her mass enough to pull her home.
*   **The Juice (The Vibe):** Impact matters. Every shot has recoil, every hit has a stop, and every reunion has a bloom. We paint with math, but we also paint with particles.

## 2. The Tech Stack (The "Nintendo-Grade" Overhaul)

*   **High-Performance ECS:**
    *   We moved from Maps to **Slices of Pointers**. This gives us $O(1)$ indexing while keeping the safety of Go's `nil` checks.
    *   **Active List Tracking:** We only iterate over entities that are alive (`ActiveEntities`). We don't waste CPU cycles on the empty void.
    *   **Spatial Hash Grid:** A 100px grid buckets static geometry. Collision cost is now constant, regardless of map density.

*   **The Destruction Engine (Shatter):**
    *   Destruction is an event. The `shatterEntity` logic uses a dual-layer particle burst with **inherited momentum** and **quirky behaviors** (Orbiting fragments, Flickering sparks).
    *   **Universal Ricochets:** Bullets are high-energy data packets. They bounce off every surface, allowing for complex multi-hit trick shots.

*   **The Audio Pipeline (Polyphonic Pooling):**
    *   We don't cut off sounds. We use a **Player Recycling Pool** (8 channels per sample). Overlapping explosions and blitz-dashes create a rich, cinematic soundscape.

## 3. The Mechanics (How the game feels)

### The Blitz Dash (You)
*   **Feel:** Supersonic, weighted, physical.
*   **Logic:** Holding `Shift` or `C` triggers **Speed Overdrive**. Your `MaxSpeed` doubles, and your thrusters emit heavy particles. The sub-bass "Blitz" thump reinforces the displacement.

### The Escape Artist (Her)
*   **Feel:** Defensive, reactive, intelligent.
*   **The Brain:** She monitors your proximity and the nearest singularity. She will actively apply counter-forces to flee the gravity well until you shoot her to increase her susceptibility.

### The Reunion (The Goal)
*   **The Transition:** When you catch her near a well, the level doesn't just end. It **blooms**. The singularities turn into beacons, and a "Sunshine Pixelizing" effect sweeps the grid, signifying that common ground has been found.

## 4. The Rules of Engagement

1.  **Recoil is a Penalty:** Every shot pushes you back. Blind firing will kill your momentum. Timing is everything.
2.  **Gravity is a Tool:** Use the singularities to slingshot yourself or trap her. 
3.  **Memory Fragments:** Narrative popups now support **multi-photo pagination**. Use A/D to explore the fragmented history of the union.

## 5. How to Expand This

*   **Adding a Level:** Go to `InitLevels` in `pkg/level/level.go`. Use `genGrid` or `genLine` to build new structures.
*   **Adding Sound:** Put a `.wav` in `assets/`, call `LoadFile` in `main.go`, and use `w.Audio.Play` anywhere. The pooler handles the rest.
*   **Adding Juice:** Every physical interaction should trigger `w.ScreenShake`. It's the primary way the game speaks to the player.

---

**Final Note:**
The code is optimized so the emotion can be loud. Keep the loops tight and the transitions beautiful. Don't let technical debt clutter the soul of the mess.
