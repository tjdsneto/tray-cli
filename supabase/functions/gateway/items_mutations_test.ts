import { assertEquals } from "jsr:@std/assert";
import { buildAddItemBody, buildStatusPatch } from "./items_mutations.ts";

Deno.test("buildAddItemBody accepts on own tray", () => {
  const b = buildAddItemBody("user-1", "t1", "  Hi  ", "user-1");
  assertEquals(b.status, "accepted");
  assertEquals(b.title, "Hi");
  assertEquals(b.tray_id, "t1");
});

Deno.test("buildAddItemBody pending on others tray", () => {
  const b = buildAddItemBody("user-1", "t1", "Hi", "owner-z");
  assertEquals(b.status, "pending");
});

Deno.test("buildStatusPatch merges extras", () => {
  assertEquals(buildStatusPatch("completed", { completion_message: "done" }), {
    status: "completed",
    completion_message: "done",
  });
});
