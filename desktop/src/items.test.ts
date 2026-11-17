import { describe, expect, it } from "vitest";
import { formatWhen, isOpenWork, joinedNav, splitItems, statusBucket, trayHeading, trayListPhase } from "./items";
import type { Item } from "./api";

function item(partial: Pick<Item, "scope" | "status"> & { id: string }): Item {
  return {
    id: partial.id,
    scope: partial.scope,
    trayRef: "t",
    title: partial.id,
    status: partial.status,
    createdAt: "2026-11-17T15:00:00.000Z",
    actions: [],
  };
}

describe("splitItems", () => {
  it("keeps open work out of the completed group", () => {
    const { open, done } = splitItems([
      item({ id: "a", scope: "local", status: "open" }),
      item({ id: "b", scope: "local", status: "done" }),
      item({ id: "c", scope: "remote", status: "pending" }),
      item({ id: "d", scope: "remote", status: "accepted" }),
      item({ id: "e", scope: "remote", status: "completed" }),
      item({ id: "f", scope: "remote", status: "snoozed" }),
      item({ id: "g", scope: "remote", status: "archived" }),
    ]);
    expect(open.map((row) => row.id)).toEqual(["a", "c", "d", "f"]);
    expect(done.map((row) => row.id)).toEqual(["b", "e", "g"]);
  });

  it("treats a declined remote item as finished", () => {
    expect(isOpenWork({ scope: "remote", status: "declined" })).toBe(false);
  });
});

describe("formatWhen", () => {
  it("returns nothing for a missing or zero date", () => {
    expect(formatWhen(undefined)).toBe("");
    expect(formatWhen("0001-01-01T00:00:00Z")).toBe("");
  });

  it("formats a real instant", () => {
    expect(formatWhen("2026-11-17T18:04:00.000Z").length).toBeGreaterThan(0);
  });
});

describe("trayListPhase", () => {
  it("stays loading until the selected tray's items arrive", () => {
    expect(trayListPhase("", "inbox", "")).toBe("loading");
    expect(trayListPhase("other", "inbox", "")).toBe("loading");
    expect(trayListPhase("inbox", "inbox", "")).toBe("ready");
  });

  it("stops loading when that tray's request fails", () => {
    expect(trayListPhase("", "inbox", "inbox")).toBe("error");
    expect(trayListPhase("inbox", "inbox", "inbox")).toBe("ready");
  });
});

describe("statusBucket", () => {
  it("keeps pending apart from open and completed", () => {
    expect(statusBucket("pending")).toBe("pending");
    expect(statusBucket("open")).toBe("open");
    expect(statusBucket("accepted")).toBe("open");
    expect(statusBucket("snoozed")).toBe("open");
    expect(statusBucket("completed")).toBe("completed");
    expect(statusBucket("done")).toBe("completed");
    expect(statusBucket("declined")).toBe("completed");
  });
});

describe("joinedNav", () => {
  it("leads with the alias and puts the owner underneath", () => {
    expect(joinedNav({ kind: "joined", name: "lindris", detail: "jessika-work", owner: "jessika@example.com" })).toEqual({
      title: "jessika-work",
      detail: "jessika@example.com",
    });
  });

  it("keeps the tray name when there is no alias", () => {
    expect(joinedNav({ kind: "joined", name: "lindris", owner: "jessika@example.com" }).title).toBe("lindris");
  });
});

describe("trayHeading", () => {
  it("drops the kind prefix used in attention headers", () => {
    expect(trayHeading("[dir] lindris-monorepo")).toBe("lindris-monorepo");
    expect(trayHeading("inbox")).toBe("inbox");
  });
});
