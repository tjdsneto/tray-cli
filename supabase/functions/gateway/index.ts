import "jsr:@supabase/functions-js/edge-runtime.d.ts";
import { createHandler } from "./handler.ts";
import { PostgrestGatewayStore } from "./store.ts";

const supabaseURL = (Deno.env.get("SUPABASE_URL") ?? "").replace(/\/$/, "");
const serviceRoleKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY") ?? "";

const handler = createHandler({
  getEnv: (name: string) => Deno.env.get(name),
  fetchFn: fetch,
  now: () => new Date(),
  randomToken: () => crypto.randomUUID().replace(/-/g, "") + crypto.randomUUID().replace(/-/g, ""),
  randomUUID: () => crypto.randomUUID(),
  store: new PostgrestGatewayStore({
    fetchFn: fetch,
    supabaseURL,
    serviceRoleKey,
  }),
});

Deno.serve(handler);
