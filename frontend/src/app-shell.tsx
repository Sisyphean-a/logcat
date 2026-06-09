import { ClearIcon, DetailCollapseIcon, DeviceIcon, DotIcon, DownloadIcon, PauseIcon, PlayIcon, SaveIcon, SearchIcon, SettingsIcon } from "./icons";
import { main } from "../wailsjs/go/models";
import { SelectControl, type SelectOption } from "./select-control";

export type AppState = main.AppState;
export type LogItemView = main.LogItemView;

export function Toolbar({
  state,
  onSelectDevice,
  onApplySavedFilter,
  onPauseToggle,
  onClearVisible,
  onExport,
}: {
  state: AppState;
  onSelectDevice: (deviceID: string) => void;
  onApplySavedFilter: (filterID: string) => void;
  onPauseToggle: () => void;
  onClearVisible: () => void;
  onExport: () => void;
}) {
  const deviceOptions: SelectOption[] = state.devices.map((device) => ({
    value: device.id,
    label: device.model || device.id,
  }));
  const filterOptions: SelectOption[] = state.filter.saved.map((filter) => ({
    value: filter.id,
    label: filter.name,
    tone: filter.id === state.filter.activeFilterId ? "accent" : "default",
  }));

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
        options={deviceOptions}
        value={state.selectedDevice}
      />
      <div className="toolbar-sep" />
      <SelectControl
        className="toolbar-filter"
        emptyLabel="未选择过滤器"
        onChange={onApplySavedFilter}
        options={filterOptions}
        value={state.filter.activeFilterId || ""}
      />
      <div className="toolbar-spacer" />
      <div className="toolbar-actions">
        <button className="icon-button" onClick={onPauseToggle} title={state.pause.active ? "恢复" : "暂停"}>
          {state.pause.active ? <PlayIcon /> : <PauseIcon />}
        </button>
        <button className="icon-button" onClick={onClearVisible} title="清空视图">
          <ClearIcon />
        </button>
        <button className="icon-button" onClick={onExport} title="导出">
          <DownloadIcon />
        </button>
        <div className="toolbar-mini-sep" />
        <button className="icon-button" title="设置">
          <SettingsIcon />
        </button>
      </div>
    </header>
  );
}

export function FilterBar({
  state,
  autoFollow,
  onSelectPackage,
  onFilterDraftChange,
  onApplyFilter,
  onSetPackageScope,
  onToggleFollow,
  onSaveFilter,
}: {
  state: AppState;
  autoFollow: boolean;
  onSelectPackage: (packageName: string) => void;
  onFilterDraftChange: (query: string) => void;
  onApplyFilter: () => void;
  onSetPackageScope: (scope: string) => void;
  onToggleFollow: () => void;
  onSaveFilter: () => void;
}) {
  const packageOptions: SelectOption[] = state.packages.map((pkg) => ({
    value: pkg.name,
    label: pkg.name,
  }));
  const scopeOptions: SelectOption[] = [
    { value: "all", label: "全部包" },
    { value: "user", label: "用户包" },
    { value: "system", label: "系统包" },
  ];

  return (
    <div className="filter-bar">
      <SelectControl
        className="package-select"
        emptyLabel="全部包名"
        filterable
        onChange={onSelectPackage}
        options={packageOptions}
        value={state.selectedPackage}
      />
      <div className="filter-input">
        <span className="filter-icon"><SearchIcon /></span>
        <input
          id="filter-input"
          value={state.filter.draft}
          onChange={(event) => onFilterDraftChange(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              onApplyFilter();
            }
          }}
          placeholder="tag:chromium & message:[H5]"
        />
      </div>
      <SelectControl
        className="scope-select"
        emptyLabel="全部包"
        onChange={onSetPackageScope}
        options={scopeOptions}
        value={state.packageScope}
      />
      <div className="filter-follow">
        <button className={`switch ${autoFollow ? "switch-on" : ""}`} onClick={onToggleFollow}>
          <span className="switch-thumb" />
        </button>
        <span className="switch-label">滚动</span>
      </div>
      <div className="filter-actions">
        <button className="text-button secondary" onClick={onApplyFilter}>
          应用
        </button>
        <button className="text-button primary" onClick={onSaveFilter}>
          <span className="button-icon"><SaveIcon /></span>
          保存
        </button>
      </div>
    </div>
  );
}

export function LogTable({
  loading,
  logs,
  visibleCount,
  scrollTop,
  viewportHeight,
  tableRef,
  onScroll,
  onSelectLog,
}: {
  loading: boolean;
  logs: LogItemView[];
  visibleCount: number;
  scrollTop: number;
  viewportHeight: number;
  tableRef: React.RefObject<HTMLDivElement>;
  onScroll: () => void;
  onSelectLog: (index: number) => void;
}) {
  const rowHeight = 27;
  const buffer = 20;
  const start = Math.max(0, Math.floor(scrollTop / rowHeight) - buffer);
  const visibleRows = Math.ceil(viewportHeight / rowHeight) + buffer * 2;
  const end = Math.min(logs.length, start + visibleRows);
  const topSpacer = start * rowHeight;
  const bottomSpacer = Math.max(0, (logs.length - end) * rowHeight);

  return (
    <div className="table-shell">
      <div className="table-head">
        <span>时间</span>
        <span>级</span>
        <span>标签</span>
        <span>消息</span>
        <span>来源</span>
      </div>
      <div className="table-body" ref={tableRef} onScroll={onScroll}>
        {loading ? (
          <div className="placeholder">正在加载状态…</div>
        ) : visibleCount === 0 ? (
          <div className="placeholder">暂无日志</div>
        ) : (
          <div style={{ paddingTop: `${topSpacer}px`, paddingBottom: `${bottomSpacer}px` }}>
            {logs.slice(start, end).map((log) => (
              <LogRow key={`${log.index}-${log.raw}`} log={log} onClick={() => onSelectLog(log.index)} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export function DetailPanel({
  state,
  collapsed,
  onToggle,
  onCopyDisplay,
  onCopyRaw,
  onCopyMessage,
}: {
  state: AppState;
  collapsed: boolean;
  onToggle: () => void;
  onCopyDisplay: () => void;
  onCopyRaw: () => void;
  onCopyMessage: () => void;
}) {
  return (
    <aside className={`detail-panel ${collapsed ? "collapsed" : ""}`}>
      <button className="detail-toggle" onClick={onToggle}>
        <DetailCollapseIcon />
      </button>
      {!collapsed && (
        <div className="detail-body">
          <div className="detail-header">详情面板</div>
          {state.selectedLog ? (
            <div className="detail-content">
              <div className="detail-actions">
                <button className="ghost-button" onClick={onCopyDisplay}>复制行</button>
                <button className="ghost-button" onClick={onCopyRaw}>复制原文</button>
                <button className="ghost-button" onClick={onCopyMessage}>复制消息</button>
              </div>
              <dl className="detail-grid">
                <div><dt>时间</dt><dd>{timeOnly(state.selectedLog.timeText)}</dd></div>
                <div><dt>级别</dt><dd>{state.selectedLog.level}</dd></div>
                <div><dt>标签</dt><dd>{state.selectedLog.tag}</dd></div>
                <div><dt>来源</dt><dd>{state.selectedLog.source || "-"}</dd></div>
              </dl>
              <div className="detail-block">
                <div className="detail-title">消息</div>
                <pre>{state.selectedLog.message}</pre>
              </div>
              <div className="detail-block">
                <div className="detail-title">原始日志</div>
                <pre>{state.selectedLog.raw}</pre>
              </div>
            </div>
          ) : (
            <div className="detail-empty">
              <div className="empty-circle">↑</div>
              <div>选择一条日志</div>
              <div>查看详情</div>
            </div>
          )}
        </div>
      )}
    </aside>
  );
}

export function StatusBar({
  state,
  autoFollow,
  statusText,
}: {
  state: AppState;
  autoFollow: boolean;
  statusText: string;
}) {
  const currentDevice = state.devices.find((item) => item.id === state.selectedDevice);
  const currentFilter = state.filter.saved.find((item) => item.id === state.filter.activeFilterId);

  return (
    <footer className="status-bar">
      <div className="status-item ok"><span className="status-dot ok"><DotIcon /></span>adb {state.adbStatus}</div>
      <div className="status-sep" />
      <div className="status-item">设备 {currentDevice?.model || "-"}</div>
      <div className="status-sep" />
      <div className="status-item accent">包名 {state.selectedPackage || "-"}</div>
      <div className="status-sep" />
      <div className="status-item">
        日志 {state.totalLogs} / <span className="accent-number">{state.visibleCount}</span> 条
      </div>
      <div className="status-sep" />
      <div className="status-item accent">过滤器 {currentFilter?.name || "自定义"}</div>
      <div className="status-fill" />
      <div className="status-item">自动滚动 {autoFollow ? "开" : "关"}</div>
      {statusText && (
        <>
          <div className="status-sep" />
          <div className="status-item">{statusText}</div>
        </>
      )}
    </footer>
  );
}

function LogRow({ log, onClick }: { log: LogItemView; onClick: () => void }) {
  const tone = log.level === "E" || log.level === "F" ? "error" : log.level === "W" ? "warn" : "info";
  return (
    <button
      className={[
        "table-row",
        `tone-${tone}`,
        log.isSelected ? "selected" : "",
        log.isCurrent ? "current" : "",
      ].join(" ")}
      onClick={onClick}
    >
      <span>{timeOnly(log.timeText)}</span>
      <span className={`level-chip ${tone}`}>{log.level}</span>
      <span className="tag-cell">{log.tag}</span>
      <span className="message-cell">{log.message}</span>
      <span className="source-cell">{log.source || "-"}</span>
    </button>
  );
}

function timeOnly(value: string) {
  const parts = value.split(" ");
  return parts.length > 1 ? parts[1] : value;
}
