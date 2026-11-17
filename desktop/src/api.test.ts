import { afterEach, describe, expect, it, vi } from "vitest";
import { connect } from "./api";

const invoke = vi.hoisted(() => vi.fn());

vi.mock("@tauri-apps/api/core", () => ({
  invoke,
}));

afterEach(() => {
  vi.unstubAllGlobals();
  invoke.mockReset();
});

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("connect", () => {
  it("sends the bearer token and defaults the item filter to open", async () => {
    const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(
      async () => jsonResponse(200, []),
    );
    vi.stubGlobal("fetch", fetchMock);
    const api = connect("http://127.0.0.1:9/", "secret");

    await api.items("in box", "open");

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe("http://127.0.0.1:9/v1/items?tray=in+box&filter=open");
    expect(init?.method).toBe("GET");
    expect((init?.headers as Record<string, string>).Authorization).toBe(
      "Bearer secret",
    );
  });

  it("asks for all items only when the caller passes all", async () => {
    const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(
      async () => jsonResponse(200, []),
    );
    vi.stubGlobal("fetch", fetchMock);
    await connect("http://127.0.0.1:9", "secret").items("inbox", "all");
    const [url] = fetchMock.mock.calls[0];
    expect(url).toContain("filter=all");
  });

  it("restarts the API once after 401 and retries with the new token", async () => {
    const fetchMock = vi
      .fn<(url: string, init?: RequestInit) => Promise<Response>>()
      .mockResolvedValueOnce(
        jsonResponse(401, { error: "unavailable", message: "unauthorized" }),
      )
      .mockResolvedValueOnce(jsonResponse(200, { signedIn: true }));
    vi.stubGlobal("fetch", fetchMock);
    invoke.mockResolvedValue({
      base: "http://127.0.0.1:10",
      token: "next",
    });
    const api = connect("http://127.0.0.1:9", "stale");

    await expect(api.session()).resolves.toEqual({ signedIn: true });

    expect(invoke).toHaveBeenCalledTimes(1);
    expect(invoke).toHaveBeenCalledWith("restart_api");
    const retry = fetchMock.mock.calls[1]?.[1];
    expect(fetchMock.mock.calls[1]?.[0]).toBe("http://127.0.0.1:10/v1/session");
    expect((retry?.headers as Record<string, string>).Authorization).toBe(
      "Bearer next",
    );
  });

  it("does not restart when the user is signed out", async () => {
    const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(
      async () =>
        jsonResponse(401, { error: "signed_out", message: "signed out" }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(connect("http://127.0.0.1:9", "secret").session()).rejects.toThrow(
      "signed out",
    );
    expect(invoke).not.toHaveBeenCalled();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("restarts once when overlapping 401s reject the bearer", async () => {
    const pending: Array<(response: Response) => void> = [];
    const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(
      (_url, init) => {
        const auth = (init?.headers as Record<string, string> | undefined)
          ?.Authorization;
        if (auth === "Bearer next") {
          return Promise.resolve(jsonResponse(200, { signedIn: true }));
        }
        return new Promise((resolve) => {
          pending.push(resolve);
        });
      },
    );
    vi.stubGlobal("fetch", fetchMock);
    let releaseRestart: (value: { base: string; token: string }) => void = () => {};
    invoke.mockImplementation(
      () =>
        new Promise((resolve) => {
          releaseRestart = resolve;
        }),
    );

    const api = connect("http://127.0.0.1:9", "stale");
    const both = Promise.all([api.session(), api.session()]);
    expect(pending).toHaveLength(2);

    const unauthorized = jsonResponse(401, {
      error: "unavailable",
      message: "unauthorized",
    });
    pending[0]?.(unauthorized);
    pending[1]?.(
      jsonResponse(401, { error: "unavailable", message: "unauthorized" }),
    );
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(invoke).toHaveBeenCalledTimes(1);
    expect(invoke).toHaveBeenCalledWith("restart_api");
    releaseRestart({ base: "http://127.0.0.1:10", token: "next" });

    await expect(both).resolves.toEqual([
      { signedIn: true },
      { signedIn: true },
    ]);
    expect(invoke).toHaveBeenCalledTimes(1);
    const retries = fetchMock.mock.calls.slice(2);
    expect(retries).toHaveLength(2);
    for (const call of retries) {
      expect((call[1]?.headers as Record<string, string>).Authorization).toBe(
        "Bearer next",
      );
    }
  });

  it("throws the JSON message when the retried request is 401", async () => {
    const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(
      async () =>
        jsonResponse(401, { error: "unavailable", message: "unauthorized" }),
    );
    vi.stubGlobal("fetch", fetchMock);
    invoke.mockResolvedValue({ base: "http://127.0.0.1:10", token: "next" });

    await expect(connect("http://127.0.0.1:9", "stale").session()).rejects.toThrow(
      "unauthorized",
    );
    expect(invoke).toHaveBeenCalledTimes(1);
  });

  it("throws the JSON message for other errors and sends a body only when given", async () => {
    const fetchMock = vi
      .fn<(url: string, init?: RequestInit) => Promise<Response>>()
      .mockResolvedValueOnce(
        jsonResponse(400, { error: "invalid", message: "title is required" }),
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    vi.stubGlobal("fetch", fetchMock);
    const api = connect("http://127.0.0.1:9", "secret");

    await expect(api.add({ tray: "inbox", title: " " })).rejects.toThrow(
      "title is required",
    );

    await api.act("item 1", "complete");
    const [url, init] = fetchMock.mock.calls[1];
    expect(url).toBe("http://127.0.0.1:9/v1/items/item%201/complete");
    expect(init?.body).toBeUndefined();
    expect(init?.headers).not.toHaveProperty("Content-Type");

    fetchMock.mockResolvedValueOnce(new Response(null, { status: 204 }));
    await api.act("item 1", "snooze", { until: "2026-06-16T12:00:00.000Z" });
    const snooze = fetchMock.mock.calls[2]?.[1];
    expect(snooze?.body).toBe(
      JSON.stringify({ until: "2026-06-16T12:00:00.000Z" }),
    );
  });
});
