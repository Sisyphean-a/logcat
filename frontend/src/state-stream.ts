import { main } from "../wailsjs/go/models";

export type AppState = main.AppState;

export type StateAppendPatch = {
  revision: number;
  totalLogs: number;
  visibleCount: number;
  dropped: number;
  appended: AppState["logs"];
  selectedCount: number;
  selectedLog?: AppState["selectedLog"];
};

export function cloneAppStateShell(current: AppState) {
  return Object.assign(Object.create(Object.getPrototypeOf(current)), current) as AppState;
}

export function applyStateAppendPatch(current: AppState, patch: StateAppendPatch) {
  const next = cloneAppStateShell(current);
  next.revision = patch.revision;
  next.totalLogs = patch.totalLogs;
  next.visibleCount = patch.visibleCount;
  next.selectedCount = patch.selectedCount;
  next.selectedLog = patch.selectedLog ?? current.selectedLog;
  next.logs = mergeAppendedLogs(current.logs, patch.dropped, patch.appended);
  return next;
}

function mergeAppendedLogs(
  currentLogs: AppState["logs"],
  dropped: number,
  appended: AppState["logs"],
) {
  const retainedStart = Math.min(currentLogs.length, Math.max(0, dropped));
  const retainedCount = currentLogs.length - retainedStart;
  const nextLogs = new Array<AppState["logs"][number]>(retainedCount + appended.length);
  for (let index = 0; index < retainedCount; index++) {
    nextLogs[index] = currentLogs[retainedStart + index];
  }
  for (let index = 0; index < appended.length; index++) {
    nextLogs[retainedCount + index] = appended[index];
  }
  return nextLogs;
}
