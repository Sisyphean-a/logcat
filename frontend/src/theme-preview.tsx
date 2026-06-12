import { buildViewStyle, type ViewSettings } from "./view-settings";
import { buildThemeStyle, themeOptions, type ThemeName } from "./themes";

const previewRows = [
  { time: "12:41:09.118", level: "I", tag: "chromium", message: "GET /api/logs 200 18ms", tone: "info" },
  { time: "12:41:10.204", level: "W", tag: "WebView", message: "render timeout after 420ms", tone: "warn" },
  { time: "12:41:11.032", level: "E", tag: "H5Bridge", message: "socket closed unexpectedly", tone: "error" },
] as const;

export function ThemePicker({
  theme,
  onChange,
}: {
  theme: ThemeName;
  onChange: (theme: ThemeName) => void;
}) {
  return (
    <div className="theme-choice-grid">
      {themeOptions.map((option) => (
        <ThemeCard
          key={option.value}
          option={option}
          selected={option.value === theme}
          onSelect={() => onChange(option.value)}
        />
      ))}
    </div>
  );
}

export function SettingsPreview({ settings }: { settings: ViewSettings }) {
  return (
    <section className="settings-preview theme-scope" style={buildViewStyle(settings)}>
      <div className="settings-preview-toolbar">
        <div className="settings-preview-brand">
          <span className="settings-preview-brand-mark">L</span>
          <span className="settings-preview-brand-title">logcat</span>
        </div>
        <span className="settings-preview-pill">Pixel 8 Pro</span>
        <span className="settings-preview-spacer" />
        <span className="settings-preview-action">筛选</span>
        <span className="settings-preview-action accent">设置</span>
      </div>

      <div className="settings-preview-filter">
        <span className="settings-preview-query">package:com.demo query:error|warn</span>
        <span className="settings-preview-toggle" aria-hidden="true">
          <span className="settings-preview-toggle-thumb" />
        </span>
      </div>

      <div className="settings-preview-table">
        <div className="settings-preview-head">
          <span>时间</span>
          <span>级</span>
          <span>标签</span>
          <span>消息</span>
        </div>
        {previewRows.map((row, index) => <PreviewRow key={row.time} row={row} emphasis={index === 1} />)}
      </div>

      <div className="settings-preview-detail">
        <div className="settings-preview-detail-title">详情</div>
        <div className="settings-preview-detail-grid">
          <span>设备</span>
          <strong>Pixel 8 Pro / Android 15</strong>
          <span>上下文</span>
          <strong>chromium · H5Bridge</strong>
        </div>
      </div>

      <div className="settings-preview-status">
        <span>running</span>
        <span>2,481 visible</span>
        <span>match: warn|error</span>
      </div>
    </section>
  );
}

function ThemeCard({
  option,
  selected,
  onSelect,
}: {
  option: (typeof themeOptions)[number];
  selected: boolean;
  onSelect: () => void;
}) {
  return (
    <button
      className={`theme-choice-card theme-scope ${selected ? "selected" : ""}`}
      type="button"
      onClick={onSelect}
      style={buildThemeStyle(option.value)}
      aria-pressed={selected}
    >
      <div className="theme-choice-header">
        <div>
          <div className="theme-choice-label">{option.label}</div>
          <div className="theme-choice-caption">{option.description}</div>
        </div>
        <span className="theme-choice-accent" />
      </div>
      <div className="theme-choice-swatches">
        <span className="theme-choice-swatch bg" />
        <span className="theme-choice-swatch panel" />
        <span className="theme-choice-swatch accent" />
      </div>
      <div className="theme-choice-mini">
        <div className="theme-choice-mini-top" />
        <div className="theme-choice-mini-row active">
          <span className="theme-choice-mini-chip" />
          <span className="theme-choice-mini-line" />
          <span className="theme-choice-mini-dot" />
        </div>
        <div className="theme-choice-mini-row">
          <span className="theme-choice-mini-chip warn" />
          <span className="theme-choice-mini-line short" />
          <span className="theme-choice-mini-dot muted" />
        </div>
      </div>
    </button>
  );
}

function PreviewRow({
  row,
  emphasis,
}: {
  row: (typeof previewRows)[number];
  emphasis: boolean;
}) {
  return (
    <div className={`settings-preview-row ${emphasis ? "emphasis" : ""}`}>
      <span className="settings-preview-cell time">{row.time}</span>
      <span className={`settings-preview-chip ${row.tone}`}>{row.level}</span>
      <span className="settings-preview-cell tag">{row.tag}</span>
      <span className="settings-preview-cell message">{row.message}</span>
    </div>
  );
}
