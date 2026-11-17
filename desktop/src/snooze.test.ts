import { describe, it, expect } from "vitest";
import { snoozeUntil } from "./snooze";

const now = new Date("2026-06-15T18:30:00-03:00");

describe("snoozeUntil", () => {
  it("adds one hour for hour preset", () => {
    expect(snoozeUntil("hour", now).getTime()).toBe(
      now.getTime() + 60 * 60 * 1000,
    );
  });

  it("sets tomorrow at 9:00 local", () => {
    expect(snoozeUntil("tomorrow", now).getTime()).toBe(
      new Date("2026-06-16T09:00:00-03:00").getTime(),
    );
  });

  it("sets next week at 9:00 local", () => {
    expect(snoozeUntil("week", now).getTime()).toBe(
      new Date("2026-06-22T09:00:00-03:00").getTime(),
    );
  });
});