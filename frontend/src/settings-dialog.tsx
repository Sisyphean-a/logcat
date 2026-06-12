import { useEffect } from "react";
import { SettingsPreview, ThemePicker } from "./theme-preview";
import { type ThemeName } from "./themes";
import { type ViewSettingKey, type ViewSettings, viewSettingFields } from "./view-settings";

type SettingsDialogProps = {
  open: boolean;
  settings: ViewSettings;
  onChange: (key: ViewSettingKey, value: number) => void;
  onThemeChange: (theme: ThemeName) => void;
  onClose: () => void;
  onReset: () => void;
};

export function SettingsDialog({
  open,
  settings,
  onChange,
  onThemeChange,
  onClose,
  onReset,
}: SettingsDialogProps) {
  useEffect(() => {
    if (!open) {
      return;
    }

    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }

    window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [onClose, open]);

  if (!open) {
    return null;
  }

  return (
    <div
      className="dialog-overlay"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <section className="dialog-card settings-dialog-card">
        <header className="dialog-header">
          <div>
            <div className="dialog-title">界面设置</div>
            <div className="dialog-subtitle">调整后立即生效，并保存在当前本地配置里。</div>
          </div>
          <button className="ghost-button dialog-close" type="button" onClick={onClose}>
            关闭
          </button>
        </header>

        <div className="dialog-body settings-layout">
          <section className="settings-section">
            <div className="settings-section-header">
              <div className="settings-section-title">主题配色</div>
              <div className="settings-section-hint">立即切换黑夜、白天和 Solarized Light，并查看实时示例。</div>
            </div>
            <ThemePicker theme={settings.theme} onChange={onThemeChange} />
          </section>

          <section className="settings-section">
            <div className="settings-section-header">
              <div className="settings-section-title">示例效果</div>
              <div className="settings-section-hint">下面的预览会同步当前主题和字号。</div>
            </div>
            <SettingsPreview settings={settings} />
          </section>

          <section className="settings-section">
            <div className="settings-section-header">
              <div className="settings-section-title">字号微调</div>
              <div className="settings-section-hint">调整后会立即作用到工具栏、表格、详情和状态栏。</div>
            </div>
            <SettingsFields settings={settings} onChange={onChange} />
          </section>
        </div>

        <footer className="dialog-footer settings-actions">
          <button className="text-button secondary" type="button" onClick={onReset}>
            恢复默认
          </button>
          <button className="text-button primary" type="button" onClick={onClose}>
            完成
          </button>
        </footer>
      </section>
    </div>
  );
}

function SettingsFields({
  settings,
  onChange,
}: {
  settings: ViewSettings;
  onChange: (key: ViewSettingKey, value: number) => void;
}) {
  return (
    <div className="settings-grid">
      {viewSettingFields.map((item) => (
        <label key={item.key} className="settings-field">
          <div className="settings-field-header">
            <span>{item.label}</span>
            <span className="settings-value">{settings[item.key]} px</span>
          </div>
          <input
            className="settings-slider"
            max={16}
            min={10}
            onChange={(event) => onChange(item.key, Number(event.target.value))}
            step={1}
            type="range"
            value={settings[item.key]}
          />
          <span className="settings-hint">{item.description}</span>
        </label>
      ))}
    </div>
  );
}
