import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type MouseEvent as ReactMouseEvent,
  type RefObject,
} from "react";
import { LogRow, type LogItemView } from "./log-row";
import { type LogSelectionMode } from "./use-app-controller";

type LogTableProps = {
  fontSize: number;
  loading: boolean;
  logs: LogItemView[];
  searchQuery: string;
  selectedCount: number;
  visibleCount: number;
  scrollTop: number;
  viewportHeight: number;
  tableRef: RefObject<HTMLDivElement>;
  onScroll: () => void;
  onSelectLog: (index: number, mode: LogSelectionMode) => void;
  onCopySelected: () => void;
  onCopyAll: () => void;
  onClearVisible: () => void;
};

type ColumnKey = "time" | "level" | "tag";
type ContextMenuState = {
  x: number;
  y: number;
  hasSelection: boolean;
};

const columnMinWidths: Record<ColumnKey, number> = {
  time: 84,
  level: 40,
  tag: 92,
};

const defaultColumnWidths: Record<ColumnKey, number> = {
  time: 96,
  level: 44,
  tag: 136,
};

function clampWindowStart(start: number, size: number, visibleRows: number) {
  const maxStart = Math.max(0, size - visibleRows);
  return Math.min(start, maxStart);
}

function resolveRowHeight(fontSize: number) {
  return Math.max(26, fontSize * 2 + 4);
}

export function LogTable({
  fontSize,
  loading,
  logs,
  searchQuery,
  selectedCount,
  visibleCount,
  scrollTop,
  viewportHeight,
  tableRef,
  onScroll,
  onSelectLog,
  onCopySelected,
  onCopyAll,
  onClearVisible,
}: LogTableProps) {
  const [columnWidths, setColumnWidths] = useState(defaultColumnWidths);
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
  const onSelectLogRef = useRef(onSelectLog);
  const contextMenuRef = useRef<HTMLDivElement | null>(null);
  onSelectLogRef.current = onSelectLog;
  const handleSelect = useCallback((index: number, mode: LogSelectionMode) => {
    onSelectLogRef.current(index, mode);
  }, []);
  const rowHeight = resolveRowHeight(fontSize);
  const chipBox = Math.max(18, fontSize + 8);
  const chipSize = Math.max(10, fontSize - 1);

  const gridTemplate = useMemo(
    () => `${columnWidths.time}px ${columnWidths.level}px ${columnWidths.tag}px minmax(260px, 1fr)`,
    [columnWidths],
  );

  const shellStyle = useMemo(
    () =>
      ({
        "--table-columns": gridTemplate,
        "--table-font-size": `${fontSize}px`,
        "--table-row-height": `${rowHeight}px`,
        "--table-head-height": `${rowHeight + 3}px`,
        "--table-chip-box": `${chipBox}px`,
        "--table-chip-size": `${chipSize}px`,
      }) as CSSProperties,
    [chipBox, chipSize, fontSize, gridTemplate, rowHeight],
  );

  const buffer = 20;
  const visibleRows = Math.ceil(viewportHeight / rowHeight) + buffer * 2;
  const start = clampWindowStart(
    Math.max(0, Math.floor(scrollTop / rowHeight) - buffer),
    logs.length,
    visibleRows,
  );
  const end = Math.min(logs.length, start + visibleRows);
  const topSpacer = start * rowHeight;
  const bottomSpacer = Math.max(0, (logs.length - end) * rowHeight);

  useEffect(() => {
    if (!contextMenu) {
      return;
    }

    function handlePointerDown(event: Event) {
      const target = event.target;
      if (target instanceof Node && contextMenuRef.current?.contains(target)) {
        return;
      }
      setContextMenu(null);
    }

    function handleEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setContextMenu(null);
      }
    }

    window.addEventListener("mousedown", handlePointerDown);
    window.addEventListener("scroll", handlePointerDown, true);
    window.addEventListener("keydown", handleEscape);
    return () => {
      window.removeEventListener("mousedown", handlePointerDown);
      window.removeEventListener("scroll", handlePointerDown, true);
      window.removeEventListener("keydown", handleEscape);
    };
  }, [contextMenu]);

  function startResize(key: ColumnKey, event: ReactMouseEvent<HTMLButtonElement>) {
    event.preventDefault();

    const startX = event.clientX;
    const startWidth = columnWidths[key];
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    function handleMove(moveEvent: globalThis.MouseEvent) {
      const nextWidth = startWidth + moveEvent.clientX - startX;
      setColumnWidths((current) => ({
        ...current,
        [key]: Math.max(columnMinWidths[key], nextWidth),
      }));
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

  function openContextMenu(event: ReactMouseEvent<HTMLButtonElement>, index: number) {
    event.preventDefault();
    event.stopPropagation();
    const row = logs[index];
    if (!row) {
      return;
    }
    if (!row.isSelected) {
      handleSelect(index, "replace");
    }
    setContextMenu({
      x: event.clientX,
      y: event.clientY,
      hasSelection: (row.isSelected ? selectedCount : 1) > 0,
    });
  }

  function closeContextMenu() {
    setContextMenu(null);
  }

  function openBodyContextMenu(event: ReactMouseEvent<HTMLDivElement>) {
    event.preventDefault();
    setContextMenu({
      x: event.clientX,
      y: event.clientY,
      hasSelection: selectedCount > 0,
    });
  }

  return (
    <div className="table-shell" style={shellStyle}>
      <div className="table-head">
        <TableHeadCell label="时间" onResize={(event) => startResize("time", event)} />
        <TableHeadCell label="级" onResize={(event) => startResize("level", event)} />
        <TableHeadCell label="标签" onResize={(event) => startResize("tag", event)} />
        <span className="table-head-cell table-head-cell-fill">消息</span>
      </div>

      <div className="table-body" ref={tableRef} onScroll={onScroll} onContextMenu={openBodyContextMenu}>
        {loading ? (
          <div className="placeholder">正在加载状态…</div>
        ) : visibleCount === 0 ? (
          <div className="placeholder">暂无日志</div>
        ) : (
          <div style={{ paddingTop: `${topSpacer}px`, paddingBottom: `${bottomSpacer}px` }}>
            {logs.slice(start, end).map((log) => (
              <LogRow
                key={`${log.index}-${log.raw}`}
                log={log}
                index={log.index}
                onSelect={handleSelect}
                onContextMenu={openContextMenu}
                searchQuery={searchQuery}
              />
            ))}
          </div>
        )}
      </div>

      {contextMenu ? (
        <div
          className="log-context-menu"
          ref={contextMenuRef}
          style={{ left: contextMenu.x, top: contextMenu.y }}
          role="menu"
        >
          {contextMenu.hasSelection ? (
            <button className="log-context-item" type="button" onClick={() => { closeContextMenu(); onCopySelected(); }}>
              <span>复制所选</span>
              <span className="log-context-shortcut">Ctrl+C</span>
            </button>
          ) : null}
          <button className="log-context-item" type="button" onClick={() => { closeContextMenu(); onCopyAll(); }}>
            <span>复制全部</span>
            <span className="log-context-shortcut">Ctrl+Shift+C</span>
          </button>
          <button className="log-context-item danger" type="button" onClick={() => { closeContextMenu(); onClearVisible(); }}>
            <span>清空</span>
            <span className="log-context-shortcut">Ctrl+L</span>
          </button>
        </div>
      ) : null}
    </div>
  );
}

function TableHeadCell({
  label,
  onResize,
}: {
  label: string;
  onResize: (event: ReactMouseEvent<HTMLButtonElement>) => void;
}) {
  return (
    <span className="table-head-cell">
      <span>{label}</span>
      <button
        className="table-resize-handle"
        type="button"
        onMouseDown={onResize}
        aria-label={`调整${label}列宽`}
      />
    </span>
  );
}
