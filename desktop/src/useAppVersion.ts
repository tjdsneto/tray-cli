import { invoke } from "@tauri-apps/api/core";
import { useEffect, useState } from "react";

export function useAppVersion(): string | null {
  const [version, setVersion] = useState<string | null>(null);
  useEffect(() => {
    invoke<string>("app_version")
      .then(setVersion)
      .catch(() => setVersion(null));
  }, []);
  return version;
}
