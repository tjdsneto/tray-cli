import { signGatewayToken, verifyGatewayToken } from "./jwt.ts";
import { buildAddItemBody, buildStatusPatch } from "./items_mutations.ts";
import {
  itemsContributedPath,
  itemsInsertPath,
  itemsListAllVisiblePath,
  itemsListPath,
  itemsPatchPath,
  trayMembersJoinedPath,
  trayOwnerSelectPath,
  traysListOwnedPath,
} from "./postgrest_paths.ts";
import type { PkceExchangeResult } from "./supabase_session.ts";
import { parsePkceExchangeBody, refreshSupabaseSession } from "./supabase_session.ts";
import type { GatewayStore } from "./store.ts";

export const corsHeaders: Record<string, string> = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers":
    "authorization, x-client-info, apikey, content-type",
  "Access-Control-Allow-Methods": "GET, POST, PATCH, DELETE, OPTIONS",
};

const UUID_PATH = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}";

export type Deps = {
  getEnv: (name: string) => string | undefined;
  fetchFn: typeof fetch;
  store: GatewayStore;
  now: () => Date;
  randomToken: () => string;
  randomUUID: () => string;
};

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { ...corsHeaders, "Content-Type": "application/json" },
  });
}

function readBearer(req: Request): string | null {
  const auth = req.headers.get("Authorization") ?? "";
  if (!auth.startsWith("Bearer ")) return null;
  return auth.slice("Bearer ".length).trim();
}

async function sha256Hex(input: string): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(input));
  return Array.from(new Uint8Array(digest)).map((b) => b.toString(16).padStart(2, "0")).join("");
}

function getRequiredEnv(deps: Deps, name: string): string {
  const value = deps.getEnv(name)?.trim();
  if (!value) throw new Error(`missing env: ${name}`);
  return value;
}

function parseScope(input: string | null): string {
  return (input ?? "").trim();
}

/** PKCE code_verifier: 32 random bytes, base64url (matches Go tray CLI). */
async function newCodeVerifier(): Promise<string> {
  const b = new Uint8Array(32);
  crypto.getRandomValues(b);
  let bin = "";
  for (let i = 0; i < b.length; i++) bin += String.fromCharCode(b[i]);
  return btoa(bin).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

async function codeChallengeS256(verifier: string): Promise<string> {
  const hash = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(verifier));
  const bytes = new Uint8Array(hash);
  let bin = "";
  for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i]);
  return btoa(bin).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

function supabaseAuthorizeURL(projectURL: string, provider: string, redirectTo: string, codeChallenge: string): string {
  const base = projectURL.replace(/\/$/, "") + "/auth/v1/authorize";
  const p = new URLSearchParams();
  p.set("provider", provider);
  p.set("redirect_to", redirectTo);
  p.set("code_challenge", codeChallenge);
  p.set("code_challenge_method", "s256");
  return `${base}?${p.toString()}`;
}

async function exchangePkce(
  deps: Deps,
  authCode: string,
  codeVerifier: string,
): Promise<PkceExchangeResult> {
  const base = getRequiredEnv(deps, "SUPABASE_URL").replace(/\/$/, "");
  const anon = getRequiredEnv(deps, "SUPABASE_ANON_KEY");
  const url = `${base}/auth/v1/token?grant_type=pkce`;
  const r = await deps.fetchFn(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      apikey: anon,
    },
    body: JSON.stringify({
      auth_code: authCode.trim(),
      code_verifier: codeVerifier,
    }),
  });
  const text = await r.text();
  if (!r.ok) {
    throw new Error(`pkce exchange ${r.status}: ${text}`);
  }
  const parsed = parsePkceExchangeBody(text);
  if (!parsed) throw new Error("pkce response missing tokens or user.id");
  return parsed;
}

function allowedClientId(clientID: string, deps: Deps): boolean {
  const raw = deps.getEnv("GATEWAY_ALLOWED_CLIENT_IDS")?.trim();
  const ids = raw && raw.length > 0
    ? raw.split(",").map((s) => s.trim()).filter(Boolean)
    : ["chatgpt-gateway-dev"];
  return ids.includes(clientID);
}

function allowedRedirectUri(redirectURI: string, deps: Deps): boolean {
  const raw = deps.getEnv("GATEWAY_ALLOWED_REDIRECT_PREFIXES")?.trim();
  const prefixes = raw && raw.length > 0
    ? raw.split(",").map((s) => s.trim()).filter(Boolean)
    : ["https://chat.openai.com/", "https://chatgpt.com/"];
  try {
    const u = new URL(redirectURI);
    if (u.protocol !== "https:") return false;
  } catch {
    return false;
  }
  return prefixes.some((pre) => redirectURI.startsWith(pre));
}

function buildRedirect(uri: string, code: string, state: string): string {
  const u = new URL(uri);
  u.searchParams.set("code", code);
  u.searchParams.set("state", state);
  return u.toString();
}

function buildOAuthErrorRedirect(uri: string, state: string, error: string): string {
  const u = new URL(uri);
  u.searchParams.set("error", error);
  u.searchParams.set("state", state);
  return u.toString();
}

async function mintAccessToken(deps: Deps, userID: string, scope: string): Promise<string> {
  const secret = getRequiredEnv(deps, "GATEWAY_JWT_SECRET");
  const issuer = getRequiredEnv(deps, "GATEWAY_TOKEN_ISSUER");
  const audience = getRequiredEnv(deps, "GATEWAY_TOKEN_AUDIENCE");
  const now = Math.floor(deps.now().getTime() / 1000);
  return signGatewayToken(secret, {
    sub: userID,
    scope,
    iss: issuer,
    aud: audience,
    iat: now,
    exp: now + 900,
  });
}

async function routeAuthorize(req: Request, deps: Deps): Promise<Response> {
  const url = new URL(req.url);
  const search = url.searchParams;
  const responseType = String(search.get("response_type") ?? "").trim();
  if (responseType !== "code") {
    return json({ error: "unsupported response_type" }, 400);
  }
  const clientID = String(search.get("client_id") ?? "").trim();
  const redirectURI = String(search.get("redirect_uri") ?? "").trim();
  const state = String(search.get("state") ?? "").trim();
  const scope = parseScope(search.get("scope"));
  if (!clientID || !redirectURI || !state) {
    return json({ error: "missing client_id, redirect_uri, or state" }, 400);
  }
  if (!allowedClientId(clientID, deps)) {
    return json({ error: "unauthorized_client" }, 400);
  }
  if (!allowedRedirectUri(redirectURI, deps)) {
    return json({ error: "invalid redirect_uri" }, 400);
  }

  const provider = (deps.getEnv("GATEWAY_OAUTH_PROVIDER") ?? "google").trim();
  const supabaseURL = getRequiredEnv(deps, "SUPABASE_URL").replace(/\/$/, "");
  const pendingId = deps.randomUUID();
  const verifier = await newCodeVerifier();
  const challenge = await codeChallengeS256(verifier);
  const expires = new Date(deps.now().getTime() + 10 * 60_000).toISOString();

  await deps.store.createOAuthPending({
    id: pendingId,
    codeVerifier: verifier,
    clientID,
    redirectURI,
    state,
    scope,
    expiresAtISO: expires,
  });

  const redirectTo = `${supabaseURL}/functions/v1/gateway/oauth/supabase-callback/${pendingId}`;
  const authURL = supabaseAuthorizeURL(supabaseURL, provider, redirectTo, challenge);
  return Response.redirect(authURL, 302);
}

async function routeSupabaseCallback(req: Request, deps: Deps, pendingId: string): Promise<Response> {
  const url = new URL(req.url);
  const pending = await deps.store.getOAuthPending(pendingId);
  if (!pending) {
    return json({ error: "invalid_or_expired_login_session" }, 400);
  }

  const oauthErr = url.searchParams.get("error");
  if (oauthErr) {
    await deps.store.deleteOAuthPending(pendingId);
    return Response.redirect(
      buildOAuthErrorRedirect(pending.redirectURI, pending.state, "access_denied"),
      302,
    );
  }

  const authCode = url.searchParams.get("code");
  if (!authCode?.trim()) {
    await deps.store.deleteOAuthPending(pendingId);
    return Response.redirect(
      buildOAuthErrorRedirect(pending.redirectURI, pending.state, "server_error"),
      302,
    );
  }

  let pkce: PkceExchangeResult;
  try {
    pkce = await exchangePkce(deps, authCode, pending.codeVerifier);
  } catch {
    await deps.store.deleteOAuthPending(pendingId);
    return Response.redirect(
      buildOAuthErrorRedirect(pending.redirectURI, pending.state, "server_error"),
      302,
    );
  }

  const accessExpiresAtISO = new Date(deps.now().getTime() + pkce.expiresIn * 1000).toISOString();
  try {
    await deps.store.upsertSupabaseAuth({
      userID: pkce.userId,
      accessToken: pkce.accessToken,
      refreshToken: pkce.refreshToken,
      accessExpiresAtISO,
    });
  } catch {
    await deps.store.deleteOAuthPending(pendingId);
    return Response.redirect(
      buildOAuthErrorRedirect(pending.redirectURI, pending.state, "server_error"),
      302,
    );
  }

  await deps.store.deleteOAuthPending(pendingId);

  const rawCode = deps.randomToken();
  const codeHash = await sha256Hex(rawCode);
  const expires = new Date(deps.now().getTime() + 5 * 60_000).toISOString();
  await deps.store.createOAuthCode({
    codeHash,
    userID: pkce.userId,
    clientID: pending.clientID,
    redirectURI: pending.redirectURI,
    scope: pending.scope,
    expiresAtISO: expires,
  });
  return Response.redirect(buildRedirect(pending.redirectURI, rawCode, pending.state), 302);
}

async function routeToken(req: Request, deps: Deps): Promise<Response> {
  const body = await req.formData();
  const grantType = String(body.get("grant_type") ?? "");
  const clientID = String(body.get("client_id") ?? "").trim();
  if (!clientID) return json({ error: "missing client_id" }, 400);

  if (grantType === "authorization_code") {
    const code = String(body.get("code") ?? "");
    const redirectURI = String(body.get("redirect_uri") ?? "");
    if (!code || !redirectURI) return json({ error: "missing code or redirect_uri" }, 400);
    const consumed = await deps.store.consumeOAuthCode(await sha256Hex(code), clientID, redirectURI);
    if (!consumed) return json({ error: "invalid_grant" }, 400);
    const accessToken = await mintAccessToken(deps, consumed.userID, consumed.scope);
    const refreshRaw = deps.randomToken();
    await deps.store.createRefreshToken({
      tokenHash: await sha256Hex(refreshRaw),
      userID: consumed.userID,
      clientID,
      scope: consumed.scope,
      expiresAtISO: new Date(deps.now().getTime() + 30 * 24 * 60 * 60_000).toISOString(),
    });
    return json({
      token_type: "Bearer",
      access_token: accessToken,
      expires_in: 900,
      refresh_token: refreshRaw,
      scope: consumed.scope,
    });
  }

  if (grantType === "refresh_token") {
    const refresh = String(body.get("refresh_token") ?? "");
    if (!refresh) return json({ error: "missing refresh_token" }, 400);
    const consumed = await deps.store.consumeRefreshToken(await sha256Hex(refresh), clientID);
    if (!consumed) return json({ error: "invalid_grant" }, 400);
    const accessToken = await mintAccessToken(deps, consumed.userID, consumed.scope);
    const nextRefreshRaw = deps.randomToken();
    await deps.store.createRefreshToken({
      tokenHash: await sha256Hex(nextRefreshRaw),
      userID: consumed.userID,
      clientID,
      scope: consumed.scope,
      expiresAtISO: new Date(deps.now().getTime() + 30 * 24 * 60 * 60_000).toISOString(),
    });
    return json({
      token_type: "Bearer",
      access_token: accessToken,
      expires_in: 900,
      refresh_token: nextRefreshRaw,
      scope: consumed.scope,
    });
  }

  return json({ error: "unsupported_grant_type" }, 400);
}

function scopeAllows(scope: string, required: string): boolean {
  return scope.split(/[\s,]+/).map((s) => s.trim()).filter(Boolean).includes(required);
}

async function ensureSupabaseUserAccess(deps: Deps, userId: string): Promise<string | null> {
  const row = await deps.store.getSupabaseAuth(userId);
  if (!row?.accessToken || !row.refreshToken || !row.accessExpiresAtISO) return null;
  const exp = Date.parse(row.accessExpiresAtISO);
  if (!Number.isFinite(exp)) return null;
  const skewMs = 120_000;
  if (exp > deps.now().getTime() + skewMs) return row.accessToken;
  const anon = getRequiredEnv(deps, "SUPABASE_ANON_KEY");
  const base = getRequiredEnv(deps, "SUPABASE_URL");
  const refreshed = await refreshSupabaseSession(deps.fetchFn, base, anon, row.refreshToken);
  if (!refreshed) return null;
  const nextExp = new Date(deps.now().getTime() + refreshed.expiresIn * 1000).toISOString();
  await deps.store.upsertSupabaseAuth({
    userID: userId,
    accessToken: refreshed.accessToken,
    refreshToken: refreshed.refreshToken,
    accessExpiresAtISO: nextExp,
  });
  return refreshed.accessToken;
}

async function postgrestGet(deps: Deps, userAccessToken: string, path: string): Promise<Response> {
  const base = getRequiredEnv(deps, "SUPABASE_URL").replace(/\/$/, "");
  const anon = getRequiredEnv(deps, "SUPABASE_ANON_KEY");
  return deps.fetchFn(`${base}${path}`, {
    headers: {
      apikey: anon,
      Authorization: `Bearer ${userAccessToken}`,
      Accept: "application/json",
    },
  });
}

async function postgrestSend(
  deps: Deps,
  userAccessToken: string,
  method: string,
  path: string,
  body?: unknown,
  extraHeaders?: Record<string, string>,
): Promise<Response> {
  const base = getRequiredEnv(deps, "SUPABASE_URL").replace(/\/$/, "");
  const anon = getRequiredEnv(deps, "SUPABASE_ANON_KEY");
  const headers: Record<string, string> = {
    apikey: anon,
    Authorization: `Bearer ${userAccessToken}`,
    Accept: "application/json",
    ...extraHeaders,
  };
  const init: RequestInit = { method, headers };
  if (body !== undefined && method !== "GET" && method !== "HEAD") {
    headers["Content-Type"] = "application/json";
    init.body = JSON.stringify(body);
  }
  return deps.fetchFn(`${base}${path}`, init);
}

async function readJsonObject(req: Request): Promise<Record<string, unknown> | null> {
  const ct = req.headers.get("content-type") ?? "";
  if (!ct.includes("application/json")) return {};
  const text = await req.text();
  if (!text.trim()) return {};
  try {
    const v = JSON.parse(text) as unknown;
    return v && typeof v === "object" && !Array.isArray(v) ? v as Record<string, unknown> : null;
  } catch {
    return null;
  }
}

function mutationErrorResponse(r: Response, text: string): Response {
  const st = r.status >= 400 && r.status < 500 ? r.status : 502;
  return json({ error: "mutation_failed", upstream_status: r.status, detail: text.slice(0, 800) }, st);
}

type WriteCtx =
  | { ok: true; uid: string; access: string }
  | { ok: false; response: Response };

async function requireWriteAccess(req: Request, deps: Deps): Promise<WriteCtx> {
  const token = readBearer(req);
  if (!token) return { ok: false, response: json({ error: "unauthorized" }, 401) };
  const claims = await verifyGatewayToken(getRequiredEnv(deps, "GATEWAY_JWT_SECRET"), token);
  if (!claims) return { ok: false, response: json({ error: "unauthorized" }, 401) };
  if (!scopeAllows(claims.scope, "tray.write")) {
    return {
      ok: false,
      response: json({ error: "insufficient_scope", detail: "OAuth scope must include tray.write" }, 403),
    };
  }
  const access = await ensureSupabaseUserAccess(deps, claims.sub);
  if (!access) {
    return {
      ok: false,
      response: json({ error: "supabase_session_missing", detail: "Complete OAuth connect again to sync tray access." }, 401),
    };
  }
  return { ok: true, uid: claims.sub, access };
}

async function fetchTrayOwnerId(deps: Deps, access: string, trayId: string): Promise<string | null> {
  const r = await postgrestGet(deps, access, trayOwnerSelectPath(trayId));
  const text = await r.text();
  if (!r.ok) return null;
  try {
    const rows = JSON.parse(text) as Array<{ owner_id?: string }>;
    if (!Array.isArray(rows) || !rows[0]?.owner_id) return null;
    return String(rows[0].owner_id);
  } catch {
    return null;
  }
}

function parseCreatedItemBody(text: string): unknown | null {
  try {
    const v = JSON.parse(text) as unknown;
    if (Array.isArray(v) && v[0]) return v[0];
    if (v && typeof v === "object" && v !== null && "id" in v) return v;
  } catch { /* ignore */ }
  return null;
}

function parseSnoozeIso(body: Record<string, unknown>): string | null {
  const raw = body.snooze_until ?? body.until;
  if (typeof raw !== "string" || !raw.trim()) return null;
  const d = new Date(raw.trim());
  if (Number.isNaN(d.getTime())) return null;
  return d.toISOString();
}

async function routeAddItem(req: Request, deps: Deps, trayId: string): Promise<Response> {
  const ctx = await requireWriteAccess(req, deps);
  if (!ctx.ok) return ctx.response;
  const body = await readJsonObject(req);
  if (body === null) return json({ error: "invalid_json" }, 400);
  const title = typeof body.title === "string" ? body.title : "";
  if (!title.trim()) return json({ error: "missing title" }, 400);
  const dueDate = typeof body.due_date === "string" ? body.due_date : undefined;
  const ownerId = await fetchTrayOwnerId(deps, ctx.access, trayId);
  if (!ownerId) return json({ error: "tray_not_found_or_inaccessible" }, 404);
  let insertBody: Record<string, unknown>;
  try {
    insertBody = buildAddItemBody(ctx.uid, trayId, title, ownerId, dueDate);
  } catch (e) {
    return json({ error: "invalid_request", detail: String(e) }, 400);
  }
  const r = await postgrestSend(deps, ctx.access, "POST", itemsInsertPath(), insertBody, {
    Prefer: "return=representation",
  });
  const text = await r.text();
  if (!r.ok) return mutationErrorResponse(r, text);
  const item = parseCreatedItemBody(text);
  if (!item) return json({ error: "invalid_upstream_response" }, 502);
  return json({ item });
}

async function routeDeleteItem(req: Request, deps: Deps, itemId: string): Promise<Response> {
  const ctx = await requireWriteAccess(req, deps);
  if (!ctx.ok) return ctx.response;
  const r = await postgrestSend(deps, ctx.access, "DELETE", itemsPatchPath(itemId), undefined, {
    Prefer: "return=minimal",
  });
  const text = await r.text();
  if (!r.ok) return mutationErrorResponse(r, text);
  return json({ ok: true, deleted_id: itemId });
}

async function routeItemTriage(
  req: Request,
  deps: Deps,
  itemId: string,
  action: string,
): Promise<Response> {
  const ctx = await requireWriteAccess(req, deps);
  if (!ctx.ok) return ctx.response;
  const body = await readJsonObject(req);
  if (body === null) return json({ error: "invalid_json" }, 400);

  let patch: Record<string, unknown>;
  switch (action) {
    case "accept": {
      patch = buildStatusPatch("accepted");
      break;
    }
    case "complete": {
      const msg = typeof body.completion_message === "string" ? body.completion_message.trim() : "";
      patch = msg
        ? buildStatusPatch("completed", { completion_message: msg })
        : buildStatusPatch("completed");
      break;
    }
    case "decline": {
      const reason = typeof body.reason === "string" ? body.reason.trim() : "";
      patch = reason
        ? buildStatusPatch("declined", { decline_reason: reason })
        : buildStatusPatch("declined");
      break;
    }
    case "archive": {
      patch = buildStatusPatch("archived");
      break;
    }
    case "snooze": {
      const until = parseSnoozeIso(body);
      if (!until) return json({ error: "missing_or_invalid_snooze_until", hint: "Use ISO-8601, e.g. 2026-05-10T15:00:00Z" }, 400);
      patch = buildStatusPatch("snoozed", { snooze_until: until });
      break;
    }
    default:
      return json({ error: "unknown_action" }, 400);
  }

  const r = await postgrestSend(deps, ctx.access, "PATCH", itemsPatchPath(itemId), patch, {
    Prefer: "return=minimal",
  });
  const text = await r.text();
  if (!r.ok) return mutationErrorResponse(r, text);
  return json({ ok: true, item_id: itemId, action });
}

type TrayRow = {
  id: string;
  owner_id: string;
  name: string;
  created_at: string;
  items?: Array<{ count?: number }>;
};

async function routeTrays(req: Request, deps: Deps): Promise<Response> {
  const token = readBearer(req);
  if (!token) return json({ error: "unauthorized" }, 401);
  const claims = await verifyGatewayToken(getRequiredEnv(deps, "GATEWAY_JWT_SECRET"), token);
  if (!claims) return json({ error: "unauthorized" }, 401);
  if (!scopeAllows(claims.scope, "tray.read")) return json({ error: "insufficient_scope" }, 403);
  const uid = claims.sub;
  const access = await ensureSupabaseUserAccess(deps, uid);
  if (!access) {
    return json({ error: "supabase_session_missing", detail: "Complete OAuth connect again to sync tray access." }, 401);
  }
  const r = await postgrestGet(deps, access, traysListOwnedPath(uid));
  const text = await r.text();
  if (!r.ok) return json({ error: "upstream_error", status: r.status }, 502);
  let rows: TrayRow[];
  try {
    rows = JSON.parse(text) as TrayRow[];
  } catch {
    return json({ error: "invalid_upstream_response" }, 502);
  }
  if (!Array.isArray(rows)) return json({ error: "invalid_upstream_response" }, 502);
  const trays = rows.map((t) => ({
    id: t.id,
    name: t.name,
    owner_id: t.owner_id,
    is_owner: t.owner_id === uid,
    created_at: t.created_at,
    item_count: Array.isArray(t.items) && t.items[0] && typeof t.items[0].count === "number" ? t.items[0].count : 0,
  }));
  return json({ trays });
}

function clampInt(n: number, min: number, max: number): number {
  if (!Number.isFinite(n)) return min;
  return Math.min(max, Math.max(min, Math.floor(n)));
}

type TrayMemberJoinRow = {
  joined_at: string;
  trays?: TrayRow | null;
};

async function routeRemotes(req: Request, deps: Deps): Promise<Response> {
  const token = readBearer(req);
  if (!token) return json({ error: "unauthorized" }, 401);
  const claims = await verifyGatewayToken(getRequiredEnv(deps, "GATEWAY_JWT_SECRET"), token);
  if (!claims) return json({ error: "unauthorized" }, 401);
  if (!scopeAllows(claims.scope, "tray.read")) return json({ error: "insufficient_scope" }, 403);
  const uid = claims.sub;
  const access = await ensureSupabaseUserAccess(deps, uid);
  if (!access) {
    return json({ error: "supabase_session_missing", detail: "Complete OAuth connect again to sync tray access." }, 401);
  }
  const r = await postgrestGet(deps, access, trayMembersJoinedPath(uid));
  const text = await r.text();
  if (!r.ok) return json({ error: "upstream_error", status: r.status }, 502);
  let rows: TrayMemberJoinRow[];
  try {
    rows = JSON.parse(text) as TrayMemberJoinRow[];
  } catch {
    return json({ error: "invalid_upstream_response" }, 502);
  }
  if (!Array.isArray(rows)) return json({ error: "invalid_upstream_response" }, 502);
  const remotes = rows
    .filter((row) => row.trays && String(row.trays.owner_id) !== uid)
    .map((row) => {
      const t = row.trays!;
      return {
        id: t.id,
        name: t.name,
        owner_id: t.owner_id,
        joined_at: row.joined_at,
        is_owner: false,
        item_count: Array.isArray(t.items) && t.items[0] && typeof t.items[0].count === "number" ? t.items[0].count : 0,
      };
    });
  return json({ remotes });
}

type ItemOutboxRow = {
  trays?: { owner_id?: string } | null;
  [key: string]: unknown;
};

async function routeItemsContributed(req: Request, deps: Deps, url: URL): Promise<Response> {
  const token = readBearer(req);
  if (!token) return json({ error: "unauthorized" }, 401);
  const claims = await verifyGatewayToken(getRequiredEnv(deps, "GATEWAY_JWT_SECRET"), token);
  if (!claims) return json({ error: "unauthorized" }, 401);
  if (!scopeAllows(claims.scope, "tray.read")) return json({ error: "insufficient_scope" }, 403);
  const uid = claims.sub;
  const access = await ensureSupabaseUserAccess(deps, uid);
  if (!access) {
    return json({ error: "supabase_session_missing", detail: "Complete OAuth connect again to sync tray access." }, 401);
  }
  const limitRaw = url.searchParams.get("limit");
  const limit = clampInt(limitRaw ? Number.parseInt(limitRaw, 10) : 100, 1, 200);
  const status = url.searchParams.get("status")?.trim() || undefined;
  const path = itemsContributedPath({ userId: uid, status, limit });
  const r = await postgrestGet(deps, access, path);
  const text = await r.text();
  if (!r.ok) return json({ error: "upstream_error", status: r.status }, 502);
  let rows: ItemOutboxRow[];
  try {
    rows = JSON.parse(text) as ItemOutboxRow[];
  } catch {
    return json({ error: "invalid_upstream_response" }, 502);
  }
  if (!Array.isArray(rows)) return json({ error: "invalid_upstream_response" }, 502);
  const items = rows
    .filter((row) => {
      const owner = row.trays?.owner_id;
      return typeof owner === "string" && owner.trim() !== "" && owner !== uid;
    })
    .map((row) => {
      const { trays: _t, ...rest } = row;
      const ownerId = row.trays?.owner_id;
      return { ...rest, tray_owner_id: ownerId };
    });
  return json({ items });
}

async function routeItemsAll(req: Request, deps: Deps, url: URL): Promise<Response> {
  const token = readBearer(req);
  if (!token) return json({ error: "unauthorized" }, 401);
  const claims = await verifyGatewayToken(getRequiredEnv(deps, "GATEWAY_JWT_SECRET"), token);
  if (!claims) return json({ error: "unauthorized" }, 401);
  if (!scopeAllows(claims.scope, "tray.read")) return json({ error: "insufficient_scope" }, 403);
  const access = await ensureSupabaseUserAccess(deps, claims.sub);
  if (!access) {
    return json({ error: "supabase_session_missing", detail: "Complete OAuth connect again to sync tray access." }, 401);
  }
  const limitRaw = url.searchParams.get("limit");
  const limit = clampInt(limitRaw ? Number.parseInt(limitRaw, 10) : 100, 1, 200);
  const status = url.searchParams.get("status")?.trim() || undefined;
  const trayId = url.searchParams.get("tray_id")?.trim() || undefined;
  const path = itemsListAllVisiblePath({ status, trayId, limit });
  const r = await postgrestGet(deps, access, path);
  const text = await r.text();
  if (!r.ok) return json({ error: "upstream_error", status: r.status }, 502);
  let items: unknown[];
  try {
    items = JSON.parse(text) as unknown[];
  } catch {
    return json({ error: "invalid_upstream_response" }, 502);
  }
  if (!Array.isArray(items)) return json({ error: "invalid_upstream_response" }, 502);
  return json({ items });
}

async function routeTrayItems(req: Request, deps: Deps, trayId: string, url: URL): Promise<Response> {
  const token = readBearer(req);
  if (!token) return json({ error: "unauthorized" }, 401);
  const claims = await verifyGatewayToken(getRequiredEnv(deps, "GATEWAY_JWT_SECRET"), token);
  if (!claims) return json({ error: "unauthorized" }, 401);
  if (!scopeAllows(claims.scope, "tray.read")) return json({ error: "insufficient_scope" }, 403);
  const access = await ensureSupabaseUserAccess(deps, claims.sub);
  if (!access) {
    return json({ error: "supabase_session_missing", detail: "Complete OAuth connect again to sync tray access." }, 401);
  }
  const limitRaw = url.searchParams.get("limit");
  const limit = clampInt(limitRaw ? Number.parseInt(limitRaw, 10) : 50, 1, 100);
  const status = url.searchParams.get("status")?.trim() || undefined;
  const path = itemsListPath({ trayId, status, limit });
  const r = await postgrestGet(deps, access, path);
  const text = await r.text();
  if (!r.ok) return json({ error: "upstream_error", status: r.status }, 502);
  let items: unknown[];
  try {
    items = JSON.parse(text) as unknown[];
  } catch {
    return json({ error: "invalid_upstream_response" }, 502);
  }
  if (!Array.isArray(items)) return json({ error: "invalid_upstream_response" }, 502);
  return json({ tray_id: trayId, items });
}

async function routeMe(req: Request, deps: Deps): Promise<Response> {
  const token = readBearer(req);
  // NOTE: Return 200 even when not connected. Some Action runtimes treat non-2xx as a tool failure
  // and hide the body, which makes debugging OAuth wiring painful.
  if (!token) return json({ connected: false }, 200);
  const claims = await verifyGatewayToken(getRequiredEnv(deps, "GATEWAY_JWT_SECRET"), token);
  if (!claims) return json({ connected: false }, 200);
  return json({ user_id: claims.sub, scope: claims.scope, issuer: claims.iss, audience: claims.aud });
}

export function createHandler(deps: Deps) {
  return async (req: Request): Promise<Response> => {
    if (req.method === "OPTIONS") return new Response("ok", { headers: corsHeaders });
    const url = new URL(req.url);

    try {
      const cbMatch = url.pathname.match(/\/oauth\/supabase-callback\/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\/?$/i);
      if (req.method === "GET" && cbMatch) {
        return await routeSupabaseCallback(req, deps, cbMatch[1]);
      }

      if (req.method === "GET" && url.pathname.endsWith("/oauth/authorize")) {
        return await routeAuthorize(req, deps);
      }
      if (req.method === "POST" && url.pathname.endsWith("/oauth/token")) return await routeToken(req, deps);

      if (req.method === "POST") {
        const addMatch = url.pathname.match(
          new RegExp(`/v1/trays/(${UUID_PATH})/items/?$`, "i"),
        );
        if (addMatch) return await routeAddItem(req, deps, addMatch[1]);
        const triageMatch = url.pathname.match(
          new RegExp(`/v1/items/(${UUID_PATH})/(complete|accept|decline|snooze|archive)/?$`, "i"),
        );
        if (triageMatch) {
          return await routeItemTriage(req, deps, triageMatch[1], triageMatch[2].toLowerCase());
        }
      }

      if (req.method === "DELETE") {
        const delMatch = url.pathname.match(new RegExp(`/v1/items/(${UUID_PATH})/?$`, "i"));
        if (delMatch) return await routeDeleteItem(req, deps, delMatch[1]);
      }

      if (req.method === "GET") {
        if (url.pathname.endsWith("/v1/items/contributed")) {
          return await routeItemsContributed(req, deps, url);
        }
        if (url.pathname.endsWith("/v1/items")) return await routeItemsAll(req, deps, url);
        const itemsMatch = url.pathname.match(
          /\/v1\/trays\/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\/items\/?$/i,
        );
        if (itemsMatch) return await routeTrayItems(req, deps, itemsMatch[1], url);
        if (url.pathname.endsWith("/v1/remotes")) return await routeRemotes(req, deps);
        if (url.pathname.endsWith("/v1/trays")) return await routeTrays(req, deps);
      }
      if (req.method === "GET" && url.pathname.endsWith("/v1/me")) return await routeMe(req, deps);
      if (req.method === "GET" && url.pathname.endsWith("/health")) return json({ ok: true, service: "gateway" });
      return json({ error: "not_found" }, 404);
    } catch (err) {
      return json({ error: "internal_error", detail: String(err) }, 500);
    }
  };
}
