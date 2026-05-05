export type GatewayClaims = {
  sub: string;
  scope: string;
  iss: string;
  aud: string;
  exp: number;
  iat: number;
};

function b64url(input: Uint8Array): string {
  const raw = btoa(String.fromCharCode(...input));
  return raw.replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

function b64urlString(input: string): string {
  return b64url(new TextEncoder().encode(input));
}

function fromB64url(input: string): Uint8Array {
  const padded = input.replace(/-/g, "+").replace(/_/g, "/") + "=".repeat((4 - (input.length % 4)) % 4);
  const raw = atob(padded);
  const out = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i += 1) out[i] = raw.charCodeAt(i);
  return out;
}

async function hmacSign(secret: string, data: string): Promise<string> {
  const key = await crypto.subtle.importKey(
    "raw",
    new TextEncoder().encode(secret),
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign"],
  );
  const sig = await crypto.subtle.sign("HMAC", key, new TextEncoder().encode(data));
  return b64url(new Uint8Array(sig));
}

export async function signGatewayToken(secret: string, claims: GatewayClaims): Promise<string> {
  const header = b64urlString(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const payload = b64urlString(JSON.stringify(claims));
  const signingInput = `${header}.${payload}`;
  const sig = await hmacSign(secret, signingInput);
  return `${signingInput}.${sig}`;
}

export async function verifyGatewayToken(secret: string, token: string): Promise<GatewayClaims | null> {
  const parts = token.split(".");
  if (parts.length !== 3) return null;
  const [header, payload, sig] = parts;
  const expected = await hmacSign(secret, `${header}.${payload}`);
  if (sig !== expected) return null;
  let parsed: GatewayClaims;
  try {
    parsed = JSON.parse(new TextDecoder().decode(fromB64url(payload)));
  } catch {
    return null;
  }
  const now = Math.floor(Date.now() / 1000);
  if (!parsed.exp || parsed.exp <= now) return null;
  if (!parsed.sub || !parsed.iss || !parsed.aud) return null;
  return parsed;
}
