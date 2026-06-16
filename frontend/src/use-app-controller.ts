import { useEffect, useMemo, useRef, useState, type MutableRefObject } from "react";
import { EventsOff, EventsOn } from "../wailsjs/runtime/runtime";
import { app, main } from "../wailsjs/go/models";
import { createMockState } from "./mock-state";
import {
  ApplyFilterDraft,
  ApplySavedFilter,
  ClearVisible,
  CopyAllVisibleLogs,
  CopySelectedLogs,
  CopyText,
  ExportVisibleLogs,
  GetState,
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

const stateEventName = "state:updated";

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
  const previewAllLogsRef = useRef<main.LogItemView[]>([]);
  const stateRef = useRef(state);
  stateRef.current = state;

  function syncTableMetrics(node: HTMLDivElement) {
    setScrollTop(node.scrollTop);
    setViewportHeight(node.clientHeight);
  }

  useEffect(() => {
    autoFollowRef.current = autoFollow;
  }, [autoFollow]);

  useEffect(() => {
    if (!isWailsRuntime()) {
      const snapshot = createMockState();
      previewAllLogsRef.current = snapshot.logs.map((item) => main.LogItemView.createFrom(item));
      setState(snapshot);
      setLoading(false);
      return;
    }

    let mounted = true;

    GetState()
      .then((next) => {
        if (!mounted) {
          return;
        }
        setState(next);
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
      setState(next);
    };

    EventsOn(stateEventName, handler);
    return () => {
      mounted = false;
      EventsOff(stateEventName);
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
    syncTableMetrics(node);
  }, [state.logs.length]);

  const api = useMemo(
    () => (isWailsRuntime() ? {
      selectDevice: (deviceID: string) => withAction(() => SelectDevice(deviceID), setActionError),
      applySavedFilter: (filterID: string) => withAction(() => ApplySavedFilter(filterID), setActionError),
      selectPackage: (packageName: string) => withAction(() => SelectPackage(packageName), setActionError),
      setPackageScope: (scope: string) => withAction(() => SetPackageScope(scope), setActionError),
      setFilterDraft: (query: string) =>
        SetFilterDraft(query).then((next: AppState) => setState(next)),
      setSearchQuery: (query: string) =>
        SetSearchQuery(query).then((next: AppState) => setState(next)),
      applyFilter: async (query?: string) => {
        if (query !== undefined) {
          const next = await SetFilterDraft(query);
          setState(next);
        }
        await withAction(ApplyFilterDraft, setActionError);
      },
      exportVisible: () => withAction(ExportVisibleLogs, setActionError),
      copySelected: async (kind: "display" | "raw" | "message") => {
        const selected = stateRef.current.selectedLog;
        if (!selected) {
          return;
        }
        const value = kind === "raw" ? selected.raw : kind === "message" ? selected.message : selected.display;
        await withAction(() => CopyText(value), setActionError);
      },
      copySelectedLogs: () => withAction(CopySelectedLogs, setActionError),
      copyAllVisibleLogs: () => withAction(CopyAllVisibleLogs, setActionError),
      selectLog: (index: number) =>
        SelectLog(index).then((next: AppState) => setState(next)),
      selectLogs: (index: number, mode: LogSelectionMode) =>
        SelectLogs({ index, mode }).then((next: AppState) => setState(next)),
      pauseToggle: async () => {
        const next = stateRef.current.pause.active ? await ResumeKeep() : await Pause();
        setState(next);
      },
      clearVisible: () =>
        ClearVisible().then((next: AppState) => setState(next)),
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

    syncTableMetrics(node);
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
      syncTableMetrics(currentNode);
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

const emptyState = new main.AppState({
  status: "loading",
  adbStatus: "未连接",
  devices: [],
  selectedDevice: "",
  packageScope: "all",
  packages: [],
  selectedPackage: "",
  processes: [],
  selectedProcess: "",
  boundPids: [],
  totalLogs: 0,
  visibleCount: 0,
  visibleStart: 0,
  filter: {
    draft: "",
    applied: "",
    error: "",
    activeFilterId: "",
    defaultFilterId: "",
    saved: [],
    history: [],
  },
  search: {
    query: "",
    matchIndexes: [],
    current: -1,
  },
  pause: {
    active: false,
    bufferedCount: 0,
    droppedCount: 0,
  },
  selectedIndex: -1,
  selectedCount: 0,
  logs: [],
});

function buildPreviewSearchState(
  state: AppState,
  allLogs: main.LogItemView[],
  query: string,
) {
  const next = main.AppState.createFrom(state);
  const selectedRaw = currentPreviewSelectionRaw(state);
  const normalizedQuery = normalizePreviewSearchQuery(query);
  const visibleLogs = normalizedQuery
    ? allLogs.filter((item) => matchesPreviewSearch(item, normalizedQuery))
    : allLogs;

  next.search.query = query;
  next.logs = visibleLogs.map((item, index) => main.LogItemView.createFrom({
    ...item,
    index,
    isMatch: false,
    isCurrent: false,
    isFocused: false,
    isSelected: false,
  }));
  next.visibleCount = next.logs.length;
  next.visibleStart = 0;
  syncPreviewSelection(next, { raw: selectedRaw, mode: "replace" }, Boolean(normalizedQuery));
  return next;
}

function syncPreviewSelection(
  state: AppState,
  request: { raw: string; mode: LogSelectionMode },
  preferFirstResult: boolean,
) {
  const hasSearch = normalizePreviewSearchQuery(state.search.query).length > 0;
  state.search.matchIndexes = hasSearch ? state.logs.map((item) => item.index) : [];
  const targetIndex = request.raw
    ? state.logs.findIndex((item) => item.raw === request.raw)
    : -1;
  const fallbackIndex = preferFirstResult && state.logs.length > 0 ? 0 : -1;
  const nextFocusedIndex = targetIndex >= 0 ? targetIndex : fallbackIndex;
  const previousFocusedIndex = state.logs.findIndex((item) => item.isFocused);
  const previousAnchorIndex = state.logs.findIndex((item) => item.isSelected);
  const selected = new Set(
    state.logs.filter((item) => item.isSelected).map((item) => item.raw),
  );

  if (nextFocusedIndex >= 0) {
    const focusedRaw = state.logs[nextFocusedIndex].raw;
    switch (request.mode) {
      case "add":
        if (selected.has(focusedRaw)) {
          selected.delete(focusedRaw);
        } else {
          selected.add(focusedRaw);
        }
        break;
      case "range": {
        const anchor = previousAnchorIndex >= 0 ? previousAnchorIndex : previousFocusedIndex;
        const start = anchor >= 0 ? Math.min(anchor, nextFocusedIndex) : nextFocusedIndex;
        const end = anchor >= 0 ? Math.max(anchor, nextFocusedIndex) : nextFocusedIndex;
        selected.clear();
        for (let index = start; index <= end; index++) {
          selected.add(state.logs[index].raw);
        }
        break;
      }
      default:
        selected.clear();
        selected.add(focusedRaw);
    }
  } else {
    selected.clear();
  }

  state.selectedIndex = nextFocusedIndex;
  state.selectedCount = selected.size;
  state.search.current = hasSearch && nextFocusedIndex >= 0 ? nextFocusedIndex : -1;
  state.logs = state.logs.map((item, index) => {
    item.isSelected = selected.has(item.raw);
    item.isFocused = index === nextFocusedIndex;
    item.isCurrent = hasSearch && index === state.search.current;
    item.isMatch = false;
    return item;
  });
  state.selectedLog = buildPreviewSelectedLog(state.logs[nextFocusedIndex]);
}

function buildPreviewSelectedLog(log?: main.LogItemView) {
  if (!log) {
    return undefined;
  }
  return {
    index: log.index,
    timeText: log.timeText,
    level: log.level,
    tag: log.tag,
    message: log.message,
    source: log.source,
    raw: log.raw,
    display: log.display ?? formatPreviewLogDisplay(log),
  };
}

function formatPreviewLogDisplay(log: Pick<main.LogItemView, "timeText" | "level" | "tag" | "message">) {
  return `${log.timeText} ${log.level} ${log.tag} ${log.message}`;
}

function currentPreviewSelectionRaw(state: AppState) {
  if (state.selectedLog?.raw) {
    return state.selectedLog.raw;
  }
  return state.selectedIndex >= 0 ? state.logs[state.selectedIndex]?.raw || "" : "";
}

function joinPreviewLogs(logs: main.LogItemView[]) {
  return logs.map((item) => item.display ?? formatPreviewLogDisplay(item)).join("\n");
}

function normalizePreviewSearchQuery(query: string) {
  return query.trim().toLowerCase();
}

function matchesPreviewSearch(log: main.LogItemView, query: string) {
  return `${log.tag}\n${log.message}`.toLowerCase().includes(query);
}

function isWailsRuntime() {
  return Boolean((window as unknown as { go?: unknown; runtime?: unknown }).go) &&
    Boolean((window as unknown as { go?: unknown; runtime?: unknown }).runtime);
}

function createPreviewApi(
  state: AppState,
  setState: (state: AppState) => void,
  setError: (value: string) => void,
  allLogsRef: MutableRefObject<main.LogItemView[]>,
) {
  return {
    selectDevice: async (_deviceID: string) => undefined,
    applySavedFilter: async (filterID: string) => {
      const next = main.AppState.createFrom(state);
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
      syncPreviewSelection(next, { raw: index >= 0 ? next.logs[index]?.raw || "" : "", mode: "replace" }, false);
      setState(next);
    },
    selectLogs: async (index: number, mode: LogSelectionMode) => {
      const next = main.AppState.createFrom(state);
      syncPreviewSelection(next, { raw: index >= 0 ? next.logs[index]?.raw || "" : "", mode }, false);
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
      next.visibleStart = 0;
      next.selectedIndex = -1;
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
