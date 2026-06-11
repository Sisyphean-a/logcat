import { useState } from "react";
import { DetailPanel, FilterBar, StatusBar } from "./app-shell";
import { suggestFilterName } from "./filter-rule-builder";
import { LogTable } from "./log-table";
import { type FilterDialogDraft, SaveFilterDialog } from "./save-filter-dialog";
import { SettingsDialog } from "./settings-dialog";
import { Toolbar } from "./toolbar";
import { useAppController } from "./use-app-controller";
import { useViewSettings } from "./view-settings";

export default function App() {
  const [filterDialogMode, setFilterDialogMode] = useState<"create" | "edit">("create");
  const [saveDialogOpen, setSaveDialogOpen] = useState(false);
  const [saveDialogBusy, setSaveDialogBusy] = useState(false);
  const [saveDialogError, setSaveDialogError] = useState("");
  const [settingsDialogOpen, setSettingsDialogOpen] = useState(false);
  const { settings, shellStyle, updateSetting, resetSettings } = useViewSettings();
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

  async function openCreateFilterDialog(query: string) {
    await api.setFilterDraft(query);
    setSaveDialogError("");
    setFilterDialogMode("create");
    setSaveDialogOpen(true);
  }

  return (
    <div className="app-shell" style={shellStyle}>
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
        onSetPackageScope={(scope) => void api.setPackageScope(scope)}
        onApplySavedFilter={(filterID) => void api.applySavedFilter(filterID)}
        onPauseToggle={() => void api.pauseToggle()}
        onClearVisible={() => void api.clearVisible()}
        onExport={() => void api.exportVisible()}
        onOpenSettings={() => setSettingsDialogOpen(true)}
      />

      <main className="workspace">
        <section className="viewer">
          <FilterBar
            state={state}
            autoFollow={autoFollow}
            onSelectPackage={(packageName) => void api.selectPackage(packageName)}
            onApplyFilter={(query) => void api.applyFilter(query)}
            onToggleFollow={() => setAutoFollow(!autoFollow)}
            onSaveFilter={(query) => void openCreateFilterDialog(query)}
          />

          <LogTable
            fontSize={settings.tableFontSize}
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
      <SettingsDialog
        open={settingsDialogOpen}
        settings={settings}
        onChange={updateSetting}
        onClose={() => setSettingsDialogOpen(false)}
        onReset={resetSettings}
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
