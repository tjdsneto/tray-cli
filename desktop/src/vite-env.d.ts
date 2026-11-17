/// <reference types="vite/client" />

interface TrayApiConfig {
  base: string;
  token: string;
}

interface Window {
  __TRAY_API__?: TrayApiConfig;
}
