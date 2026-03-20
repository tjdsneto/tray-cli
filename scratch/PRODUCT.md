# Tray CLI

**Product:** **Tray CLI** — a remote, shared inbox-tray: people add things for you to get to *when you get to it*, across life areas. **Implementation:** CLI written in **Go**; backend **Supabase** (Postgres, Auth, RLS, PostgREST). **Repo / historical label:** attention-queue.

**Idea:** A list others can add to without pulling you into urgent chat or email. “Add that to my tray.”

- **Named trays:** Work, life, other things — configurable. Each tray is a list you own (e.g. “my work tray,” “my life tray”).
- **Share a tray, not a project:** You share a *tray* with people, not a whole project. Example: create a “Work” tray, share it with colleagues; they add items when they want your attention asynchronously. One tray = one shared “drop things here for me” channel.
- **Review + prioritization:** You review what landed in your tray and can set prioritization rules to process it.
- **CLI-first / agent-friendly:** Command **`tray`** (the Tray CLI binary) so it’s easy to use from Claude or other agents in the terminal (add, list, prioritize, review from the shell).
- **Positioning:** Simpler than full task management; more global/personal; has an email-like “inbox others can send to” feel.

**Why not just message / email?** For things you *don’t* want urgent attention: the tray is dedicated to “when you get to it” — you don’t have a bunch of other stuff mixed in (like email), so you might get to it sooner or more predictably. It’s a separate channel for non-urgent “put this on my radar” that isn’t lost in the noise.

**Naming note:** “Tray” overlaps with [Tray.io / Tray.ai](https://tray.io) (integration/automation). Different product; see [naming-ideas.md](naming-ideas.md) for conflict check and alternatives if you need a clearer namespace.

**Question to revisit:** Is this meaningfully different from existing tools (task managers, shared inboxes, “assign to me” flows), or mostly a different framing of the same thing?

---

*Moved from tjdsneto-setup/notes/product-ideas.md — 2026-03-18*
