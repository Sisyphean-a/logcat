import { DetailPanel, FilterBar, LogTable, StatusBar, Toolbar } from "./app-shell";
import { useAppController } from "./use-app-controller";

export default function App() {
  const {
    state,
    loading,
    actionError,
    autoFollow,
    detailCollapsed,
    tableRef,
    setAutoFollow,
    setDetailCollapsed,
    handleScroll,
    api,
  } = useAppController();

  return (
    <div className="app-shell">
      <Toolbar
        state={state}
        onSelectDevice={(deviceID) => void api.selectDevice(deviceID)}
        onApplySavedFilter={(filterID) => void api.applySavedFilter(filterID)}
        onPauseToggle={() => void api.pauseToggle()}
        onClearVisible={() => void api.clearVisible()}
        onExport={() => void api.exportVisible()}
      />

      <main className="workspace">
        <section className="viewer">
          <FilterBar
            state={state}
            autoFollow={autoFollow}
            onSelectPackage={(packageName) => void api.selectPackage(packageName)}
            onFilterDraftChange={(query) => void api.setFilterDraft(query)}
            onApplyFilter={() => void api.applyFilter()}
            onSetPackageScope={(scope) => void api.setPackageScope(scope)}
            onToggleFollow={() => setAutoFollow((value) => !value)}
            onSaveFilter={() => void api.saveFilter()}
          />

          <LogTable
            loading={loading}
            logs={state.logs}
            tableRef={tableRef}
            onScroll={handleScroll}
            onSelectLog={(index) => void api.selectLog(index)}
          />
        </section>

        <DetailPanel
          state={state}
          collapsed={detailCollapsed}
          onToggle={() => setDetailCollapsed((value) => !value)}
          onCopyDisplay={() => void api.copySelected("display")}
          onCopyRaw={() => void api.copySelected("raw")}
          onCopyMessage={() => void api.copySelected("message")}
        />
      </main>

      <StatusBar
        state={state}
        autoFollow={autoFollow}
        statusText={displayStatus(actionError, state.filter.error, state.status)}
      />

    </div>
  );
}

function displayStatus(actionError: string, filterError: string, status: string) {
  if (actionError) {
    return actionError;
  }
  if (filterError) {
    return filterError;
  }
  switch (status) {
    case "":
    case "running":
      return "";
    default:
      if (status.startsWith("adb ")) {
        return "";
      }
      return status;
  }
}
