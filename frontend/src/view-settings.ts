import { useEffect, useMemo, useState, type CSSProperties } from "react";
import { buildThemeStyle, isThemeName, type ThemeName } from "./themes";

const minFontSize = 10;
const maxFontSize = 16;
const storageKey = "logcat:view-settings";

type FontSettings = {
  toolbarFontSize: number;
  filterFontSize: number;
  tableFontSize: number;
  detailFontSize: number;
  statusFontSize: number;
};

export type ViewSettings = FontSettings & {
  theme: ThemeName;
};

export type ViewSettingKey = keyof FontSettings;

export const defaultViewSettings: ViewSettings = {
  theme: "dark",
  toolbarFontSize: 11,
  filterFontSize: 11,
  tableFontSize: 11,
  detailFontSize: 11,
  statusFontSize: 10,
};

export const viewSettingFields: Array<{ key: ViewSettingKey; label: string; description: string }> = [
  { key: "toolbarFontSize", label: "顶部工具栏", description: "设备选择、过滤器入口和右上角工具按钮。" },
  { key: "filterFontSize", label: "筛选栏", description: "包名、查询输入和应用按钮。" },
  { key: "tableFontSize", label: "日志表格", description: "日志列表、列头和级别标签。" },
  { key: "detailFontSize", label: "详情面板", description: "右侧详情内容与复制按钮。" },
  { key: "statusFontSize", label: "状态栏", description: "底部运行状态与统计信息。" },
];

export function useViewSettings() {
  const [settings, setSettings] = useState(loadViewSettings);

  useEffect(() => {
    window.localStorage.setItem(storageKey, JSON.stringify(settings));
  }, [settings]);

  const shellStyle = useMemo(
    () => buildViewStyle(settings),
    [settings],
  );

  function updateSetting(key: ViewSettingKey, value: number) {
    setSettings((current) => ({
      ...current,
      [key]: clampFontSize(value, current[key]),
    }));
  }

  return {
    settings,
    shellStyle,
    updateSetting,
    updateTheme: (theme: ThemeName) => setSettings((current) => ({ ...current, theme })),
    resetSettings: () => setSettings(defaultViewSettings),
  };
}

export function buildViewStyle(settings: ViewSettings): CSSProperties {
  return {
    ...buildThemeStyle(settings.theme),
    "--toolbar-font-size": `${settings.toolbarFontSize}px`,
    "--filter-font-size": `${settings.filterFontSize}px`,
    "--table-font-size": `${settings.tableFontSize}px`,
    "--detail-font-size": `${settings.detailFontSize}px`,
    "--status-font-size": `${settings.statusFontSize}px`,
  } as CSSProperties;
}

function loadViewSettings(): ViewSettings {
  const raw = window.localStorage.getItem(storageKey);
  if (!raw) {
    return defaultViewSettings;
  }

  const parsed = JSON.parse(raw) as Partial<ViewSettings>;
  const theme = parsed.theme;
  return {
    theme: typeof theme === "string" && isThemeName(theme) ? theme : defaultViewSettings.theme,
    toolbarFontSize: clampFontSize(parsed.toolbarFontSize, defaultViewSettings.toolbarFontSize),
    filterFontSize: clampFontSize(parsed.filterFontSize, defaultViewSettings.filterFontSize),
    tableFontSize: clampFontSize(parsed.tableFontSize, defaultViewSettings.tableFontSize),
    detailFontSize: clampFontSize(parsed.detailFontSize, defaultViewSettings.detailFontSize),
    statusFontSize: clampFontSize(parsed.statusFontSize, defaultViewSettings.statusFontSize),
  };
}

function clampFontSize(value: number | undefined, fallback: number) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return fallback;
  }
  return Math.min(maxFontSize, Math.max(minFontSize, Math.round(value)));
}
