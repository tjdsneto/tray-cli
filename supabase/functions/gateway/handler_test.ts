import { assertEquals } from "jsr:@std/assert";
import { createHandler } from "./handler.ts";
import { signGatewayToken } from "./jwt.ts";
import type {
  GatewayStore,
  OAuthCodeRecord,
  OAuthPendingRecord,
  RefreshTokenRecord,
  SupabaseAuthRecord,
  SupabaseAuthUpsert,
} from "./store.ts";

class InMemoryStore implements GatewayStore {
  public codes: OAuthCodeRecord[] = [];
  public refresh: RefreshTokenRecord[] = [];
  public pending: OAuthPendingRecord[] = [];
  public supabaseAuth = new Map<string, SupabaseAuthRecord>();

  async createOAuthCode(record: OAuthCodeRecord): Promise<void> {
    this.codes.push(record);
  }

  async consumeOAuthCode(codeHash: string, clientID: string, redirectURI: string) {
    const row = this.codes.find((c) =>
      c.codeHash === codeHash && c.clientID === clientID && c.redirectURI === redirectURI
    );
    if (!row) return null;
    return { userID: row.userID, scope: row.scope, clientID: row.clientID };
  }

  async createRefreshToken(record: RefreshTokenRecord): Promise<void> {
    this.refresh.push(record);
  }

  async consumeRefreshToken(tokenHash: string, clientID: string) {
    const row = this.refresh.find((r) => r.tokenHash === tokenHash && r.clientID === clientID);
    if (!row) return null;
    return { userID: row.userID, scope: row.scope, clientID: row.clientID };
  }

  async createOAuthPending(record: OAuthPendingRecord): Promise<void> {
    this.pending.push(record);
  }

  async getOAuthPending(id: string): Promise<OAuthPendingRecord | null> {
    const row = this.pending.find((p) => p.id === id);
    if (!row) return null;
    const exp = Date.parse(row.expiresAtISO);
    if (exp <= Date.now()) return null;
    return row;
  }

  async deleteOAuthPending(id: string): Promise<void> {
    this.pending = this.pending.filter((p) => p.id !== id);
  }

  async upsertSupabaseAuth(record: SupabaseAuthUpsert): Promise<void> {
    this.supabaseAuth.set(record.userID, {
      accessToken: record.accessToken,
      refreshToken: record.refreshToken,
      accessExpiresAtISO: record.accessExpiresAtISO,
    });
  }

  async getSupabaseAuth(userID: string): Promise<SupabaseAuthRecord | null> {
    return this.supabaseAuth.get(userID) ?? null;
  }
}

function req(path: string, init?: RequestInit) {
  return new Request(`https://example.test${path}`, init);
}

function deps(store: GatewayStore, fetchFn: typeof fetch) {
  let uuidSeq = 0;
  return {
    getEnv: (name: string) => {
      if (name === "SUPABASE_URL") return "https://proj.supabase.co";
      if (name === "SUPABASE_ANON_KEY") return "anon";
      if (name === "GATEWAY_JWT_SECRET") return "secret";
      if (name === "GATEWAY_TOKEN_ISSUER") return "tray-gateway";
      if (name === "GATEWAY_TOKEN_AUDIENCE") return "chatgpt";
      if (name === "GATEWAY_OAUTH_PROVIDER") return "google";
      return undefined;
    },
    fetchFn,
    store,
    now: () => new Date("2026-05-05T12:00:00Z"),
    randomToken: () => "raw-code-token",
    randomUUID: () => {
      uuidSeq += 1;
      return `00000000-0000-4000-8000-${String(100000000000 + uuidSeq).slice(0, 12)}`;
    },
  };
}

Deno.test("health endpoint responds ok", async () => {
  const store = new InMemoryStore();
  const handler = createHandler(deps(store, (async () => new Response("unused")) as typeof fetch));
  const r = await handler(req("/health"));
  assertEquals(r.status, 200);
  assertEquals((await r.json()).ok, true);
});

Deno.test("oauth authorize redirects to Supabase Auth with PKCE", async () => {
  const store = new InMemoryStore();
  const handler = createHandler(deps(store, (async () => new Response("unused")) as typeof fetch));
  const r = await handler(
    req(
      "/oauth/authorize?response_type=code&client_id=chatgpt-gateway-dev&redirect_uri=https%3A%2F%2Fchat.openai.com%2Fcb&state=s1&scope=tray.read",
    ),
  );
  assertEquals(r.status, 302);
  const loc = r.headers.get("Location") ?? "";
  assertEquals(loc.startsWith("https://proj.supabase.co/auth/v1/authorize?"), true);
  assertEquals(loc.includes("provider=google"), true);
  assertEquals(loc.includes("code_challenge="), true);
  assertEquals(loc.includes("code_challenge_method=s256"), true);
  assertEquals(loc.includes(encodeURIComponent("https://proj.supabase.co/functions/v1/gateway/oauth/supabase-callback/")), true);
  assertEquals(store.pending.length, 1);
  assertEquals(store.pending[0].clientID, "chatgpt-gateway-dev");
});

Deno.test("supabase callback exchanges PKCE and redirects to ChatGPT redirect_uri", async () => {
  const store = new InMemoryStore();
  const pendingId = "11111111-1111-4111-8111-111111111111";
  await store.createOAuthPending({
    id: pendingId,
    codeVerifier: "verifier-secret",
    clientID: "chatgpt-gateway-dev",
    redirectURI: "https://chat.openai.com/cb",
    state: "state-xyz",
    scope: "tray.read",
    expiresAtISO: "2099-01-01T00:00:00Z",
  });

  const calls: string[] = [];
  const handler = createHandler(
    deps(store, (async (input, init) => {
      calls.push(String(input));
      const body = init && "body" in init && init.body ? String(init.body) : "";
      if (String(input).includes("/auth/v1/token?grant_type=pkce")) {
        assertEquals(body.includes("auth_code"), true);
        assertEquals(body.includes("code_verifier"), true);
        return new Response(
          JSON.stringify({
            access_token: "at",
            refresh_token: "rt",
            expires_in: 3600,
            user: { id: "user-9" },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      return new Response("unexpected", { status: 500 });
    }) as typeof fetch),
  );

  const r = await handler(
    req(`/oauth/supabase-callback/${pendingId}?code=supabase-auth-code-123`),
  );
  assertEquals(r.status, 302);
  const loc = r.headers.get("Location") ?? "";
  assertEquals(loc.startsWith("https://chat.openai.com/cb?"), true);
  assertEquals(loc.includes("code=raw-code-token"), true);
  assertEquals(loc.includes("state=state-xyz"), true);
  assertEquals(store.pending.length, 0);
  assertEquals(store.codes.length, 1);
  assertEquals(store.codes[0].userID, "user-9");
  const sa = await store.getSupabaseAuth("user-9");
  assertEquals(sa?.accessToken, "at");
  assertEquals(sa?.refreshToken, "rt");
});

Deno.test("token exchange issues access and refresh tokens", async () => {
  const store = new InMemoryStore();
  store.codes.push({
    codeHash: "1fd164b62fe53b187ddfcaa31c6320901b30263e9f204fa40d98dda0f1913780",
    userID: "user-1",
    clientID: "c1",
    redirectURI: "https://app.example/cb",
    scope: "tray.read",
    expiresAtISO: "2099-01-01T00:00:00Z",
  });
  const handler = createHandler(deps(store, (async () => new Response("unused")) as typeof fetch));

  const body = new URLSearchParams({
    grant_type: "authorization_code",
    client_id: "c1",
    code: "raw-code-token",
    redirect_uri: "https://app.example/cb",
  });
  const r = await handler(req("/oauth/token", { method: "POST", body }));
  assertEquals(r.status, 200);
  const json = await r.json();
  assertEquals(typeof json.access_token, "string");
  assertEquals(typeof json.refresh_token, "string");
  assertEquals(json.scope, "tray.read");
});

Deno.test("v1/trays proxies PostgREST with user bearer", async () => {
  const store = new InMemoryStore();
  await store.upsertSupabaseAuth({
    userID: "user-1",
    accessToken: "sat",
    refreshToken: "srt",
    accessExpiresAtISO: "2099-01-01T00:00:00Z",
  });
  const handler = createHandler(
    deps(store, (async (input, init) => {
      const url = String(input);
      const headers = init && typeof init === "object" && "headers" in init
        ? (init as RequestInit).headers
        : undefined;
      if (url.includes("/rest/v1/trays?") && url.includes("owner_id=eq.user-1") && headers) {
        const h = new Headers(headers as HeadersInit);
        assertEquals(h.get("Authorization"), "Bearer sat");
        return new Response(
          JSON.stringify([{
            id: "t1",
            owner_id: "user-1",
            name: "Inbox",
            created_at: "2026-01-01T00:00:00Z",
            items: [{ count: 3 }],
          }]),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      return new Response("unexpected", { status: 500 });
    }) as typeof fetch),
  );
  const token = await signGatewayToken("secret", {
    sub: "user-1",
    scope: "tray.read",
    iss: "tray-gateway",
    aud: "chatgpt",
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 600,
  });
  const r = await handler(req("/v1/trays", { headers: { Authorization: `Bearer ${token}` } }));
  assertEquals(r.status, 200);
  const j = await r.json() as { trays: Array<{ name: string; item_count: number; is_owner: boolean }> };
  assertEquals(j.trays.length, 1);
  assertEquals(j.trays[0].name, "Inbox");
  assertEquals(j.trays[0].item_count, 3);
  assertEquals(j.trays[0].is_owner, true);
});

Deno.test("v1/remotes lists joined trays excluding own", async () => {
  const store = new InMemoryStore();
  await store.upsertSupabaseAuth({
    userID: "user-1",
    accessToken: "sat",
    refreshToken: "srt",
    accessExpiresAtISO: "2099-01-01T00:00:00Z",
  });
  const handler = createHandler(
    deps(store, (async (input, init) => {
      const url = String(input);
      const headers = init && typeof init === "object" && "headers" in init
        ? (init as RequestInit).headers
        : undefined;
      if (url.includes("/rest/v1/tray_members?") && url.includes("user_id=eq.user-1") && headers) {
        return new Response(
          JSON.stringify([
            {
              joined_at: "2026-02-01T00:00:00Z",
              trays: {
                id: "t-remote",
                owner_id: "owner-z",
                name: "Team",
                created_at: "2026-01-01T00:00:00Z",
                items: [{ count: 1 }],
              },
            },
            {
              joined_at: "2026-03-01T00:00:00Z",
              trays: {
                id: "t-own",
                owner_id: "user-1",
                name: "Mine",
                created_at: "2026-01-01T00:00:00Z",
                items: [{ count: 0 }],
              },
            },
          ]),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      return new Response("unexpected", { status: 500 });
    }) as typeof fetch),
  );
  const token = await signGatewayToken("secret", {
    sub: "user-1",
    scope: "tray.read",
    iss: "tray-gateway",
    aud: "chatgpt",
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 600,
  });
  const r = await handler(req("/v1/remotes", { headers: { Authorization: `Bearer ${token}` } }));
  assertEquals(r.status, 200);
  const j = await r.json() as { remotes: Array<{ id: string; is_owner: boolean }> };
  assertEquals(j.remotes.length, 1);
  assertEquals(j.remotes[0].id, "t-remote");
  assertEquals(j.remotes[0].is_owner, false);
});

Deno.test("v1/items/contributed excludes items on trays you own", async () => {
  const store = new InMemoryStore();
  await store.upsertSupabaseAuth({
    userID: "user-1",
    accessToken: "sat",
    refreshToken: "srt",
    accessExpiresAtISO: "2099-01-01T00:00:00Z",
  });
  const handler = createHandler(
    deps(store, (async (input, init) => {
      const url = String(input);
      if (url.includes("source_user_id=eq.user-1") && decodeURIComponent(url).includes("trays(owner_id)")) {
        return new Response(
          JSON.stringify([
            {
              id: "i-own-tray",
              tray_id: "t1",
              title: "On my tray",
              status: "pending",
              trays: { owner_id: "user-1" },
            },
            {
              id: "i-theirs",
              tray_id: "t2",
              title: "On Alice tray",
              status: "pending",
              trays: { owner_id: "alice" },
            },
          ]),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      return new Response("unexpected", { status: 500 });
    }) as typeof fetch),
  );
  const token = await signGatewayToken("secret", {
    sub: "user-1",
    scope: "tray.read",
    iss: "tray-gateway",
    aud: "chatgpt",
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 600,
  });
  const r = await handler(req("/v1/items/contributed", { headers: { Authorization: `Bearer ${token}` } }));
  assertEquals(r.status, 200);
  const j = await r.json() as { items: Array<{ id: string; tray_owner_id?: string }> };
  assertEquals(j.items.length, 1);
  assertEquals(j.items[0].id, "i-theirs");
  assertEquals(j.items[0].tray_owner_id, "alice");
});

Deno.test("v1/items lists without tray filter", async () => {
  const store = new InMemoryStore();
  await store.upsertSupabaseAuth({
    userID: "user-1",
    accessToken: "sat",
    refreshToken: "srt",
    accessExpiresAtISO: "2099-01-01T00:00:00Z",
  });
  const handler = createHandler(
    deps(store, (async (input, init) => {
      const url = String(input);
      if (url.includes("/rest/v1/items?") && !url.includes("tray_id=eq.") && init) {
        assertEquals(url.includes("limit=100"), true);
        return new Response(
          JSON.stringify([{ id: "i1", tray_id: "t1", title: "Hi", status: "pending" }]),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      return new Response("unexpected", { status: 500 });
    }) as typeof fetch),
  );
  const token = await signGatewayToken("secret", {
    sub: "user-1",
    scope: "tray.read",
    iss: "tray-gateway",
    aud: "chatgpt",
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 600,
  });
  const r = await handler(req("/v1/items", { headers: { Authorization: `Bearer ${token}` } }));
  assertEquals(r.status, 200);
  const j = await r.json() as { items: Array<{ id: string }> };
  assertEquals(j.items.length, 1);
  assertEquals(j.items[0].id, "i1");
});

async function writeToken(scope = "tray.read tray.write") {
  return await signGatewayToken("secret", {
    sub: "user-1",
    scope,
    iss: "tray-gateway",
    aud: "chatgpt",
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 600,
  });
}

Deno.test("POST add item requires tray.write", async () => {
  const store = new InMemoryStore();
  await store.upsertSupabaseAuth({
    userID: "user-1",
    accessToken: "sat",
    refreshToken: "srt",
    accessExpiresAtISO: "2099-01-01T00:00:00Z",
  });
  const handler = createHandler(deps(store, (async () => new Response("unused")) as typeof fetch));
  const tid = "11111111-1111-4111-8111-111111111111";
  const readOnly = await writeToken("tray.read");
  const denied = await handler(
    req(`/v1/trays/${tid}/items`, {
      method: "POST",
      headers: { Authorization: `Bearer ${readOnly}`, "Content-Type": "application/json" },
      body: JSON.stringify({ title: "x" }),
    }),
  );
  assertEquals(denied.status, 403);
});

Deno.test("POST add item proxies PostgREST", async () => {
  const store = new InMemoryStore();
  await store.upsertSupabaseAuth({
    userID: "user-1",
    accessToken: "sat",
    refreshToken: "srt",
    accessExpiresAtISO: "2099-01-01T00:00:00Z",
  });
  const tid = "11111111-1111-4111-8111-111111111111";
  const handler = createHandler(
    deps(store, (async (input, init) => {
      const url = String(input);
      if (url.includes("/rest/v1/trays?") && url.includes("select=owner_id")) {
        return new Response(JSON.stringify([{ owner_id: "user-1" }]), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }
      const ri = init && typeof init === "object" ? init as RequestInit : undefined;
      if (url.includes("/rest/v1/items?") && ri?.method === "POST") {
        const body = JSON.parse(String(ri.body));
        assertEquals(body.title, "Fix bug");
        assertEquals(body.status, "accepted");
        return new Response(
          JSON.stringify([{ id: "new-1", tray_id: tid, title: "Fix bug", status: "accepted" }]),
          { status: 201, headers: { "Content-Type": "application/json" } },
        );
      }
      return new Response("unexpected", { status: 500 });
    }) as typeof fetch),
  );
  const tok = await writeToken();
  const r = await handler(
    req(`/v1/trays/${tid}/items`, {
      method: "POST",
      headers: { Authorization: `Bearer ${tok}`, "Content-Type": "application/json" },
      body: JSON.stringify({ title: "Fix bug" }),
    }),
  );
  assertEquals(r.status, 200);
  const j = await r.json() as { item: { id: string } };
  assertEquals(j.item.id, "new-1");
});

Deno.test("POST item complete sends PATCH", async () => {
  const store = new InMemoryStore();
  await store.upsertSupabaseAuth({
    userID: "user-1",
    accessToken: "sat",
    refreshToken: "srt",
    accessExpiresAtISO: "2099-01-01T00:00:00Z",
  });
  const iid = "22222222-2222-4222-8222-222222222222";
  const handler = createHandler(
    deps(store, (async (input, init) => {
      const url = String(input);
      const ri = init && typeof init === "object" ? init as RequestInit : undefined;
      if (url.includes("/rest/v1/items?") && url.includes(iid) && ri?.method === "PATCH") {
        assertEquals(JSON.parse(String(ri.body)).status, "completed");
        return new Response(null, { status: 204 });
      }
      return new Response("unexpected", { status: 500 });
    }) as typeof fetch),
  );
  const tok = await writeToken();
  const r = await handler(
    req(`/v1/items/${iid}/complete`, {
      method: "POST",
      headers: { Authorization: `Bearer ${tok}`, "Content-Type": "application/json" },
      body: JSON.stringify({ completion_message: "Shipped" }),
    }),
  );
  assertEquals(r.status, 200);
  assertEquals((await r.json()).ok, true);
});

Deno.test("v1/me requires valid bearer", async () => {
  const store = new InMemoryStore();
  const handler = createHandler(deps(store, (async () => new Response("unused")) as typeof fetch));
  const unauthorized = await handler(req("/v1/me"));
  assertEquals(unauthorized.status, 200);
  assertEquals((await unauthorized.json()).connected, false);

  const token = await signGatewayToken("secret", {
    sub: "user-1",
    scope: "tray.read",
    iss: "tray-gateway",
    aud: "chatgpt",
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 600,
  });
  const ok = await handler(req("/v1/me", { headers: { Authorization: `Bearer ${token}` } }));
  assertEquals(ok.status, 200);
  assertEquals((await ok.json()).user_id, "user-1");
});
