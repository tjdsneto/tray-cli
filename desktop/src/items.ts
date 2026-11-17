import type { ActionId } from "./actions";
import type { Item } from "./api";

export function isOpenWork(item: Pick<Item, "scope" | "status">): boolean {
  if (item.scope === "local") return item.status === "open";
  return (
    item.status === "pending" ||
    item.status === "accepted" ||
    item.status === "snoozed"
  );
}

export function splitItems(items: Item[]): { open: Item[]; done: Item[] } {
  const open: Item[] = [];
  const done: Item[] = [];
  for (const item of items) {
    if (isOpenWork(item)) open.push(item);
    else done.push(item);
  }
  return { open, done };
}

export function formatWhen(value: string | undefined): string {
  if (!value) return "";
  const date = new Date(value);
  const time = date.getTime();
  if (Number.isNaN(time) || time < 0) return "";
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(date);
}

export function attentionActions(scope: string): ActionId[] {
  if (scope === "remote") return ["accept", "decline", "snooze"];
  if (scope === "local") return ["complete"];
  return [];
}

export function trayHeading(header: string): string {
  return header.replace(/^\[[^\]]+\]\s*/, "");
}

export type StatusBucket = "pending" | "open" | "completed";

export function joinedNav(tray: { kind?: string; name: string; detail?: string; owner?: string }): { title: string; detail: string } {
  if (tray.kind !== "joined") return { title: tray.name, detail: "" };
  const alias = tray.detail?.trim() ?? "";
  return { title: alias || tray.name, detail: tray.owner?.trim() ?? "" };
}

export function trayListPhase(loadedKey: string, trayKey: string, errorKey: string): "loading" | "ready" | "error" {
  if (loadedKey === trayKey) return "ready";
  if (errorKey === trayKey) return "error";
  return "loading";
}

export function statusBucket(status: string): StatusBucket {
  switch (status) {
    case "pending":
      return "pending";
    case "done":
    case "completed":
    case "declined":
    case "archived":
      return "completed";
    default:
      return "open";
  }
}
