import type { CSSProperties } from "react";

export type ThemeName = "dark" | "light" | "solarized-light";

type ThemePalette = {
  colorScheme: "dark" | "light";
  vars: Record<string, string>;
};

export const themeOptions: Array<{ value: ThemeName; label: string; description: string }> = [
  { value: "dark", label: "黑夜", description: "低眩光深底，适合长时间盯日志。" },
  { value: "light", label: "白天", description: "中性浅底，信息层次更干净。" },
  { value: "solarized-light", label: "Solarized Light", description: "柔和米色底，阅读压力更低。" },
];

const palettes: Record<ThemeName, ThemePalette> = {
  dark: {
    colorScheme: "dark",
    vars: {
      "--bg": "#111113", "--bg-elevated": "#18191b", "--bg-muted": "#212225", "--surface-card": "#18191b",
      "--surface-card-muted": "#212225", "--surface-card-strong": "#272a2d", "--surface-sidebar": "#18191b",
      "--surface-sidebar-alt": "#212225", "--surface-menu": "#18191b", "--surface-input": "#18191b",
      "--surface-input-strong": "#111113", "--surface-hover": "rgba(255, 255, 255, 0.05)",
      "--surface-hover-strong": "#272a2d", "--surface-hover-accent": "rgba(0, 162, 199, 0.14)",
      "--surface-overlay": "rgba(5, 8, 14, 0.72)", "--border": "#2e3135", "--border-strong": "#363a3f",
      "--border-hover": "#5a6169", "--border-subtle": "#212225", "--text": "#edeef0", "--text-strong": "#f8f9fb",
      "--text-secondary": "#b0b4ba", "--text-muted": "#777b84", "--text-dim": "#5a6169", "--placeholder": "#777b84",
      "--accent": "#11809c", "--accent-hover": "#00a2c7", "--accent-fg": "#4ccce6",
      "--accent-soft": "rgba(0, 162, 199, 0.18)", "--accent-contrast": "#f7fdff", "--success": "#30a46c",
      "--info": "#70b8ff", "--warn": "#ffca16", "--error": "#ec5d5e", "--shadow-dialog": "0 24px 60px rgba(0, 0, 0, 0.45)",
      "--shadow-menu": "0 14px 30px rgba(0, 0, 0, 0.45)", "--row": "#18191b", "--row-info-bg": "#111113",
      "--row-head-border": "#272a2d", "--row-head-text": "#777b84", "--row-border": "#212225", "--row-hover": "#1c2024",
      "--row-marker": "#70b8ff", "--row-current-marker": "#4ccce6", "--row-match-marker": "#ffca16",
      "--row-selected": "rgba(58, 158, 255, 0.16)", "--row-current": "rgba(0, 162, 199, 0.12)",
      "--row-match": "rgba(255, 197, 61, 0.12)", "--tag-text": "#b0b4ba", "--time-text": "#777b84",
      "--message-text": "#edeef0", "--search-hit-bg": "rgba(255, 197, 61, 0.24)", "--search-hit-text": "#ffe7b3",
      "--chip-outline": "rgba(255, 255, 255, 0.06)", "--chip-info-bg": "#111927", "--chip-info-text": "#70b8ff",
      "--chip-warn-bg": "#302008", "--chip-warn-text": "#ffca16", "--chip-error-bg": "#3b1219", "--chip-error-text": "#ff9592",
      "--url-underline": "rgba(112, 184, 255, 0.36)", "--json-key": "#70b8ff", "--json-string": "#3dd68c",
      "--json-number": "#ffca16", "--json-boolean": "#4ccce6", "--json-null": "#777b84", "--token-url": "#70b8ff",
      "--token-error": "#ff9592", "--token-metric": "#ffca16", "--token-method": "#3dd68c", "--token-path": "#edeef0",
      "--token-stack": "#b0b4ba", "--dialog-card-bg": "#18191b", "--dialog-section-bg": "#212225",
      "--dialog-body-bg": "#18191b", "--dialog-sidebar-bg": "#18191b", "--dialog-sidebar-border": "#363a3f",
      "--dialog-nav-bg": "#111113", "--dialog-nav-border": "#2e3135", "--dialog-nav-hover-bg": "#212225",
      "--dialog-nav-hover-border": "#5a6169", "--dialog-nav-text": "#edeef0", "--dialog-nav-meta": "#b0b4ba",
      "--dialog-badge-bg": "rgba(58, 158, 255, 0.18)", "--dialog-badge-text": "#c2e6ff", "--dialog-main-bg": "#212225",
      "--dialog-main-header-bg": "#18191b", "--dialog-main-header-border": "#363a3f", "--dialog-empty-text": "#777b84",
      "--dialog-code-bg": "#111113", "--dialog-rule-bg": "#111113", "--dialog-rule-title": "#edeef0",
      "--dialog-rule-join-bg": "#212225", "--dialog-rule-join-border": "#363a3f", "--dialog-rule-join-text": "#b0b4ba",
      "--dialog-danger-bg": "rgba(236, 93, 94, 0.12)", "--dialog-danger-border": "rgba(236, 93, 94, 0.32)",
      "--dialog-danger-text": "#ff9592", "--switch-off": "#43484e", "--switch-on": "#11809c", "--switch-thumb": "#f7f8fa",
      "--status-bg": "#111113", "--status-border": "#212225", "--detail-resize-hover": "rgba(0, 162, 199, 0.18)",
      "--empty-circle-fg": "#777b84",
    },
  },
  light: {
    colorScheme: "light",
    vars: {
      "--bg": "#fcfcfd", "--bg-elevated": "#f9f9fb", "--bg-muted": "#f0f0f3", "--surface-card": "#f9f9fb",
      "--surface-card-muted": "#f0f0f3", "--surface-card-strong": "#e8e8ec", "--surface-sidebar": "#f9f9fb",
      "--surface-sidebar-alt": "#f4faff", "--surface-menu": "#fcfcfd", "--surface-input": "#ffffff",
      "--surface-input-strong": "#f4faff", "--surface-hover": "rgba(17, 50, 100, 0.04)", "--surface-hover-strong": "#e8e8ec",
      "--surface-hover-accent": "rgba(13, 116, 206, 0.12)", "--surface-overlay": "rgba(78, 89, 118, 0.24)",
      "--border": "#d9d9e0", "--border-strong": "#cdced6", "--border-hover": "#8b8d98", "--border-subtle": "#e8e8ec",
      "--text": "#1c2024", "--text-strong": "#0f1720", "--text-secondary": "#60646c", "--text-muted": "#80838d",
      "--text-dim": "#8b8d98", "--placeholder": "#8b8d98", "--accent": "#0d74ce", "--accent-hover": "#0588f0",
      "--accent-fg": "#0d74ce", "--accent-soft": "rgba(13, 116, 206, 0.12)", "--accent-contrast": "#ffffff",
      "--success": "#218358", "--info": "#0588f0", "--warn": "#ab6400", "--error": "#ce2c31",
      "--shadow-dialog": "0 24px 60px rgba(28, 32, 36, 0.12)", "--shadow-menu": "0 14px 30px rgba(28, 32, 36, 0.12)",
      "--row": "#f0f0f3", "--row-info-bg": "#fcfcfd", "--row-head-border": "#e0e1e6", "--row-head-text": "#60646c",
      "--row-border": "#e8e8ec", "--row-hover": "#f4faff", "--row-marker": "#0588f0", "--row-current-marker": "#0d74ce",
      "--row-match-marker": "#e2a336", "--row-selected": "rgba(13, 116, 206, 0.1)", "--row-current": "rgba(13, 116, 206, 0.08)",
      "--row-match": "rgba(227, 163, 54, 0.16)", "--tag-text": "#60646c", "--time-text": "#80838d",
      "--message-text": "#1c2024", "--search-hit-bg": "rgba(227, 163, 54, 0.22)", "--search-hit-text": "#4f3422",
      "--chip-outline": "rgba(17, 50, 100, 0.08)", "--chip-info-bg": "#e6f4fe", "--chip-info-text": "#0d74ce",
      "--chip-warn-bg": "#fff7c2", "--chip-warn-text": "#ab6400", "--chip-error-bg": "#feebec", "--chip-error-text": "#ce2c31",
      "--url-underline": "rgba(13, 116, 206, 0.28)", "--json-key": "#0d74ce", "--json-string": "#218358",
      "--json-number": "#ab6400", "--json-boolean": "#107d98", "--json-null": "#8b8d98", "--token-url": "#0d74ce",
      "--token-error": "#ce2c31", "--token-metric": "#ab6400", "--token-method": "#218358", "--token-path": "#1c2024",
      "--token-stack": "#60646c", "--dialog-card-bg": "#f9f9fb", "--dialog-section-bg": "#f0f0f3",
      "--dialog-body-bg": "#fcfcfd", "--dialog-sidebar-bg": "#f9f9fb", "--dialog-sidebar-border": "#d9d9e0",
      "--dialog-nav-bg": "#ffffff", "--dialog-nav-border": "#d9d9e0", "--dialog-nav-hover-bg": "#f4faff",
      "--dialog-nav-hover-border": "#8ec8f6", "--dialog-nav-text": "#1c2024", "--dialog-nav-meta": "#60646c",
      "--dialog-badge-bg": "rgba(13, 116, 206, 0.12)", "--dialog-badge-text": "#113264", "--dialog-main-bg": "#fcfcfd",
      "--dialog-main-header-bg": "#f0f0f3", "--dialog-main-header-border": "#d9d9e0", "--dialog-empty-text": "#80838d",
      "--dialog-code-bg": "#f4faff", "--dialog-rule-bg": "#f9f9fb", "--dialog-rule-title": "#1c2024",
      "--dialog-rule-join-bg": "#f0f0f3", "--dialog-rule-join-border": "#d9d9e0", "--dialog-rule-join-text": "#60646c",
      "--dialog-danger-bg": "rgba(206, 44, 49, 0.08)", "--dialog-danger-border": "rgba(206, 44, 49, 0.24)",
      "--dialog-danger-text": "#ce2c31", "--switch-off": "#b9bbc6", "--switch-on": "#0d74ce", "--switch-thumb": "#ffffff",
      "--status-bg": "#f0f0f3", "--status-border": "#d9d9e0", "--detail-resize-hover": "rgba(13, 116, 206, 0.14)",
      "--empty-circle-fg": "#8b8d98",
    },
  },
  "solarized-light": {
    colorScheme: "light",
    vars: {
      "--bg": "#fdf6e3", "--bg-elevated": "#f5efdc", "--bg-muted": "#eee8d5", "--surface-card": "#f7f0dd",
      "--surface-card-muted": "#eee8d5", "--surface-card-strong": "#e7dfcb", "--surface-sidebar": "#f2ead7",
      "--surface-sidebar-alt": "#ece4d1", "--surface-menu": "#fffaf0", "--surface-input": "#fffdf7",
      "--surface-input-strong": "#f8f2e2", "--surface-hover": "rgba(101, 123, 131, 0.08)", "--surface-hover-strong": "#ece4d1",
      "--surface-hover-accent": "rgba(38, 139, 210, 0.12)", "--surface-overlay": "rgba(88, 110, 117, 0.24)",
      "--border": "#d8d0ba", "--border-strong": "#cabfa7", "--border-hover": "#93a1a1", "--border-subtle": "#e7dfcb",
      "--text": "#586e75", "--text-strong": "#4c626a", "--text-secondary": "#657b83", "--text-muted": "#93a1a1",
      "--text-dim": "#a9b6b6", "--placeholder": "#93a1a1", "--accent": "#2176b5", "--accent-hover": "#268bd2",
      "--accent-fg": "#268bd2", "--accent-soft": "rgba(38, 139, 210, 0.12)", "--accent-contrast": "#fdf6e3",
      "--success": "#859900", "--info": "#2aa198", "--warn": "#b58900", "--error": "#dc322f",
      "--shadow-dialog": "0 24px 60px rgba(101, 123, 131, 0.18)", "--shadow-menu": "0 14px 30px rgba(101, 123, 131, 0.14)",
      "--row": "#eee8d5", "--row-info-bg": "#fdf6e3", "--row-head-border": "#e2dbc8", "--row-head-text": "#657b83",
      "--row-border": "#ece4d1", "--row-hover": "#f5efdc", "--row-marker": "#6c71c4", "--row-current-marker": "#268bd2",
      "--row-match-marker": "#b58900", "--row-selected": "rgba(108, 113, 196, 0.12)", "--row-current": "rgba(38, 139, 210, 0.1)",
      "--row-match": "rgba(181, 137, 0, 0.14)", "--tag-text": "#657b83", "--time-text": "#93a1a1",
      "--message-text": "#586e75", "--search-hit-bg": "rgba(181, 137, 0, 0.18)", "--search-hit-text": "#7b6000",
      "--chip-outline": "rgba(101, 123, 131, 0.1)", "--chip-info-bg": "#e8f2f3", "--chip-info-text": "#2176b5",
      "--chip-warn-bg": "#f8f0cd", "--chip-warn-text": "#9b7600", "--chip-error-bg": "#fae3dd", "--chip-error-text": "#c5302d",
      "--url-underline": "rgba(38, 139, 210, 0.24)", "--json-key": "#268bd2", "--json-string": "#859900",
      "--json-number": "#b58900", "--json-boolean": "#2aa198", "--json-null": "#93a1a1", "--token-url": "#268bd2",
      "--token-error": "#dc322f", "--token-metric": "#b58900", "--token-method": "#2aa198", "--token-path": "#586e75",
      "--token-stack": "#657b83", "--dialog-card-bg": "#f7f0dd", "--dialog-section-bg": "#eee8d5",
      "--dialog-body-bg": "#fdf6e3", "--dialog-sidebar-bg": "#f2ead7", "--dialog-sidebar-border": "#d8d0ba",
      "--dialog-nav-bg": "#fffaf0", "--dialog-nav-border": "#d8d0ba", "--dialog-nav-hover-bg": "#f5efdc",
      "--dialog-nav-hover-border": "#93a1a1", "--dialog-nav-text": "#586e75", "--dialog-nav-meta": "#657b83",
      "--dialog-badge-bg": "rgba(108, 113, 196, 0.14)", "--dialog-badge-text": "#5c61ad", "--dialog-main-bg": "#fdf6e3",
      "--dialog-main-header-bg": "#eee8d5", "--dialog-main-header-border": "#d8d0ba", "--dialog-empty-text": "#93a1a1",
      "--dialog-code-bg": "#f8f2e2", "--dialog-rule-bg": "#f8f2e2", "--dialog-rule-title": "#586e75",
      "--dialog-rule-join-bg": "#eee8d5", "--dialog-rule-join-border": "#d8d0ba", "--dialog-rule-join-text": "#657b83",
      "--dialog-danger-bg": "rgba(220, 50, 47, 0.08)", "--dialog-danger-border": "rgba(220, 50, 47, 0.22)",
      "--dialog-danger-text": "#c5302d", "--switch-off": "#c2b9a3", "--switch-on": "#2176b5", "--switch-thumb": "#fffdf7",
      "--status-bg": "#eee8d5", "--status-border": "#d8d0ba", "--detail-resize-hover": "rgba(38, 139, 210, 0.12)",
      "--empty-circle-fg": "#93a1a1",
    },
  },
};

export function isThemeName(value: string): value is ThemeName {
  return value === "dark" || value === "light" || value === "solarized-light";
}

export function buildThemeStyle(theme: ThemeName): CSSProperties {
  return {
    backgroundColor: palettes[theme].vars["--bg"],
    colorScheme: palettes[theme].colorScheme,
    color: palettes[theme].vars["--text"],
    ...(palettes[theme].vars as unknown as CSSProperties),
  };
}
