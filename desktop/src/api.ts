import { invoke } from "@tauri-apps/api/core";
import type { ActionId } from "./actions";

export type Session = {
  signedIn: boolean;
};

export type SnapshotRow = {
  scope: string;
  trayRef: string;
  itemId: string;
  title: string;
  clipboard: string;
};

export type SnapshotSection = {
  header: string;
  rows: SnapshotRow[];
};

export type Snapshot = {
  badge: number;
  remoteStatus: string;
  sections: SnapshotSection[];
};

export type Tray = {
  scope: string;
  ref: string;
  name: string;
  openCount: number;
};

export type Item = {
  id: string;
  scope: string;
  trayRef: string;
  title: string;
  status: string;
  createdAt: string;
  snoozeUntil?: string;
  body?: string;
  actions: ActionId[];
};

export type AddInput = {
  tray: string;
  title: string;
  body?: string;
};

type ErrorBody = {
  message?: unknown;
};

function endpoint(base: string, path: string): string {
  return `${base.replace(/\/+$/, "")}${path}`;
}

function itemsQuery(tray: string, filter: string | undefined): string {
  const params = new URLSearchParams();
  params.set("tray", tray);
  params.set("filter", filter === "all" ? "all" : "open");
  return `/v1/items?${params.toString()}`;
}

async function errorMessage(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as ErrorBody;
    if (typeof data.message === "string") return data.message;
  } catch {
    // Non-JSON bodies still surface as a failed request.
  }
  return "request failed";
}

export function connect(base: string, token: string) {
  let apiBase = base;
  let apiToken = token;

  async function request<T>(
    method: string,
    path: string,
    body?: unknown,
    retry = true,
  ): Promise<T> {
    const headers: Record<string, string> = {
      Authorization: `Bearer ${apiToken}`,
    };
    const init: RequestInit = { method, headers };
    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
      init.body = JSON.stringify(body);
    }

    const res = await fetch(endpoint(apiBase, path), init);
    if (res.status === 401) {
      if (!retry) throw new Error(await errorMessage(res));
      const next = await invoke<{ base: string; token: string }>("restart_api");
      apiBase = next.base;
      apiToken = next.token;
      return request<T>(method, path, body, false);
    }
    if (!res.ok) throw new Error(await errorMessage(res));
    if (res.status === 204) return undefined as T;
    const text = await res.text();
    if (!text) return undefined as T;
    return JSON.parse(text) as T;
  }

  return {
    session: () => request<Session>("GET", "/v1/session"),
    snapshot: () => request<Snapshot>("GET", "/v1/snapshot"),
    trays: () => request<Tray[]>("GET", "/v1/trays"),
    items: (tray: string, filter?: string) =>
      request<Item[]>("GET", itemsQuery(tray, filter)),
    item: (id: string) =>
      request<Item>("GET", `/v1/items/${encodeURIComponent(id)}`),
    add: (input: AddInput) => request<Item>("POST", "/v1/items", input),
    act: (id: string, action: string, body?: unknown) =>
      request<void>(
        "POST",
        `/v1/items/${encodeURIComponent(id)}/${encodeURIComponent(action)}`,
        body,
      ),
  };
}

export type TrayClient = ReturnType<typeof connect>;
