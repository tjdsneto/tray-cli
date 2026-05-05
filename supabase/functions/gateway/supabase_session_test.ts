import { assertEquals } from "jsr:@std/assert";
import { parsePkceExchangeBody } from "./supabase_session.ts";

Deno.test("parsePkceExchangeBody accepts full Supabase token response", () => {
  const out = parsePkceExchangeBody(
    JSON.stringify({
      access_token: "a",
      refresh_token: "r",
      expires_in: 99,
      user: { id: "f5e7e830-ac9b-483c-aaa5-0a158bf5ace3" },
    }),
  );
  assertEquals(out?.userId, "f5e7e830-ac9b-483c-aaa5-0a158bf5ace3");
  assertEquals(out?.accessToken, "a");
  assertEquals(out?.refreshToken, "r");
  assertEquals(out?.expiresIn, 99);
});

Deno.test("parsePkceExchangeBody defaults expires_in", () => {
  const out = parsePkceExchangeBody(
    JSON.stringify({
      access_token: "a",
      refresh_token: "r",
      user: { id: "u1" },
    }),
  );
  assertEquals(out?.expiresIn, 3600);
});

Deno.test("parsePkceExchangeBody rejects missing refresh", () => {
  const out = parsePkceExchangeBody(
    JSON.stringify({
      access_token: "a",
      user: { id: "u1" },
    }),
  );
  assertEquals(out, null);
});
