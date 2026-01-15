# THE ARCHITECTURE OF A BEAUTIFUL MESS

**To Future Me (and anyone else lost in the void):**

This project is not just a game; it is an engine for generating tragedy and catharsis at 60 FPS. If you are reading this, you are probably trying to fix a bug or add a new feature.

Here is the philosophy. Don't break the magic.

---

## 1. The Core Philosophy: "Rigid World, Fluid Soul"

We use a strict separation of concerns because feelings are messy, but code shouldn't be.

*   **Go (The Laws of Physics):** The Universe. It is unyielding. It handles Gravity, Collision, and Time. It doesn't care if you are sad; it only cares about `F = ma`.
*   **Lua (The Ghosts):** The AI. It is fluid, hot-reloadable, and messy. The `Spectre` lives here. She calculates her next move based on fear, stamina, and orbital mechanics.
*   **Shaders (The Vibe):** The Mood. We don't draw pixels; we paint math. The background is a fragment shader because textures are too static for this universe.

## 2. The Tech Stack (Why we chose this)

*   **ECS (Entity Component System):**
    *   We don't use Object-Oriented hierarchies (`class Player extends Entity`). That path leads to spaghetti hell.
    *   We use **Data-Driven Design**. A "Black Hole" is just an Entity with a `Position` and a `GravityWell` component.
    *   *Benefit:* You want to make the Runner have gravity? Just add the component. You want the Spectre to leave a frost trail? Add the `Render` tag. It is infinitely expandable.

*   **Toroidal Topology:**
    *   The world wraps. `x = -10` is actually `x = ScreenWidth - 10`.
    *   *Why?* Because corners are boring. In a relationship, you can't just run away; you eventually end up back where you started.

## 3. The Entities (How they think)

### The Runner (You)
*   **Archetype:** *Tokyo-Spliff Cyberpunk.*
*   **Feel:** Twitchy, high-acceleration, "glitch" movement.
*   **Code:** You are the only thing that breaks the rules (using `Shift` to overdrive physics).

### The Spectre (Her)
*   **Archetype:** *Gothic Spirit.*
*   **Feel:** Heavy, inertial, flowing. She doesn't turn instantly; she steers.
*   **The Brain:** She uses a Utility-Based State Machine.
    1.  **Cruise:** Save energy.
    2.  **Jink:** You got too close? Break ankles.
    3.  **Orbit:** Use a Black Hole to sling-shot away.

## 4. The Rules of Engagement

1.  **Gravity is a Trap:** Inside the Event Horizon, gravity multiplies by 4x. It goes from "Influence" to "Cage."
2.  **The Memory:** The goal is always *inside* the danger zone. You have to fly into the singularity to find the truth.
3.  **Entropy:** The `FrostMask`. It heals over time. Your impact on the world is temporary. Keep moving or be forgotten.

## 5. How to Expand This

*   **Adding a Level:** Go to `initLevels`. Add a struct. Done.
*   **Adding a New Behavior:** Write a Lua script. The Go engine exposes `apply_force` and `cast_ray`. You don't need to recompile Go to make the Spectre smarter.
*   **Making it Prettier:** Edit the Kage shaders in `main.go`. We paint with math here.

---

**Final Note:**
The code is clean so the experience can be messy. Keep the ECS pure. Don't let game logic leak into the rendering loop.
