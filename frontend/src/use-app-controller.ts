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

export type AppState = main.AppState;
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

type StateAppendPatch = {
  revision: number;
  totalLogs: number;
  visibleCount: number;
  dropped: number;
  appended: AppState["logs"];
  selectedCount: number;
  selectedLog?: AppState["selectedLog"];
};

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
  const pendingEventStateRef = useRef<AppState | null>(null);
  const pendingEventFrameRef = useRef<number | null>(null);
  const pendingMetricsNodeRef = useRef<HTMLDivElement | null>(null);
  const pendingMetricsFrameRef = useRef<number | null>(null);
  stateRef.current = state;
  latestRevisionRef.current = state.revision;

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

  function queueEventState(next: AppState) {
    if (next.revision < latestRevisionRef.current) {
      return;
    }
    const pending = pendingEventStateRef.current;
    if (!pending || next.revision >= pending.revision) {
      pendingEventStateRef.current = next;
    }
    if (pendingEventFrameRef.current !== null) {
      return;
    }
    pendingEventFrameRef.current = requestAnimationFrame(flushPendingEventState);
  }

  function applyAppendPatch(patch: StateAppendPatch) {
    if (patch.revision < latestRevisionRef.current) {
      return;
    }
    startTransition(() => {
      setState((current) => {
        if (patch.revision < current.revision) {
          return current;
        }
        const retained = patch.dropped > 0 ? current.logs.slice(patch.dropped) : current.logs.slice();
        const appended = patch.appended.map((row) => main.LogItemView.createFrom(row));
        const next = cloneAppStateShell(current);
        next.revision = patch.revision;
        next.totalLogs = patch.totalLogs;
        next.visibleCount = patch.visibleCount;
        next.selectedCount = patch.selectedCount;
        next.selectedLog = patch.selectedLog
          ? main.SelectedLogView.createFrom(patch.selectedLog)
          : undefined;
        next.logs = retained.concat(appended);
        latestRevisionRef.current = patch.revision;
        return next;
      });
    });
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
        const selected = new Set(patch.selectedSourceIndexes);
        let changed = false;
        const nextLogs = current.logs.map((row) => {
          const isFocused = row.sourceIndex === patch.focusedSourceIndex;
          const isSelected = selected.has(row.sourceIndex);
          if (row.isFocused === isFocused && row.isSelected === isSelected) {
            return row;
          }
          changed = true;
          return main.LogItemView.createFrom({
            ...row,
            isFocused,
            isSelected,
          });
        });
        const nextSelectedLog = patch.selectedLog
          ? current.selectedLog && sameSelectedLog(current.selectedLog, patch.selectedLog)
            ? current.selectedLog
            : main.SelectedLogView.createFrom(patch.selectedLog)
          : undefined;
        latestRevisionRef.current = patch.revision;
        const next = cloneAppStateShell(current);
        next.revision = patch.revision;
        if (!changed && current.selectedCount === patch.selectedCount && current.selectedLog === nextSelectedLog) {
          return next;
        }
        next.selectedCount = patch.selectedCount;
        next.selectedLog = nextSelectedLog;
        next.logs = nextLogs;
        return next;
      });
    });
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
      setSearchQuery: (query: string) =>
        SetSearchQuery(query).then((next: AppState) => applyNextState(next)),
      applyFilter: async (query?: string) => {
        if (query !== undefined) {
          const next = await SetFilterDraft(query);
          applyNextState(next, { urgent: true });
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
        const next = stateRef.current.pause.active ? await ResumeKeep() : await Pause();
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

function cloneAppStateShell(current: AppState) {
  return Object.assign(Object.create(Object.getPrototypeOf(current)), current) as AppState;
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
