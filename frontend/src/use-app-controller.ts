import { useEffect, useMemo, useRef, useState } from "react";
import { EventsOff, EventsOn } from "../wailsjs/runtime/runtime";
import { main } from "../wailsjs/go/models";
import { createMockState } from "./mock-state";
import {
  ApplyFilterDraft,
  ApplySavedFilter,
  ClearVisible,
  CopyText,
  ExportVisibleLogs,
  GetState,
  Pause,
  ResumeKeep,
  SaveFilterDefinition,
  SelectDevice,
  SelectLog,
  SelectPackage,
  SetFilterDraft,
  SetPackageScope,
  UpdateSavedFilterDefinition,
} from "../wailsjs/go/main/App";
import { type SaveFilterDraft } from "./filter-rule-builder";

export type AppState = main.AppState;

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

  function syncTableMetrics(node: HTMLDivElement) {
    setScrollTop(node.scrollTop);
    setViewportHeight(node.clientHeight);
  }

  useEffect(() => {
    autoFollowRef.current = autoFollow;
  }, [autoFollow]);

  useEffect(() => {
    if (!isWailsRuntime()) {
      setState(createMockState());
      setLoading(false);
      return;
    }

    let mounted = true;

    GetState()
      .then((next) => {
        if (!mounted) {
          return;
        }
        const snapshot = main.AppState.createFrom(next);
        setState(snapshot);
        setLoading(false);
      })
      .catch((error: unknown) => {
        if (!mounted) {
          return;
        }
        setActionError(String(error));
        setLoading(false);
      });

    const handler = (next: AppState) => {
      const snapshot = main.AppState.createFrom(next);
      console.debug("state:updated", {
        selectedDevice: snapshot.selectedDevice,
        devices: snapshot.devices.map((item) => ({
          id: item.id,
          model: item.model,
          status: item.status,
        })),
        filterDraft: snapshot.filter.draft,
      });
      setState(snapshot);
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
        SetFilterDraft(query).then((next: AppState) => setState(main.AppState.createFrom(next))),
      applyFilter: () => withAction(ApplyFilterDraft, setActionError),
      exportVisible: () => withAction(ExportVisibleLogs, setActionError),
      copySelected: async (kind: "display" | "raw" | "message") => {
        const selected = state.selectedLog;
        if (!selected) {
          return;
        }
        const value = kind === "raw" ? selected.raw : kind === "message" ? selected.message : selected.display;
        await withAction(() => CopyText(value), setActionError);
      },
      selectLog: (index: number) =>
        SelectLog(index).then((next: AppState) => setState(main.AppState.createFrom(next))),
      pauseToggle: async () => {
        const next = state.pause.active ? await ResumeKeep() : await Pause();
        setState(main.AppState.createFrom(next));
      },
      clearVisible: () =>
        ClearVisible().then((next: AppState) => setState(main.AppState.createFrom(next))),
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
    } : createPreviewApi(state, setState, setActionError)),
    [state],
  );

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
  logs: [],
});

function isWailsRuntime() {
  return Boolean((window as unknown as { go?: unknown; runtime?: unknown }).go) &&
    Boolean((window as unknown as { go?: unknown; runtime?: unknown }).runtime);
}

function createPreviewApi(
  state: AppState,
  setState: (state: AppState) => void,
  setError: (value: string) => void,
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
    applyFilter: async () => {
      const next = main.AppState.createFrom(state);
      next.filter.applied = next.filter.draft;
      setState(next);
    },
    exportVisible: async () => setError("浏览器预览模式不执行导出"),
    copySelected: async () => undefined,
    selectLog: async (index: number) => {
      const next = main.AppState.createFrom(state);
      next.selectedIndex = index;
      next.logs = next.logs.map((item) => {
        item.isSelected = item.index === index;
        return item;
      });
      const current = next.logs.find((item) => item.index === index);
      next.selectedLog = current ? {
        index: current.index,
        timeText: current.timeText,
        level: current.level,
        tag: current.tag,
        message: current.message,
        source: current.source,
        raw: current.raw,
        display: current.display,
      } : undefined;
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
      next.logs = [];
      next.visibleCount = 0;
      next.visibleStart = 0;
      next.selectedIndex = -1;
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
