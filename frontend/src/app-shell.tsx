import { useEffect, useState, type MouseEvent as ReactMouseEvent } from "react";
import { main } from "../wailsjs/go/models";
import {
  DetailCollapseIcon,
  DotIcon,
  SaveIcon,
  SearchIcon,
} from "./icons";
import { LogDetailView } from "./log-detail";
import { timeOnly } from "./log-text";
import { SelectControl } from "./select-control";

export type AppState = main.AppState;

type FilterBarProps = {
  state: AppState;
  autoFollow: boolean;
  onSelectPackage: (packageName: string) => void;
  onApplyFilter: (query: string) => void;
  onToggleFollow: () => void;
  onSaveFilter: (query: string) => void;
};

type DetailPanelProps = {
  state: AppState;
  collapsed: boolean;
  onToggle: () => void;
  onCopyDisplay: () => void;
  onCopyRaw: () => void;
  onCopyMessage: () => void;
};

type StatusBarProps = {
  state: AppState;
  autoFollow: boolean;
  statusText: string;
};

export function FilterBar({
  state,
  autoFollow,
  onSelectPackage,
  onApplyFilter,
  onToggleFollow,
  onSaveFilter,
}: FilterBarProps) {
  const packageOptions = state.packages.map((pkg) => ({
    value: pkg.name,
    label: pkg.name,
  }));

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
      <FilterDraftInput
        value={state.filter.draft}
        onApply={onApplyFilter}
        onSave={onSaveFilter}
      />
      <div className="filter-follow">
        <button className={`switch ${autoFollow ? "switch-on" : ""}`} onClick={onToggleFollow}>
          <span className="switch-thumb" />
        </button>
        <span className="switch-label">滚动</span>
      </div>
    </div>
  );
}

type FilterDraftInputProps = {
  value: string;
  onApply: (query: string) => void;
  onSave: (query: string) => void;
};

function FilterDraftInput({ value, onApply, onSave }: FilterDraftInputProps) {
  const [draft, setDraft] = useState(value);
  const [isComposing, setIsComposing] = useState(false);

  useEffect(() => {
    setDraft(value);
  }, [value]);

  function applyDraft() {
    onApply(draft);
  }

  return (
    <>
      <div className="filter-input">
        <span className="filter-icon"><SearchIcon /></span>
        <input
          id="filter-input"
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onCompositionStart={() => setIsComposing(true)}
          onCompositionEnd={() => setIsComposing(false)}
          onKeyDown={(event) => {
            if (event.key !== "Enter" || isComposing || event.nativeEvent.isComposing) {
              return;
            }
            applyDraft();
          }}
          placeholder='tag=jsbridge || message~:"jsbridge"'
        />
      </div>
      <div className="filter-actions">
        <button className="text-button secondary" onClick={applyDraft}>
          应用
        </button>
        <button className="text-button primary" onClick={() => onSave(draft)}>
          <span className="button-icon"><SaveIcon /></span>
          保存
        </button>
      </div>
    </>
  );
}

export function DetailPanel({
  state,
  collapsed,
  onToggle,
  onCopyDisplay,
  onCopyRaw,
  onCopyMessage,
}: DetailPanelProps) {
  const [panelWidth, setPanelWidth] = useState(320);

  function startResize(event: ReactMouseEvent<HTMLButtonElement>) {
    event.preventDefault();

    const startX = event.clientX;
    const startWidth = panelWidth;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    function handleMove(moveEvent: globalThis.MouseEvent) {
      const nextWidth = startWidth + startX - moveEvent.clientX;
      setPanelWidth(Math.max(280, Math.min(760, nextWidth)));
    }

    function handleUp() {
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      window.removeEventListener("mousemove", handleMove);
      window.removeEventListener("mouseup", handleUp);
    }

    window.addEventListener("mousemove", handleMove);
    window.addEventListener("mouseup", handleUp);
  }

  return (
    <aside
      className={`detail-panel ${collapsed ? "collapsed" : ""}`}
      style={collapsed ? undefined : { minWidth: `${panelWidth}px`, width: `${panelWidth}px` }}
    >
      {!collapsed ? (
        <button
          className="detail-resize-handle"
          type="button"
          onMouseDown={startResize}
          aria-label="调整详情面板宽度"
        />
      ) : null}
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
                <div className="detail-title">消息解析</div>
                <LogDetailView text={state.selectedLog.message} />
              </div>
              <div className="detail-block">
                <div className="detail-title">原始日志</div>
                <pre className="detail-rich-text">{state.selectedLog.raw}</pre>
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

export function StatusBar({ state, autoFollow, statusText }: StatusBarProps) {
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
