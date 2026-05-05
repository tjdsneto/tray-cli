export type OAuthCodeRecord = {
  codeHash: string;
  userID: string;
  clientID: string;
  redirectURI: string;
  scope: string;
  expiresAtISO: string;
};

export type RefreshTokenRecord = {
  tokenHash: string;
  userID: string;
  clientID: string;
  scope: string;
  expiresAtISO: string;
};

export type ConsumedCode = {
  userID: string;
  scope: string;
  clientID: string;
};

export type ConsumedRefresh = {
  userID: string;
  scope: string;
  clientID: string;
};

export type OAuthPendingRecord = {
  id: string;
  codeVerifier: string;
  clientID: string;
  redirectURI: string;
  state: string;
  scope: string;
  expiresAtISO: string;
};

export type SupabaseAuthRecord = {
  accessToken: string;
  refreshToken: string;
  accessExpiresAtISO: string;
};

export type SupabaseAuthUpsert = {
  userID: string;
  accessToken: string;
  refreshToken: string;
  accessExpiresAtISO: string;
};

export interface GatewayStore {
  createOAuthCode(record: OAuthCodeRecord): Promise<void>;
  consumeOAuthCode(codeHash: string, clientID: string, redirectURI: string): Promise<ConsumedCode | null>;
  createRefreshToken(record: RefreshTokenRecord): Promise<void>;
  consumeRefreshToken(tokenHash: string, clientID: string): Promise<ConsumedRefresh | null>;
  createOAuthPending(record: OAuthPendingRecord): Promise<void>;
  getOAuthPending(id: string): Promise<OAuthPendingRecord | null>;
  deleteOAuthPending(id: string): Promise<void>;
  upsertSupabaseAuth(record: SupabaseAuthUpsert): Promise<void>;
  getSupabaseAuth(userID: string): Promise<SupabaseAuthRecord | null>;
}

type StoreDeps = {
  fetchFn: typeof fetch;
  supabaseURL: string;
  serviceRoleKey: string;
};

function jsonHeaders(serviceRoleKey: string): HeadersInit {
  return {
    apikey: serviceRoleKey,
    Authorization: `Bearer ${serviceRoleKey}`,
    "Content-Type": "application/json",
  };
}

export class PostgrestGatewayStore implements GatewayStore {
  constructor(private readonly deps: StoreDeps) {}

  async createOAuthCode(record: OAuthCodeRecord): Promise<void> {
    const r = await this.deps.fetchFn(`${this.deps.supabaseURL}/rest/v1/gateway_oauth_codes`, {
      method: "POST",
      headers: jsonHeaders(this.deps.serviceRoleKey),
      body: JSON.stringify({
        code_hash: record.codeHash,
        user_id: record.userID,
        client_id: record.clientID,
        redirect_uri: record.redirectURI,
        scope: record.scope,
        expires_at: record.expiresAtISO,
      }),
    });
    if (!r.ok) throw new Error(`store createOAuthCode failed: ${r.status}`);
  }

  async consumeOAuthCode(codeHash: string, clientID: string, redirectURI: string): Promise<ConsumedCode | null> {
    const u = new URL(`${this.deps.supabaseURL}/rest/v1/gateway_oauth_codes`);
    u.searchParams.set("select", "id,user_id,scope,client_id,redirect_uri,expires_at,consumed_at");
    u.searchParams.set("code_hash", `eq.${codeHash}`);
    u.searchParams.set("client_id", `eq.${clientID}`);
    u.searchParams.set("redirect_uri", `eq.${redirectURI}`);
    u.searchParams.set("limit", "1");
    const r = await this.deps.fetchFn(u.toString(), { headers: jsonHeaders(this.deps.serviceRoleKey) });
    if (!r.ok) throw new Error(`store consumeOAuthCode select failed: ${r.status}`);
    const rows = await r.json() as Array<Record<string, string | null>>;
    if (rows.length === 0) return null;
    const row = rows[0];
    const exp = row.expires_at ? Date.parse(String(row.expires_at)) : 0;
    const consumed = row.consumed_at ? Date.parse(String(row.consumed_at)) : 0;
    if (!exp || exp <= Date.now() || consumed) return null;
    const id = String(row.id);
    const mark = await this.deps.fetchFn(`${this.deps.supabaseURL}/rest/v1/gateway_oauth_codes?id=eq.${id}`, {
      method: "PATCH",
      headers: {
        ...jsonHeaders(this.deps.serviceRoleKey),
        Prefer: "return=minimal",
      },
      body: JSON.stringify({ consumed_at: new Date().toISOString() }),
    });
    if (!mark.ok) throw new Error(`store consumeOAuthCode mark failed: ${mark.status}`);
    return {
      userID: String(row.user_id),
      scope: String(row.scope ?? ""),
      clientID: String(row.client_id),
    };
  }

  async createRefreshToken(record: RefreshTokenRecord): Promise<void> {
    const r = await this.deps.fetchFn(`${this.deps.supabaseURL}/rest/v1/gateway_refresh_tokens`, {
      method: "POST",
      headers: jsonHeaders(this.deps.serviceRoleKey),
      body: JSON.stringify({
        token_hash: record.tokenHash,
        user_id: record.userID,
        client_id: record.clientID,
        scope: record.scope,
        expires_at: record.expiresAtISO,
      }),
    });
    if (!r.ok) throw new Error(`store createRefreshToken failed: ${r.status}`);
  }

  async consumeRefreshToken(tokenHash: string, clientID: string): Promise<ConsumedRefresh | null> {
    const u = new URL(`${this.deps.supabaseURL}/rest/v1/gateway_refresh_tokens`);
    u.searchParams.set("select", "id,user_id,scope,client_id,expires_at,revoked_at");
    u.searchParams.set("token_hash", `eq.${tokenHash}`);
    u.searchParams.set("client_id", `eq.${clientID}`);
    u.searchParams.set("limit", "1");
    const r = await this.deps.fetchFn(u.toString(), { headers: jsonHeaders(this.deps.serviceRoleKey) });
    if (!r.ok) throw new Error(`store consumeRefreshToken select failed: ${r.status}`);
    const rows = await r.json() as Array<Record<string, string | null>>;
    if (rows.length === 0) return null;
    const row = rows[0];
    const exp = row.expires_at ? Date.parse(String(row.expires_at)) : 0;
    const revoked = row.revoked_at ? Date.parse(String(row.revoked_at)) : 0;
    if (!exp || exp <= Date.now() || revoked) return null;
    const id = String(row.id);
    const mark = await this.deps.fetchFn(`${this.deps.supabaseURL}/rest/v1/gateway_refresh_tokens?id=eq.${id}`, {
      method: "PATCH",
      headers: {
        ...jsonHeaders(this.deps.serviceRoleKey),
        Prefer: "return=minimal",
      },
      body: JSON.stringify({ revoked_at: new Date().toISOString() }),
    });
    if (!mark.ok) throw new Error(`store consumeRefreshToken mark failed: ${mark.status}`);
    return {
      userID: String(row.user_id),
      scope: String(row.scope ?? ""),
      clientID: String(row.client_id),
    };
  }

  async createOAuthPending(record: OAuthPendingRecord): Promise<void> {
    const r = await this.deps.fetchFn(`${this.deps.supabaseURL}/rest/v1/gateway_oauth_pending`, {
      method: "POST",
      headers: jsonHeaders(this.deps.serviceRoleKey),
      body: JSON.stringify({
        id: record.id,
        code_verifier: record.codeVerifier,
        client_id: record.clientID,
        redirect_uri: record.redirectURI,
        state: record.state,
        scope: record.scope,
        expires_at: record.expiresAtISO,
      }),
    });
    if (!r.ok) throw new Error(`store createOAuthPending failed: ${r.status}`);
  }

  async getOAuthPending(id: string): Promise<OAuthPendingRecord | null> {
    const u = new URL(`${this.deps.supabaseURL}/rest/v1/gateway_oauth_pending`);
    u.searchParams.set("select", "id,code_verifier,client_id,redirect_uri,state,scope,expires_at");
    u.searchParams.set("id", `eq.${id}`);
    u.searchParams.set("limit", "1");
    const r = await this.deps.fetchFn(u.toString(), { headers: jsonHeaders(this.deps.serviceRoleKey) });
    if (!r.ok) throw new Error(`store getOAuthPending failed: ${r.status}`);
    const rows = await r.json() as Array<Record<string, string | null>>;
    if (rows.length === 0) return null;
    const row = rows[0];
    const exp = row.expires_at ? Date.parse(String(row.expires_at)) : 0;
    if (!exp || exp <= Date.now()) return null;
    return {
      id: String(row.id),
      codeVerifier: String(row.code_verifier),
      clientID: String(row.client_id),
      redirectURI: String(row.redirect_uri),
      state: String(row.state ?? ""),
      scope: String(row.scope ?? ""),
      expiresAtISO: String(row.expires_at),
    };
  }

  async deleteOAuthPending(id: string): Promise<void> {
    const r = await this.deps.fetchFn(`${this.deps.supabaseURL}/rest/v1/gateway_oauth_pending?id=eq.${id}`, {
      method: "DELETE",
      headers: jsonHeaders(this.deps.serviceRoleKey),
    });
    if (!r.ok) throw new Error(`store deleteOAuthPending failed: ${r.status}`);
  }

  async upsertSupabaseAuth(record: SupabaseAuthUpsert): Promise<void> {
    const r = await this.deps.fetchFn(`${this.deps.supabaseURL}/rest/v1/gateway_user_supabase_auth`, {
      method: "POST",
      headers: {
        ...jsonHeaders(this.deps.serviceRoleKey),
        Prefer: "resolution=merge-duplicates",
      },
      body: JSON.stringify({
        user_id: record.userID,
        access_token: record.accessToken,
        refresh_token: record.refreshToken,
        access_expires_at: record.accessExpiresAtISO,
        updated_at: new Date().toISOString(),
      }),
    });
    if (!r.ok) {
      const t = await r.text();
      throw new Error(`store upsertSupabaseAuth failed: ${r.status} ${t}`);
    }
  }

  async getSupabaseAuth(userID: string): Promise<SupabaseAuthRecord | null> {
    const u = new URL(`${this.deps.supabaseURL}/rest/v1/gateway_user_supabase_auth`);
    u.searchParams.set("select", "access_token,refresh_token,access_expires_at");
    u.searchParams.set("user_id", `eq.${userID}`);
    u.searchParams.set("limit", "1");
    const r = await this.deps.fetchFn(u.toString(), { headers: jsonHeaders(this.deps.serviceRoleKey) });
    if (!r.ok) throw new Error(`store getSupabaseAuth failed: ${r.status}`);
    const rows = await r.json() as Array<Record<string, string | null>>;
    if (rows.length === 0) return null;
    const row = rows[0];
    return {
      accessToken: String(row.access_token ?? ""),
      refreshToken: String(row.refresh_token ?? ""),
      accessExpiresAtISO: String(row.access_expires_at ?? ""),
    };
  }
}
