import { main } from "../wailsjs/go/models";
import { ClearIcon, DeviceIcon, DownloadIcon, PauseIcon, PlayIcon, SettingsIcon } from "./icons";
import { SelectControl, type SelectOption } from "./select-control";

type AppState = main.AppState;

export type ToolbarProps = {
  canEditSavedFilter: boolean;
  state: AppState;
  onEditSavedFilter: () => void;
  onSelectDevice: (deviceID: string) => void;
  onSetPackageScope: (scope: string) => void;
  onApplySavedFilter: (filterID: string) => void;
  onPauseToggle: () => void;
  onClearVisible: () => void;
  onExport: () => void;
  onOpenSettings: () => void;
};

export function Toolbar({
  canEditSavedFilter,
  state,
  onEditSavedFilter,
  onSelectDevice,
  onSetPackageScope,
  onApplySavedFilter,
  onPauseToggle,
  onClearVisible,
  onExport,
  onOpenSettings,
}: ToolbarProps) {
  return (
    <header className="toolbar">
      <div className="brand">
        <div className="brand-mark">H5</div>
        <div className="brand-title">Logcat Viewer</div>
      </div>
      <div className="toolbar-sep" />
      <SelectControl
        className="toolbar-device"
        emptyLabel="选择设备"
        leading={<DeviceIcon />}
        onChange={onSelectDevice}
        options={buildDeviceOptions(state)}
        value={state.selectedDevice}
      />
      <SelectControl
        className="toolbar-scope"
        clearable={false}
        emptyLabel="全部包"
        onChange={onSetPackageScope}
        options={buildScopeOptions()}
        value={state.packageScope}
      />
      <div className="toolbar-sep" />
      <SavedFilterControls
        activeFilterID={state.filter.activeFilterId}
        canEditSavedFilter={canEditSavedFilter}
        filters={state.filter.saved}
        onApplySavedFilter={onApplySavedFilter}
        onEditSavedFilter={onEditSavedFilter}
      />
      <div className="toolbar-spacer" />
      <ToolbarActions
        paused={state.pause.active}
        onClearVisible={onClearVisible}
        onExport={onExport}
        onOpenSettings={onOpenSettings}
        onPauseToggle={onPauseToggle}
      />
    </header>
  );
}

function buildDeviceOptions(state: AppState): SelectOption[] {
  return state.devices.map((device) => ({
    value: device.id,
    label: device.model || device.id,
  }));
}

function buildScopeOptions(): SelectOption[] {
  return [
    { value: "all", label: "全部包" },
    { value: "user", label: "用户包" },
    { value: "system", label: "系统包" },
  ];
}

function SavedFilterControls({
  activeFilterID,
  canEditSavedFilter,
  filters,
  onApplySavedFilter,
  onEditSavedFilter,
}: {
  activeFilterID: string;
  canEditSavedFilter: boolean;
  filters: AppState["filter"]["saved"];
  onApplySavedFilter: (filterID: string) => void;
  onEditSavedFilter: () => void;
}) {
  const filterOptions: SelectOption[] = filters.map((filter) => ({
    value: filter.id,
    label: filter.name,
    tone: filter.id === activeFilterID ? "accent" : "default",
  }));

  return (
    <div className="toolbar-filter-group">
      <SelectControl
        className="toolbar-filter"
        emptyLabel="未选择过滤器"
        onChange={onApplySavedFilter}
        options={filterOptions}
        value={activeFilterID || ""}
      />
      <button
        className="ghost-button mini-button"
        disabled={!canEditSavedFilter}
        onClick={onEditSavedFilter}
        type="button"
      >
        编辑
      </button>
    </div>
  );
}

function ToolbarActions({
  paused,
  onClearVisible,
  onExport,
  onOpenSettings,
  onPauseToggle,
}: {
  paused: boolean;
  onClearVisible: () => void;
  onExport: () => void;
  onOpenSettings: () => void;
  onPauseToggle: () => void;
}) {
  return (
    <div className="toolbar-actions">
      <button className="icon-button" onClick={onPauseToggle} title={paused ? "恢复" : "暂停"} type="button">
        {paused ? <PlayIcon /> : <PauseIcon />}
      </button>
      <button className="icon-button" onClick={onClearVisible} title="清空视图" type="button">
        <ClearIcon />
      </button>
      <button className="icon-button" onClick={onExport} title="导出" type="button">
        <DownloadIcon />
      </button>
      <div className="toolbar-mini-sep" />
      <button className="icon-button" onClick={onOpenSettings} title="设置" type="button">
        <SettingsIcon />
      </button>
    </div>
  );
}
