import { startTransition, useEffect, useMemo, useRef, useState, type MutableRefObject } from "react";
import { EventsOff, EventsOn } from "../wailsjs/runtime/runtime";
import { app, main } from "../wailsjs/go/models";
import { createMockPreviewLogs, createMockState, type PreviewLogRow } from "./mock-state";
import {
  ApplyFilterDraft,
  ApplySavedFilter,
  ClearVisible,
  CopyAllVisibleLogs,
  CopySelectedLogs,
  CopyText,
  ExportVisibleLogs,
  GetState,
  GetSelectedLogRaw,
  Pause,
  ReplaceSavedFilterDefinitions,
  ResumeKeep,
  SaveFilterDefinition,
  SelectDevice,
  SelectLog,
  SelectLogs,
  SelectPackage,
  SetFilterDraft,
  SetPackageScope,
  SetSearchQuery,
  UpdateSavedFilterDefinition,
} from "../wailsjs/go/main/App";
import { type SaveFilterDraft } from "./filter-rule-builder";
import { type SavedFiltersDraft } from "./saved-filter-types";
import {
  applyStateAppendPatch,
  cloneAppStateShell,
  type AppState,
  type StateAppendPatch,
} from "./state-stream";

export type { AppState } from "./state-stream";
export type LogSelectionMode = "replace" | "add" | "range";
export type ResultSearchPreview = {
  query: string;
  highlightTerms: string[];
};
export type SelectedLogDetail = {
  raw: string;
};

const stateEventName = "state:updated";
const stateAppendEventName = "state:append";

type SelectionPatch = main.SelectionPatch;

export function useAppController() {
  const [state, setState] = useState<AppState>(emptyState);
  const [loading, setLoading] = useState(true);
  const [actionError, setActionError] = useState("");
  const [autoFollow, setAutoFollow] = useState(true);
  const [detailCollapsed, setDetailCollapsed] = useState(false);
  const [scrollTop, setScrollTop] = useState(0);
  const [viewportHeight, setViewportHeight] = useState(0);
  const tableRef = useRef<HTMLDivElement | null>(null);
  const autoFollowRef = useRef(autoFollow);
  const ignoreScrollRef = useRef(false);
  const previewAllLogsRef = useRef<PreviewLogRow[]>([]);
  const stateRef = useRef(state);
  const selectedLogRawRef = useRef("");
  const latestRevisionRef = useRef(state.revision);
  const latestSearchRequestRef = useRef(0);
  const pendingSearchQueryRef = useRef<{ id: number; query: string } | null>(null);
  const selectedSourceIndexesRef = useRef<number[]>([]);
  const focusedSourceIndexRef = useRef(-1);
  const selectionTrackingRevisionRef = useRef(state.revision);
  const pendingEventStateRef = useRef<AppState | null>(null);
  const pendingEventFrameRef = useRef<number | null>(null);
  const pendingMetricsNodeRef = useRef<HTMLDivElement | null>(null);
  const pendingMetricsFrameRef = useRef<number | null>(null);
  if (!pendingEventStateRef.current || state.revision >= pendingEventStateRef.current.revision) {
    stateRef.current = state;
    latestRevisionRef.current = state.revision;
  }

  function syncTableMetrics(node: HTMLDivElement) {
    setScrollTop(node.scrollTop);
    setViewportHeight(node.clientHeight);
  }

  function flushPendingMetrics() {
    pendingMetricsFrameRef.current = null;
    const node = pendingMetricsNodeRef.current;
    pendingMetricsNodeRef.current = null;
    if (!node) {
      return;
    }
    syncTableMetrics(node);
  }

  function queueTableMetrics(node: HTMLDivElement) {
    pendingMetricsNodeRef.current = node;
    if (pendingMetricsFrameRef.current !== null) {
      return;
    }
    pendingMetricsFrameRef.current = requestAnimationFrame(flushPendingMetrics);
  }

  function applyNextState(next: AppState, options?: { clearSelectedRaw?: boolean; urgent?: boolean }) {
    const clearSelectedRaw = options?.clearSelectedRaw ?? false;
    const urgent = options?.urgent ?? false;
    if (next.revision < latestRevisionRef.current) {
      return;
    }
    latestRevisionRef.current = next.revision;
    if (clearSelectedRaw) {
      selectedLogRawRef.current = "";
    }
    updateSelectionTracking(next);
    stateRef.current = next;
    const commit = () => setState(next);
    if (urgent) {
      commit();
      return;
    }
    startTransition(commit);
  }

  function cancelPendingEventFrame() {
    if (pendingEventFrameRef.current === null) {
      return;
    }
    cancelAnimationFrame(pendingEventFrameRef.current);
    pendingEventFrameRef.current = null;
  }

  function flushPendingEventState() {
    pendingEventFrameRef.current = null;
    const next = pendingEventStateRef.current;
    pendingEventStateRef.current = null;
    if (!next) {
      return;
    }
    applyNextState(next);
  }

  function schedulePendingEventState() {
    if (pendingEventFrameRef.current !== null) {
      return;
    }
    pendingEventFrameRef.current = requestAnimationFrame(flushPendingEventState);
  }

  function queuePendingEventState(next: AppState) {
    if (next.revision < latestRevisionRef.current) {
      return;
    }
    const pending = pendingEventStateRef.current;
    if (pending && next.revision < pending.revision) {
      return;
    }
    pendingEventStateRef.current = next;
    stateRef.current = next;
    updateSelectionTracking(next);
    latestRevisionRef.current = next.revision;
    schedulePendingEventState();
  }

  function queueEventState(next: AppState) {
    queuePendingEventState(next);
  }

  function setLatestSearchState(query: string) {
    const pending = pendingSearchQueryRef.current;
    if (pending?.query === query) {
      return Promise.resolve();
    }
    if (!pending && stateRef.current.search.query === query) {
      return Promise.resolve();
    }
    const requestID = latestSearchRequestRef.current + 1;
    latestSearchRequestRef.current = requestID;
    pendingSearchQueryRef.current = { id: requestID, query };
    return SetSearchQuery(query).then((next: AppState) => {
      const latest = pendingSearchQueryRef.current;
      if (!latest || latest.id !== requestID) {
        return;
      }
      pendingSearchQueryRef.current = null;
      applyNextState(next);
    }).catch((error) => {
      const latest = pendingSearchQueryRef.current;
      if (latest?.id === requestID) {
        pendingSearchQueryRef.current = null;
      }
      throw error;
    });
  }

  function applyAppendPatch(patch: StateAppendPatch) {
    const pending = pendingEventStateRef.current;
    const base = pending && pending.revision >= stateRef.current.revision
      ? pending
      : stateRef.current;
    if (patch.revision < base.revision) {
      return;
    }
    queuePendingEventState(applyStateAppendPatch(base, patch));
  }

  function applySelectionPatch(patch: SelectionPatch) {
    if (patch.revision < latestRevisionRef.current) {
      return;
    }
    selectedLogRawRef.current = "";
    startTransition(() => {
      setState((current) => {
        if (patch.revision < current.revision) {
          return current;
        }
        const tracking = resolveTrackedSelection(current);
        const nextLogs = applySelectionRows(
          current.logs,
          collectSelectionChangeSources(
            tracking.selectedSourceIndexes,
            patch.selectedSourceIndexes,
            tracking.focusedSourceIndex,
            patch.focusedSourceIndex,
          ),
          patch.selectedSourceIndexes,
          patch.focusedSourceIndex,
        );
        const nextSelectedLog = patch.selectedLog
          ? current.selectedLog && sameSelectedLog(current.selectedLog, patch.selectedLog)
            ? current.selectedLog
            : patch.selectedLog
          : undefined;
        latestRevisionRef.current = patch.revision;
        selectedSourceIndexesRef.current = patch.selectedSourceIndexes;
        focusedSourceIndexRef.current = patch.focusedSourceIndex;
        selectionTrackingRevisionRef.current = patch.revision;
        const next = cloneAppStateShell(current);
        next.revision = patch.revision;
        if (nextLogs === current.logs &&
          current.selectedCount === patch.selectedCount &&
          current.selectedLog === nextSelectedLog) {
          return next;
        }
        next.selectedCount = patch.selectedCount;
        next.selectedLog = nextSelectedLog;
        next.logs = nextLogs;
        return next;
      });
    });
  }

  function updateSelectionTracking(next: AppState) {
    selectedSourceIndexesRef.current = collectSelectedSourceIndexes(next.logs, next.selectedCount);
    focusedSourceIndexRef.current = next.selectedLog?.sourceIndex ?? -1;
    selectionTrackingRevisionRef.current = next.revision;
  }

  function resolveTrackedSelection(current: AppState) {
    if (selectionTrackingRevisionRef.current === current.revision) {
      return {
        selectedSourceIndexes: selectedSourceIndexesRef.current,
        focusedSourceIndex: focusedSourceIndexRef.current,
      };
    }
    return {
      selectedSourceIndexes: collectSelectedSourceIndexes(current.logs, current.selectedCount),
      focusedSourceIndex: current.selectedLog?.sourceIndex ?? -1,
    };
  }

  useEffect(() => {
    autoFollowRef.current = autoFollow;
  }, [autoFollow]);

  useEffect(() => {
    if (!isWailsRuntime()) {
      const snapshot = createMockState();
      previewAllLogsRef.current = createMockPreviewLogs();
      applyNextState(snapshot, { urgent: true });
      setLoading(false);
      return;
    }

    let mounted = true;

    GetState()
      .then((next) => {
        if (!mounted) {
          return;
        }
        applyNextState(next, { clearSelectedRaw: true, urgent: true });
        setLoading(false);
      })
      .catch((error: unknown) => {
        if (!mounted) {
          return;
        }
        setActionError(String(error));
        setLoading(false);
      });

    // 流式热路径：事件 payload 已是反序列化的纯对象，组件只读字段、
    // 不调用类方法，直接 setState 可省去每帧对 ~1000 条日志的深拷贝重建。
    const handler = (next: AppState) => {
      queueEventState(next);
    };
    const appendHandler = (patch: StateAppendPatch) => {
      applyAppendPatch(patch);
    };

    EventsOn(stateEventName, handler);
    EventsOn(stateAppendEventName, appendHandler);
    return () => {
      mounted = false;
      cancelPendingEventFrame();
      pendingEventStateRef.current = null;
      if (pendingMetricsFrameRef.current !== null) {
        cancelAnimationFrame(pendingMetricsFrameRef.current);
        pendingMetricsFrameRef.current = null;
      }
      pendingMetricsNodeRef.current = null;
      EventsOff(stateEventName);
      EventsOff(stateAppendEventName);
    };
  }, []);

  useEffect(() => {
    if (autoFollow) {
      scrollToBottom();
    }
  }, [autoFollow, state.logs.length]);

  useEffect(() => {
    const node = tableRef.current;
    if (!node) {
      return;
    }
    queueTableMetrics(node);
  }, [state.logs.length]);

  const api = useMemo(
    () => (isWailsRuntime() ? {
      selectDevice: (deviceID: string) => withAction(() => SelectDevice(deviceID), setActionError),
      applySavedFilter: (filterID: string) => withAction(() => ApplySavedFilter(filterID), setActionError),
      selectPackage: (packageName: string) => withAction(() => SelectPackage(packageName), setActionError),
      setPackageScope: (scope: string) => withAction(() => SetPackageScope(scope), setActionError),
      getSelectedLogDetail: async (): Promise<SelectedLogDetail | undefined> => {
        const selected = stateRef.current.selectedLog;
        if (!selected) {
          selectedLogRawRef.current = "";
          return undefined;
        }
        const raw = await GetSelectedLogRaw();
        selectedLogRawRef.current = raw;
        return { raw };
      },
      setFilterDraft: (query: string) =>
        SetFilterDraft(query).then((next: AppState) => applyNextState(next, { urgent: true })),
      setSearchQuery: (query: string) => setLatestSearchState(query),
      applyFilter: async (query?: string) => {
        if (query !== undefined) {
          setFilterDraftStateInRef(stateRef, setState, query);
        }
        await withAction(ApplyFilterDraft, setActionError);
      },
      exportVisible: () => withAction(ExportVisibleLogs, setActionError),
      copySelected: async (kind: "display" | "raw" | "message") => {
        const selected = stateRef.current.selectedLog;
        if (!selected) {
          return;
        }
        let value = selected.message;
        if (kind === "display") {
          value = formatPreviewLogDisplay(selected);
        } else if (kind === "raw") {
          value = selectedLogRawRef.current || await GetSelectedLogRaw();
          selectedLogRawRef.current = value;
        }
        await withAction(() => CopyText(value), setActionError);
      },
      copySelectedLogs: () => withAction(CopySelectedLogs, setActionError),
      copyAllVisibleLogs: () => withAction(CopyAllVisibleLogs, setActionError),
      selectLog: (index: number) =>
        SelectLog(index).then((patch: SelectionPatch) => {
          applySelectionPatch(patch);
        }),
      selectLogs: (index: number, mode: LogSelectionMode) =>
        SelectLogs({ index, mode }).then((patch: SelectionPatch) => {
          applySelectionPatch(patch);
        }),
      pauseToggle: async () => {
        const current = stateRef.current;
        const wantsResume = shouldResumeStreaming(current);
        const next = wantsResume ? await ResumeKeep() : await Pause();
        if (wantsResume && !next.sessionActive && hasStartFailureStatus(next.status)) {
          setActionError(next.status);
        } else {
          setActionError("");
        }
        applyNextState(next);
      },
      clearVisible: () =>
        ClearVisible().then((next: AppState) => {
          applyNextState(next, { clearSelectedRaw: true });
        }),
      saveFilter: async (draft: SaveFilterDraft) => {
        await withAction(
          () => SaveFilterDefinition(draft.name, draft.packageName, draft.query),
          setActionError,
        );
      },
      updateFilter: async (filterID: string, draft: SaveFilterDraft) => {
        await withAction(
          () => UpdateSavedFilterDefinition(filterID, draft.name, draft.packageName, draft.query),
          setActionError,
        );
      },
      replaceSavedFilters: async (draft: SavedFiltersDraft) => {
        const payload = draft.filters.map((filter) => new app.SavedFilterDraft({
          ExistingID: filter.existingID,
          Name: filter.name,
          PackageName: filter.packageName,
          Query: filter.query,
        }));
        await withAction(
          () => ReplaceSavedFilterDefinitions(payload, draft.defaultFilterID, draft.activeFilterID),
          setActionError,
        );
      },
    } : createPreviewApi(state, setState, setActionError, previewAllLogsRef)),
    // Wails 模式下回调全部走 stateRef，api 可稳定（依赖恒定空键），
    // 避免流式每帧重建 api 击穿子组件 memo；预览模式无流式，仍依赖 state。
    [isWailsRuntime() ? null : state],
  );

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (isEditableTarget(event.target)) {
        return;
      }
      if (hasActiveTextSelection()) {
        return;
      }
      const command = event.ctrlKey || event.metaKey;
      if (!command) {
        return;
      }
      const key = event.key.toLowerCase();
      if (key === "c" && event.shiftKey) {
        event.preventDefault();
        void api.copyAllVisibleLogs();
        return;
      }
      if (key === "c" && stateRef.current.selectedCount > 0) {
        event.preventDefault();
        void api.copySelectedLogs();
        return;
      }
      if (key === "l") {
        event.preventDefault();
        void api.clearVisible();
      }
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [api]);

  function handleScroll() {
    const node = tableRef.current;
    if (!node) {
      return;
    }

    queueTableMetrics(node);
    if (ignoreScrollRef.current) {
      return;
    }

    const atBottom = node.scrollHeight - node.scrollTop - node.clientHeight < 8;
    if (!atBottom && autoFollowRef.current) {
      setAutoFollow(false);
    }
  }

  function setAutoFollowEnabled(next: boolean) {
    setAutoFollow(next);
    if (next) {
      scrollToBottom();
    }
  }

  function scrollToBottom() {
    const node = tableRef.current;
    if (!node) {
      return;
    }

    ignoreScrollRef.current = true;
    requestAnimationFrame(() => {
      const currentNode = tableRef.current;
      if (!currentNode) {
        ignoreScrollRef.current = false;
        return;
      }
      currentNode.scrollTop = currentNode.scrollHeight;
      queueTableMetrics(currentNode);
      requestAnimationFrame(() => {
        ignoreScrollRef.current = false;
      });
    });
  }

  return {
    state,
    loading,
    actionError,
    autoFollow,
    detailCollapsed,
    scrollTop,
    viewportHeight,
    tableRef,
    setAutoFollow: setAutoFollowEnabled,
    setDetailCollapsed,
    handleScroll,
    api,
  };
}

async function withAction<T>(action: () => Promise<T>, setError: (value: string) => void) {
  try {
    setError("");
    return await action();
  } catch (error) {
    setError(String(error));
    throw error;
  }
}

function isEditableTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) {
    return false;
  }
  if (target.isContentEditable) {
    return true;
  }
  return target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.tagName === "SELECT";
}

function hasActiveTextSelection() {
  const selection = window.getSelection();
  if (!selection || selection.isCollapsed || selection.rangeCount === 0) {
    return false;
  }
  return selection.toString().length > 0;
}

const emptyState = new main.AppState({
  revision: 0,
  status: "loading",
  adbStatus: "未连接",
  sessionActive: false,
  devices: [],
  selectedDevice: "",
  packageScope: "all",
  packages: [],
  selectedPackage: "",
  totalLogs: 0,
  visibleCount: 0,
  filter: {
    draft: "",
    applied: "",
    error: "",
    activeFilterId: "",
    defaultFilterId: "",
    saved: [],
  },
  search: {
    query: "",
  },
  pause: {
    active: false,
  },
  selectedCount: 0,
  logs: [],
});

function buildPreviewSearchState(
  state: AppState,
  allLogs: PreviewLogRow[],
  query: string,
) {
  const next = main.AppState.createFrom(state);
  const selectedSourceIndex = currentPreviewSelectionSourceIndex(state);
  const compiledQuery = compilePreviewSearchQuery(query);
  const visibleLogs = !compiledQuery.active
    ? allLogs
    : allLogs.filter((item) => matchesPreviewSearch(item, compiledQuery));

  next.search.query = query;
  next.logs = visibleLogs.map((item) => main.LogItemView.createFrom({
    ...item,
    isFocused: false,
    isSelected: false,
  }));
  next.visibleCount = next.logs.length;
  syncPreviewSelection(next, { sourceIndex: selectedSourceIndex, mode: "replace" }, compiledQuery.active, allLogs);
  return next;
}

function sameSelectedLog(
  current: NonNullable<AppState["selectedLog"]>,
  next: NonNullable<AppState["selectedLog"]>,
) {
  return current.sourceIndex === next.sourceIndex &&
    current.timeText === next.timeText &&
    current.level === next.level &&
    current.tag === next.tag &&
    current.message === next.message &&
    current.source === next.source;
}

function cloneAppStateForFilterDraft(
  current: AppState,
  query: string,
) {
  const next = cloneAppStateShell(current);
  next.filter = Object.assign(
    Object.create(Object.getPrototypeOf(current.filter)),
    current.filter,
    { draft: query },
  );
  return next;
}

function setFilterDraftStateInRef(
  stateRef: MutableRefObject<AppState>,
  setState: (value: AppState | ((current: AppState) => AppState)) => void,
  query: string,
) {
  setState((current) => {
    const next = cloneAppStateForFilterDraft(current, query);
    stateRef.current = next;
    return next;
  });
}

function syncPreviewSelection(
  state: AppState,
  request: { sourceIndex: number; mode: LogSelectionMode },
  preferFirstResult: boolean,
  allLogs: PreviewLogRow[],
) {
  const targetIndex = request.sourceIndex >= 0
    ? state.logs.findIndex((item) => item.sourceIndex === request.sourceIndex)
    : -1;
  const fallbackIndex = preferFirstResult && state.logs.length > 0 ? 0 : -1;
  const nextFocusedIndex = targetIndex >= 0 ? targetIndex : fallbackIndex;
  const previousFocusedIndex = state.logs.findIndex((item) => item.isFocused);
  const previousAnchorIndex = state.logs.findIndex((item) => item.isSelected);
  const selected = new Set(
    state.logs.filter((item) => item.isSelected).map((item) => item.sourceIndex),
  );

  if (nextFocusedIndex >= 0) {
    const focusedSourceIndex = state.logs[nextFocusedIndex].sourceIndex;
    switch (request.mode) {
      case "add":
        if (selected.has(focusedSourceIndex)) {
          selected.delete(focusedSourceIndex);
        } else {
          selected.add(focusedSourceIndex);
        }
        break;
      case "range": {
        const anchor = previousAnchorIndex >= 0 ? previousAnchorIndex : previousFocusedIndex;
        const start = anchor >= 0 ? Math.min(anchor, nextFocusedIndex) : nextFocusedIndex;
        const end = anchor >= 0 ? Math.max(anchor, nextFocusedIndex) : nextFocusedIndex;
        selected.clear();
        for (let index = start; index <= end; index++) {
          selected.add(state.logs[index].sourceIndex);
        }
        break;
      }
      default:
        selected.clear();
        selected.add(focusedSourceIndex);
    }
  } else {
    selected.clear();
  }

  state.selectedCount = selected.size;
  state.logs = state.logs.map((item, index) => {
    item.isSelected = selected.has(item.sourceIndex);
    item.isFocused = index === nextFocusedIndex;
    return item;
  });
  state.selectedLog = buildPreviewSelectedLog(state.logs[nextFocusedIndex], allLogs);
}

function buildPreviewSelectedLog(log?: main.LogItemView, allLogs: PreviewLogRow[] = []) {
  if (!log) {
    return undefined;
  }
  const source = allLogs.find((item) => item.sourceIndex === log.sourceIndex)?.source ?? "";
  return {
    sourceIndex: log.sourceIndex,
    timeText: log.timeText,
    level: log.level,
    tag: log.tag,
    message: log.message,
    source,
  };
}

function formatPreviewLogDisplay(log: Pick<main.LogItemView, "timeText" | "level" | "tag" | "message">) {
  return `${log.timeText} ${log.level} ${log.tag} ${log.message}`;
}

function currentPreviewSelectionSourceIndex(state: AppState) {
  return state.logs.find((item) => item.isFocused)?.sourceIndex ?? -1;
}

function joinPreviewLogs(logs: main.LogItemView[]) {
  return logs.map((item) => formatPreviewLogDisplay(item)).join("\n");
}

function collectSelectedSourceIndexes(logs: AppState["logs"], selectedCount: number) {
  if (selectedCount <= 0) {
    return [];
  }
  const selected: number[] = [];
  for (const row of logs) {
    if (!row.isSelected) {
      continue;
    }
    selected.push(row.sourceIndex);
    if (selected.length === selectedCount) {
      break;
    }
  }
  return selected;
}

function collectSelectionChangeSources(
  previousSelected: readonly number[],
  nextSelected: readonly number[],
  previousFocused: number,
  nextFocused: number,
) {
  const changed: number[] = [];
  let previousIndex = 0;
  let nextIndex = 0;
  while (previousIndex < previousSelected.length && nextIndex < nextSelected.length) {
    const previousSource = previousSelected[previousIndex];
    const nextSource = nextSelected[nextIndex];
    if (previousSource === nextSource) {
      previousIndex++;
      nextIndex++;
      continue;
    }
    if (previousSource < nextSource) {
      changed.push(previousSource);
      previousIndex++;
      continue;
    }
    changed.push(nextSource);
    nextIndex++;
  }
  while (previousIndex < previousSelected.length) {
    changed.push(previousSelected[previousIndex]);
    previousIndex++;
  }
  while (nextIndex < nextSelected.length) {
    changed.push(nextSelected[nextIndex]);
    nextIndex++;
  }
  pushUniqueSourceIndex(changed, previousFocused);
  pushUniqueSourceIndex(changed, nextFocused);
  return changed;
}

function pushUniqueSourceIndex(sourceIndexes: number[], sourceIndex: number) {
  if (sourceIndex < 0 || sourceIndexes.includes(sourceIndex)) {
    return;
  }
  sourceIndexes.push(sourceIndex);
}

function applySelectionRows(
  logs: AppState["logs"],
  changedSourceIndexes: readonly number[],
  nextSelected: readonly number[],
  nextFocused: number,
) {
  let nextLogs = logs;
  for (const sourceIndex of changedSourceIndexes) {
    const logIndex = findLogIndexBySource(logs, sourceIndex);
    if (logIndex === -1) {
      continue;
    }
    const row = logs[logIndex];
    const isFocused = sourceIndex === nextFocused;
    const isSelected = hasSortedSourceIndex(nextSelected, sourceIndex);
    if (row.isFocused === isFocused && row.isSelected === isSelected) {
      continue;
    }
    if (nextLogs === logs) {
      nextLogs = logs.slice();
    }
    nextLogs[logIndex] = cloneLogRowWithSelection(row, isFocused, isSelected);
  }
  return nextLogs;
}

function cloneLogRowWithSelection(
  row: AppState["logs"][number],
  isFocused: boolean,
  isSelected: boolean,
) {
  return {
    sourceIndex: row.sourceIndex,
    timeText: row.timeText,
    level: row.level,
    tag: row.tag,
    message: row.message,
    isFocused,
    isSelected,
  } satisfies AppState["logs"][number];
}

function findLogIndexBySource(logs: AppState["logs"], sourceIndex: number) {
  let low = 0;
  let high = logs.length - 1;
  while (low <= high) {
    const middle = low + ((high - low) >> 1);
    const current = logs[middle].sourceIndex;
    if (current === sourceIndex) {
      return middle;
    }
    if (current < sourceIndex) {
      low = middle + 1;
      continue;
    }
    high = middle - 1;
  }
  return -1;
}

function hasSortedSourceIndex(sourceIndexes: readonly number[], sourceIndex: number) {
  let low = 0;
  let high = sourceIndexes.length - 1;
  while (low <= high) {
    const middle = low + ((high - low) >> 1);
    const current = sourceIndexes[middle];
    if (current === sourceIndex) {
      return true;
    }
    if (current < sourceIndex) {
      low = middle + 1;
      continue;
    }
    high = middle - 1;
  }
  return false;
}

function compilePreviewSearchQuery(query: string) {
  const normalized = normalizePreviewSearchQuery(query);
  if (!normalized) {
    return {
      active: false,
      literal: "",
      groups: [] as Array<Array<{ needle: string; negated: boolean }>>,
      highlightTerms: [] as string[],
    };
  }
  if (!usesPreviewSearchOperators(normalized)) {
    return {
      active: true,
      literal: normalized,
      groups: [] as Array<Array<{ needle: string; negated: boolean }>>,
      highlightTerms: [normalized],
    };
  }

  const groups = normalized
    .split("||")
    .map((group) => group
      .split("&&")
      .map((term) => buildPreviewSearchTerm(term))
      .filter((term): term is { needle: string; negated: boolean } => term !== null))
    .filter((group) => group.length > 0);

  if (groups.length === 0) {
    return {
      active: true,
      literal: normalized,
      groups: [] as Array<Array<{ needle: string; negated: boolean }>>,
      highlightTerms: [normalized],
    };
  }

  const highlightTerms = Array.from(new Set(groups
    .flatMap((group) => group.filter((term) => !term.negated).map((term) => term.needle))));

  return {
    active: true,
    literal: "",
    groups,
    highlightTerms,
  };
}

function normalizePreviewSearchQuery(query: string) {
  return query.trim().toLowerCase();
}

function usesPreviewSearchOperators(query: string) {
  return query.includes("&&") || query.includes("||") || query.startsWith("-");
}

function buildPreviewSearchTerm(term: string) {
  let next = term.trim();
  if (!next) {
    return null;
  }
  let negated = false;
  while (next.startsWith("-")) {
    negated = !negated;
    next = next.slice(1).trim();
  }
  if (!next) {
    return null;
  }
  return { needle: next, negated };
}

function matchesPreviewSearch(
  log: Pick<PreviewLogRow, "tag" | "message">,
  query: ReturnType<typeof compilePreviewSearchQuery>,
) {
  const text = `${log.tag}\n${log.message}`.toLowerCase();
  if (query.literal) {
    return text.includes(query.literal);
  }
  return query.groups.some((group) => group.every((term) => {
    const matched = text.includes(term.needle);
    return term.negated ? !matched : matched;
  }));
}

export function buildResultSearchPreview(query: string): ResultSearchPreview {
  const compiled = compilePreviewSearchQuery(query);
  return {
    query,
    highlightTerms: compiled.highlightTerms,
  };
}

function isWailsRuntime() {
  return Boolean((window as unknown as { go?: unknown; runtime?: unknown }).go) &&
    Boolean((window as unknown as { go?: unknown; runtime?: unknown }).runtime);
}

function createPreviewApi(
  state: AppState,
  setState: (state: AppState) => void,
  setError: (value: string) => void,
  allLogsRef: MutableRefObject<PreviewLogRow[]>,
) {
  return {
    selectDevice: async (_deviceID: string) => undefined,
    applySavedFilter: async (filterID: string) => {
      const next = main.AppState.createFrom(state);
      if (!filterID) {
        next.filter.activeFilterId = "";
        next.filter.draft = "";
        next.filter.applied = "";
        next.selectedPackage = "";
        setState(next);
        return;
      }
      const selected = next.filter.saved.find((filter) => filter.id === filterID);
      next.filter.activeFilterId = filterID;
      if (selected) {
        next.filter.draft = selected.query;
        next.filter.applied = selected.query;
        next.selectedPackage = selected.packageName;
      }
      setState(next);
    },
    selectPackage: async (packageName: string) => {
      const next = main.AppState.createFrom(state);
      next.selectedPackage = packageName;
      setState(next);
    },
    setPackageScope: async (scope: string) => {
      const next = main.AppState.createFrom(state);
      next.packageScope = scope || "all";
      setState(next);
    },
    setFilterDraft: async (query: string) => {
      const next = main.AppState.createFrom(state);
      next.filter.draft = query;
      setState(next);
    },
    setSearchQuery: async (query: string) => {
      const next = buildPreviewSearchState(state, allLogsRef.current, query);
      setState(next);
    },
    applyFilter: async (query?: string) => {
      const next = main.AppState.createFrom(state);
      if (query !== undefined) {
        next.filter.draft = query;
      }
      next.filter.draft = next.filter.draft.trim();
      next.filter.applied = next.filter.draft;
      setState(next);
    },
    exportVisible: async () => setError("浏览器预览模式不执行导出"),
    getSelectedLogDetail: async () => {
      const selected = state.selectedLog;
      if (!selected) {
        return undefined;
      }
      return { raw: formatPreviewLogDisplay(selected) };
    },
    copySelected: async () => undefined,
    copySelectedLogs: async () => {
      const selected = state.logs.filter((item) => item.isSelected);
      await navigator.clipboard.writeText(joinPreviewLogs(selected));
    },
    copyAllVisibleLogs: async () => {
      await navigator.clipboard.writeText(joinPreviewLogs(state.logs));
    },
    selectLog: async (index: number) => {
      const next = main.AppState.createFrom(state);
      syncPreviewSelection(
        next,
        { sourceIndex: index >= 0 ? next.logs[index]?.sourceIndex ?? -1 : -1, mode: "replace" },
        false,
        allLogsRef.current,
      );
      setState(next);
    },
    selectLogs: async (index: number, mode: LogSelectionMode) => {
      const next = main.AppState.createFrom(state);
      syncPreviewSelection(
        next,
        { sourceIndex: index >= 0 ? next.logs[index]?.sourceIndex ?? -1 : -1, mode },
        false,
        allLogsRef.current,
      );
      setState(next);
    },
    pauseToggle: async () => {
      const next = main.AppState.createFrom(state);
      next.pause.active = !next.pause.active;
      next.status = next.pause.active ? "Paused，缓存 0 条新日志" : "running";
      setState(next);
    },
    clearVisible: async () => {
      const next = main.AppState.createFrom(state);
      allLogsRef.current = [];
      next.logs = [];
      next.visibleCount = 0;
      next.selectedCount = 0;
      next.selectedLog = undefined;
      setState(next);
    },
    saveFilter: async (draft: SaveFilterDraft) => {
      const next = main.AppState.createFrom(state);
      const id = draft.name.trim().toLowerCase().replaceAll(" ", "-");
      next.filter.draft = draft.query;
      next.filter.applied = draft.query;
      next.filter.activeFilterId = id;
      next.selectedPackage = draft.packageName;
      next.filter.saved = upsertPreviewFilter(next.filter.saved, {
        id,
        name: draft.name,
        packageName: draft.packageName,
        query: draft.query,
      });
      setState(next);
      setError("");
    },
    updateFilter: async (filterID: string, draft: SaveFilterDraft) => {
      const next = main.AppState.createFrom(state);
      const id = draft.name.trim().toLowerCase().replaceAll(" ", "-");
      next.filter.draft = draft.query;
      next.filter.applied = draft.query;
      next.filter.activeFilterId = id;
      next.selectedPackage = draft.packageName;
      next.filter.saved = upsertPreviewFilter(
        next.filter.saved.filter((filter) => filter.id !== filterID),
        {
          id,
          name: draft.name,
          packageName: draft.packageName,
          query: draft.query,
        },
      );
      setState(next);
      setError("");
    },
    replaceSavedFilters: async (draft: SavedFiltersDraft) => {
      const next = main.AppState.createFrom(state);
      const renamedIDs = new Map<string, string>();
      next.filter.saved = draft.filters.map((filter) => {
        const id = filter.name.trim().toLowerCase().replaceAll(" ", "-");
        renamedIDs.set(filter.existingID, id);
        return {
          id,
          name: filter.name.trim(),
          packageName: filter.packageName.trim(),
          query: filter.query.trim(),
        };
      });
      next.filter.defaultFilterId = resolvePreviewFilterID(
        draft.defaultFilterID,
        next.filter.saved,
        renamedIDs,
      );
      next.filter.activeFilterId = resolvePreviewFilterID(
        draft.activeFilterID,
        next.filter.saved,
        renamedIDs,
      );

      const selected = next.filter.saved.find((filter) => filter.id === next.filter.activeFilterId);
      if (selected) {
        next.filter.draft = selected.query;
        next.filter.applied = selected.query;
        next.selectedPackage = selected.packageName;
      }
      if (!selected && next.filter.saved.length === 0) {
        next.filter.draft = "";
        next.filter.applied = "";
        next.selectedPackage = "";
      }

      setState(next);
      setError("");
    },
  };
}

function shouldResumeStreaming(state: Pick<AppState, "sessionActive" | "pause">) {
  return state.pause.active || !state.sessionActive;
}

function hasStartFailureStatus(status: string) {
  return status !== "" && status !== "running" && status !== "idle";
}

function upsertPreviewFilter(filters: main.SavedFilterView[], nextFilter: main.SavedFilterView) {
  const index = filters.findIndex((filter) => filter.id === nextFilter.id);
  if (index === -1) {
    return [...filters, nextFilter];
  }
  const next = [...filters];
  next[index] = nextFilter;
  return next;
}

function resolvePreviewFilterID(
  existingID: string,
  filters: main.SavedFilterView[],
  renamedIDs: Map<string, string>,
) {
  const nextID = renamedIDs.get(existingID) || "";
  return filters.some((filter) => filter.id === nextID) ? nextID : "";
}
