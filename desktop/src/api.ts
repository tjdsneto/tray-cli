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
  error?: unknown;
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

async function readErrorBody(res: Response): Promise<ErrorBody | null> {
  try {
    const data = (await res.json()) as unknown;
    if (!data || typeof data !== "object" || Array.isArray(data)) return null;
    return data as ErrorBody;
  } catch {
    return null;
  }
}

function errorText(body: ErrorBody | null): string {
  return typeof body?.message === "string" ? body.message : "request failed";
}

export function connect(base: string, token: string) {
  let apiBase = base;
  let apiToken = token;
  let restarting: Promise<void> | null = null;

  function restartApi(): Promise<void> {
    if (!restarting) {
      restarting = invoke<{ base: string; token: string }>("restart_api")
        .then((next) => {
          apiBase = next.base;
          apiToken = next.token;
        })
        .finally(() => {
          restarting = null;
        });
    }
    return restarting;
  }

  async function request<T>(
    method: string,
    path: string,
    body?: unknown,
    allowRestart = true,
  ): Promise<T> {
    const sentToken = apiToken;
    const headers: Record<string, string> = {
      Authorization: `Bearer ${sentToken}`,
    };
    const init: RequestInit = { method, headers };
    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
      init.body = JSON.stringify(body);
    }

    const res = await fetch(endpoint(apiBase, path), init);
    if (res.status === 401) {
      const failure = await readErrorBody(res);
      const message = errorText(failure);
      if (failure?.error === "signed_out") throw new Error(message);
      const bearerRejected = failure === null || failure.error === "unavailable";
      if (!bearerRejected || !allowRestart) throw new Error(message);
      if (sentToken !== apiToken) return request<T>(method, path, body, false);
      await restartApi();
      return request<T>(method, path, body, false);
    }
    if (!res.ok) throw new Error(errorText(await readErrorBody(res)));
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
