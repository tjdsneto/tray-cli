import { assertEquals } from "jsr:@std/assert";
import {
  itemsContributedPath,
  itemsListAllVisiblePath,
  itemsListPath,
  trayMembersJoinedPath,
  trayOwnerSelectPath,
  traysListOwnedPath,
} from "./postgrest_paths.ts";

Deno.test("trayOwnerSelectPath requests owner_id for one tray", () => {
  const p = trayOwnerSelectPath("11111111-1111-4111-8111-111111111111");
  assertEquals(p.includes("select=owner_id"), true);
  assertEquals(p.includes("id=eq.11111111-1111-4111-8111-111111111111"), true);
});

Deno.test("traysListOwnedPath filters by owner like tray ls", () => {
  const uid = "11111111-1111-4111-8111-111111111111";
  const p = traysListOwnedPath(uid);
  assertEquals(decodeURIComponent(p).includes("items(count)"), true);
  assertEquals(p.includes(`owner_id=eq.${uid}`), true);
  assertEquals(p.startsWith("/rest/v1/trays?"), true);
});

Deno.test("itemsListPath sets tray filter and limit", () => {
  const p = itemsListPath({
    trayId: "11111111-1111-4111-8111-111111111111",
    status: "pending",
    limit: 10,
  });
  assertEquals(p.includes("tray_id=eq.11111111-1111-4111-8111-111111111111"), true);
  assertEquals(p.includes("status=eq.pending"), true);
  assertEquals(p.includes("limit=10"), true);
});

Deno.test("trayMembersJoinedPath queries tray_members embed", () => {
  const uid = "22222222-2222-4222-8222-222222222222";
  const p = trayMembersJoinedPath(uid);
  assertEquals(p.includes("tray_members?"), true);
  assertEquals(p.includes(`user_id=eq.${uid}`), true);
  assertEquals(decodeURIComponent(p).includes("trays(id"), true);
});

Deno.test("itemsListAllVisiblePath omits tray_id by default", () => {
  const p = itemsListAllVisiblePath({ limit: 50 });
  assertEquals(p.includes("tray_id="), false);
  assertEquals(p.includes("limit=50"), true);
  assertEquals(p.includes("updated_at.desc"), true);
});

Deno.test("itemsContributedPath filters by source_user_id and embeds tray owner", () => {
  const uid = "33333333-3333-4333-8333-333333333333";
  const p = itemsContributedPath({ userId: uid, limit: 25 });
  assertEquals(p.includes(`source_user_id=eq.${uid}`), true);
  assertEquals(decodeURIComponent(p).includes("trays(owner_id)"), true);
  assertEquals(p.includes("limit=25"), true);
});
