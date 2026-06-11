import { useEffect } from "react";
import { type ViewSettingKey, type ViewSettings, viewSettingFields } from "./view-settings";

type SettingsDialogProps = {
  open: boolean;
  settings: ViewSettings;
  onChange: (key: ViewSettingKey, value: number) => void;
  onClose: () => void;
  onReset: () => void;
};

export function SettingsDialog({ open, settings, onChange, onClose, onReset }: SettingsDialogProps) {
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

        <SettingsFields settings={settings} onChange={onChange} />

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
    <div className="dialog-body settings-grid">
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
