# Invite Flow, Item Lifecycle & Scenarios

For **Tray CLI** (Go client, **Supabase** backend). This document covers how users join buckets (invite flows), how bucket access is managed, how items move through states, and example end-to-end scenarios. It feeds into CLI/API design and backend requirements.

---

## 1. Invite flow: two models

We need to support (at least one of, or both) ways for people to become members of a bucket.

### Model A: Owner invites individual users

- **Flow:** Bucket owner invites by identifier (e.g. email). Invitee receives a link or in-app prompt; they accept and are added as a member.
- **Pros:** Owner explicitly chooses who can add items; good for small, known groups.
- **Cons:** Friction for each new person; owner must know them in advance.

**Implied capabilities:** Invite table (pending by email/token), send email/link, accept endpoint that creates a bucket member (e.g. `bucket_members` row) (and optionally triggers sign-up).

### Model B: Shareable invite (link or token) — self-serve join

- **Flow:** Bucket owner generates a **shareable invite** (link or token). Anyone with the link can use it to **add themselves** as a member (no owner approval per user).
- **Pros:** Low friction; easy to share in a doc, Slack, or profile (“put things on my radar: https://…”).
- **Cons:** Anyone with the link can join; need a way to revoke and rotate so the link isn’t permanent.

**Implied capabilities:** Per-bucket (or per-invite) **invite token**; “join with this token” flow that creates a member; **token rotation** so the owner can invalidate the old link and issue a new one.

### Recommendation

- Support **both** for flexibility: some buckets use “invite by email” (Model A), others use “share this link” (Model B).
- For MVP, **Model B (shareable link)** is often enough and simpler to implement (no email sending). Add Model A when you need explicit approval or email-based onboarding.

---

## 2. Bucket access management

Who can add items to a bucket, and how the owner controls it.

### List who’s in a bucket

- **Owner** can list members (and optionally pending invites).
- **CLI:** e.g. `tray members <tray>`.

### Revoke a member

- **Owner** can remove a user from a bucket; they can no longer add items (and optionally lose read access to “items I added” — see below).
- **CLI:** e.g. `tray revoke <tray> <user-id-or-email>`.

### Rotate invite token (for shareable-invite buckets)

- **Owner** can rotate the invite token so the previous link/token stops working; new link is generated.
- **CLI:** e.g. `tray rotate-invite work` → outputs new invite URL/token.
- **Use case:** Link leaked, or periodic rotation for security.

### Data model (conceptual)

- **Buckets:** `id`, `owner_id`, `name`, optional `invite_token` (for Model B), `created_at`, etc.
- **Bucket members:** `bucket_id`, `user_id`, `joined_at`, optional `invited_via` (email vs token).
- **Invites (Model A):** `bucket_id`, `email`, `token`, `status` (pending/accepted/expired), `created_at`.

### Identifying buckets: remotes (aliases) vs owner-provided names

When a **member** adds an item to someone else’s bucket, they need a **stable, unambiguous** way to refer to that bucket. The bucket name alone is not unique globally (many owners can have a bucket named `work`).

**Option A: Owner-provided identifier**  
The backend exposes a unique id per bucket (e.g. UUID) or a compound key (e.g. `owner_handle/bucket_slug` if owners have unique handles). So Jordan would run `tray add "..." alex/work` or `tray add "..." <bucket-uuid>`.  
- **Pros:** No local state; one canonical identifier.  
- **Cons:** Requires globally unique owner handles (or opaque IDs); users must remember or look up the identifier.

**Option B: Git-remote style aliases (recommended)**  
Members define **local aliases** (remotes) that point to a specific bucket. After joining via invite link, the user either gets a suggested alias or runs something like `tray remote add alex-work <invite-url-or-token>`. That stores locally: alias `alex-work` → resolved bucket (e.g. bucket_id or owner+bucket). Then: `tray add "Review deck" alex-work`.  
- **Pros:** No global uniqueness constraint on bucket names; each user picks memorable names; familiar model (git remotes); owner can rename their bucket without breaking members (alias stays bound to bucket_id).  
- **Cons:** One-time setup per bucket; alias list is local (could sync later).

**Recommendation:** Use **remotes (aliases)**. The backend only needs to resolve invite token → bucket_id (and optionally return owner + bucket name for a suggested alias). Members use `tray remote add <alias> <invite-url>` (or `tray join <invite-url>` creates an alias automatically). Then `tray add "..." <alias>` is unambiguous. For **owners**, their own buckets can still be referred to by local bucket name (e.g. `tray list work`) since those are already unambiguous in their context.

**CLI sketch:** `tray remote add <alias> <invite-url-or-token>`, `tray remote ls`, `tray remote remove <alias>`. After join, `tray add "title" <alias>`.

---

## 3. Member view: “Items I added to others’ buckets”

Members need to see **what they put on someone else’s radar** and the **status** of those items (so they know if it’s pending, accepted, in progress, or done).

### Requirements

- **Query:** “List items where I am `source_user_id`” (items I added), optionally filtered by bucket or owner.
- **Fields:** At least item `id`, `title`, `bucket` (or bucket name), **status**, `created_at`, and optionally `updated_at`, `due_date`, `decline_reason`, `completion_message`, owner's display name.
- **CLI:** e.g. `tray contributed` or `tray list --mine-added` or `tray outgoing` — list items I added to any bucket, with status.

### Backend

- Items table has `source_user_id` (who added it). Member can read items where `source_user_id = auth.uid()`; RLS or equivalent must allow that even when they’re not the bucket owner (read-only for “their” items across buckets).

---

## 4. Item lifecycle: statuses and actions

Items are not just “in the queue”; they have a **status** and the owner (and possibly the member who added it) can perform **actions** on them.

### Statuses (proposed)

| Status      | Meaning |
|------------|---------|
| `pending`  | New; owner hasn’t triaged yet. Default for new items. |
| `accepted` | Owner acknowledged; will (or might) act on it. |
| `declined` | Owner explicitly declined (won’t do it, or not relevant). |
| `snoozed`  | Deferred; show again after a date (or until “unsnooze”). |
| `completed`| Done. |
| `archived` | Done or dismissed without a formal “completed”; hide from active list. |

### Owner actions (triage)

- **Accept** → `accepted` (item stays in queue, marked as acknowledged).
- **Decline** → `declined` (member who added it can see it was declined, and optional reason).
- **Snooze** → `snoozed` + optional `snooze_until` (or default e.g. 1 week).
- **Complete** → `completed`.
- **Archive** → `archived` (e.g. “done but not worth marking completed” or “dismiss”).

### Optional: due date

- **Field:** `due_date` (optional) on the item.
- **Set by:** Member when adding, or owner when editing. Useful for “when you get to it, ideally by X.”
- **CLI:** `tray add "Review deck" work --due 2025-03-25`; `tray list` can show or sort by due date.

### Default list view

- **Active** items: `pending`, `accepted`, `snoozed` (if `snooze_until` is past or null). Exclude `declined`, `completed`, `archived` from the default “my queue” view, or show them in a separate view/filter (e.g. `tray list --include archived`).

---

## 5. Example scenarios

Concrete flows to validate the design.

### Scenario 1: Shareable link, colleague adds an item

1. **Alex** (owner) creates bucket `work`, runs `tray invite work` (or similar), gets link: `https://app.example.com/join/abc123` (or token `abc123` for CLI).
2. Alex pastes the link in team Slack: “Put non-urgent stuff for me here.”
3. **Jordan** (colleague) opens link, signs up or logs in, and is added as member. The app (or CLI) creates a **remote** for that bucket — e.g. Jordan runs `tray remote add alex-work <invite-url>` (or `tray join <url>` and is prompted for an alias, default e.g. `alex-work`).
4. Jordan runs: `tray add "Review Q1 deck when you have time" alex-work` — using the **alias**, not Alex's bucket name, so the target is unambiguous.
5. Item appears in Alex’s queue with status `pending`, `source_user_id` = Jordan.

### Scenario 2: Owner triages — accept, snooze, complete

1. **Alex** runs `tray list work` and sees the item from Jordan.
2. Alex runs `tray accept <item-id>` (or `tray review` and chooses accept) → status `accepted`.
3. Later Alex runs `tray snooze <item-id> --until 2025-03-22` → status `snoozed`, item hidden until that date.
4. On 2025-03-22 the item reappears in the active list. Alex runs `tray complete <item-id>` (optionally `tray complete <id> --message "Deck shared with team"`) → status `completed`, optional `completion_message` stored.
5. Jordan can run `tray contributed` and sees that item as `completed` (and the completion message if set).

### Scenario 3: Owner declines; member sees status

1. **Sam** adds an item to **Morgan’s** bucket: “Can we switch to tool X?”
2. Morgan decides not to pursue it and runs `tray decline <item-id>` (optionally with reason: `tray decline <id> --reason "Not this quarter"`).
3. Sam runs `tray contributed` and sees the item with status `declined` (and optional reason). No need to follow up.

### Scenario 4: Member checks status of their requests

1. **Jordan** has added several items to Alex’s and Sam’s buckets over the week.
2. Jordan runs `tray contributed` (or `tray contributed --json`).
3. Sees a table: item title, bucket (or owner), status (`pending` / `accepted` / `completed` / `declined`), date added. Can prioritize follow-ups or stop wondering.

### Scenario 5: Owner rotates invite link after leak

1. **Alex** shared the invite link in a public doc by mistake.
2. Alex runs `tray rotate-invite work`. Old link stops working; new link is printed.
3. Alex updates the doc with the new link. Existing members are unchanged; only the join link is rotated.

### Scenario 6: Due date and prioritization

1. **Jordan** adds: `tray add "Sign off on contract" alex-work --due 2025-03-20` (using their remote alias for Alex's bucket).
2. Alex runs `tray list work` and sees due date; can sort or filter by due (`tray list work --due-before 2025-03-25`).
3. Alex completes the item before the due date; Jordan sees status `completed` in `tray contributed`.

---

## 6. CLI/API sketch (additions)

Summary of commands or flags that this doc implies, to align with the main CLI design.

| Area | Command / flag | Purpose |
|------|----------------|---------|
| Invite (Model B) | `tray invite <tray>` | Show or generate shareable invite link/token. |
| Invite (Model A) | `tray invite-email <email> [tray]` | Invite by email (sends link). |
| Join | `tray join <token-or-url>` | Self-join using shareable invite; optionally create remote alias. |
| Remotes | `tray remote add <alias> <invite-url>`, `tray remote ls`, `tray remote remove <alias>` | Alias for "someone else's bucket"; use alias in `tray add "..." <alias>`. |
| Members | `tray members <tray>` | List members (and optionally pending invites). |
| Revoke | `tray revoke <tray> <user>` | Remove member. |
| Rotate | `tray rotate-invite <tray>` | Rotate invite token; output new link. |
| Member view | `tray contributed` / `tray list --mine-added` | List items I added to others’ buckets, with status. |
| Triage | `tray accept <id>`, `tray decline <id>`, `tray snooze <id> [--until DATE]`, `tray complete <id> [--message ...]`, `tray archive <id>` | Set item status; optional `--reason` for decline, `--message` for complete. |
| Due date | `tray add "..." bucket --due DATE`; `tray list --due-before DATE` | Set or filter by due date. |

---

## 7. Backend implications (brief)

Full backend and DB spec: [backend-spec.md](backend-spec.md). CLI command and local-storage spec: [cli-design.md](cli-design.md).

- **Invite token:** Store on bucket (or invite table); rotation = new token, old one invalid for new joins. Existing members unchanged.
- **RLS:** Members can read items where they are `source_user_id` (across buckets they're a member of). Owner can read/write all items in their buckets.
- **Item table:** Add `status`, optional `snooze_until`, optional `due_date`, optional `decline_reason`, optional `completion_message`.
- **Remotes (aliases):** Stored locally (e.g. `~/.config/tray/remotes` or similar): alias → bucket_id or invite_token. Backend resolves alias on the client, or CLI sends alias and backend resolves via a "remotes" table keyed by user + alias. Simplest: store alias → (bucket_id or invite_token) locally; `tray add "..." <alias>` looks up bucket_id and sends it to API.
- **Model A (email invite):** Still requires Edge/Cloud Function (send email, accept endpoint). Model B can be implemented with just token + “join” endpoint and no email.

---

*Last updated: 2026-03-20 (Tray CLI; `tray` commands aligned with cli-design).*
