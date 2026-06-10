import { useState } from "react";
import { DetailPanel, FilterBar, StatusBar, Toolbar } from "./app-shell";
import { suggestFilterName } from "./filter-rule-builder";
import { LogTable } from "./log-table";
import { type FilterDialogDraft, SaveFilterDialog } from "./save-filter-dialog";
import { useAppController } from "./use-app-controller";

export default function App() {
  const [filterDialogMode, setFilterDialogMode] = useState<"create" | "edit">("create");
  const [saveDialogOpen, setSaveDialogOpen] = useState(false);
  const [saveDialogBusy, setSaveDialogBusy] = useState(false);
  const [saveDialogError, setSaveDialogError] = useState("");
  const {
    state,
    loading,
    actionError,
    autoFollow,
    detailCollapsed,
    scrollTop,
    viewportHeight,
    tableRef,
    setAutoFollow,
    setDetailCollapsed,
    handleScroll,
    api,
  } = useAppController();
  const activeFilter = state.filter.saved.find((item) => item.id === state.filter.activeFilterId);

  async function handleSaveFilter(draft: FilterDialogDraft) {
    setSaveDialogBusy(true);
    setSaveDialogError("");
    try {
      if (draft.existingID) {
        await api.updateFilter(draft.existingID, draft);
      } else {
        await api.saveFilter(draft);
      }
      setSaveDialogOpen(false);
    } catch (error) {
      setSaveDialogError(String(error));
    } finally {
      setSaveDialogBusy(false);
    }
  }

  return (
    <div className="app-shell">
      <Toolbar
        canEditSavedFilter={Boolean(activeFilter)}
        state={state}
        onEditSavedFilter={() => {
          if (!activeFilter) {
            return;
          }
          setSaveDialogError("");
          setFilterDialogMode("edit");
          setSaveDialogOpen(true);
        }}
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
            onToggleFollow={() => setAutoFollow(!autoFollow)}
            onSaveFilter={() => {
              setSaveDialogError("");
              setFilterDialogMode("create");
              setSaveDialogOpen(true);
            }}
          />

          <LogTable
            loading={loading}
            logs={state.logs}
            visibleCount={state.visibleCount}
            scrollTop={scrollTop}
            viewportHeight={viewportHeight}
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
      <SaveFilterDialog
        errorMessage={saveDialogError}
        initialFilterID={filterDialogMode === "edit" ? activeFilter?.id : undefined}
        initialName={filterDialogMode === "edit"
          ? activeFilter?.name || suggestFilterName(state.selectedPackage, state.filter.draft)
          : suggestFilterName(state.selectedPackage, state.filter.draft)}
        initialPackageName={filterDialogMode === "edit" ? activeFilter?.packageName || "" : state.selectedPackage}
        initialQuery={filterDialogMode === "edit" ? activeFilter?.query || state.filter.draft : state.filter.draft}
        mode={filterDialogMode}
        open={saveDialogOpen}
        packageOptions={state.packages.map((pkg) => pkg.name)}
        saving={saveDialogBusy}
        onClose={() => {
          if (!saveDialogBusy) {
            setSaveDialogOpen(false);
          }
        }}
        onSubmit={handleSaveFilter}
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
