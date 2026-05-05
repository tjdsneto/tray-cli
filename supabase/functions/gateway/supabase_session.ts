/** Parsed Supabase Auth token response (PKCE or refresh). */
export type SupabaseAuthTokens = {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
};

export type PkceExchangeResult = {
  userId: string;
} & SupabaseAuthTokens;

export function parsePkceExchangeBody(text: string): PkceExchangeResult | null {
  let body: unknown;
  try {
    body = JSON.parse(text);
  } catch {
    return null;
  }
  if (!body || typeof body !== "object") return null;
  const o = body as Record<string, unknown>;
  const user = o.user;
  const userId = user && typeof user === "object" && "id" in user
    ? String((user as { id?: unknown }).id ?? "").trim()
    : "";
  const accessToken = String(o.access_token ?? "").trim();
  const refreshToken = String(o.refresh_token ?? "").trim();
  const expiresInRaw = o.expires_in;
  const expiresIn = typeof expiresInRaw === "number" && Number.isFinite(expiresInRaw)
    ? expiresInRaw
    : 3600;
  if (!userId || !accessToken || !refreshToken) return null;
  return { userId, accessToken, refreshToken, expiresIn };
}

export async function refreshSupabaseSession(
  fetchFn: typeof fetch,
  supabaseURL: string,
  anonKey: string,
  refreshToken: string,
): Promise<SupabaseAuthTokens | null> {
  const base = supabaseURL.replace(/\/$/, "");
  const url = `${base}/auth/v1/token?grant_type=refresh_token`;
  const r = await fetchFn(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      apikey: anonKey,
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
  const text = await r.text();
  if (!r.ok) return null;
  let body: unknown;
  try {
    body = JSON.parse(text);
  } catch {
    return null;
  }
  if (!body || typeof body !== "object") return null;
  const o = body as Record<string, unknown>;
  const accessToken = String(o.access_token ?? "").trim();
  const nextRefresh = String(o.refresh_token ?? "").trim() || refreshToken;
  const expiresInRaw = o.expires_in;
  const expiresIn = typeof expiresInRaw === "number" && Number.isFinite(expiresInRaw)
    ? expiresInRaw
    : 3600;
  if (!accessToken) return null;
  return { accessToken, refreshToken: nextRefresh, expiresIn };
}
