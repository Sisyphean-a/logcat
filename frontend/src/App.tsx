import { startTransition, useEffect, useMemo, useState } from "react";
import { DetailPanel, FilterBar, StatusBar } from "./app-shell";
import { suggestFilterName } from "./filter-rule-builder";
import { LogTable } from "./log-table";
import { SaveFilterDialog } from "./save-filter-dialog";
import { SavedFiltersDialog } from "./saved-filters-dialog";
import { type SavedFiltersDraft } from "./saved-filter-types";
import { SettingsDialog } from "./settings-dialog";
import { Toolbar } from "./toolbar";
import { buildResultSearchPreview, type SelectedLogDetail, useAppController } from "./use-app-controller";
import { useViewSettings } from "./view-settings";

export default function App() {
  const [saveDialogOpen, setSaveDialogOpen] = useState(false);
  const [saveDialogBusy, setSaveDialogBusy] = useState(false);
  const [saveDialogError, setSaveDialogError] = useState("");
  const [manageDialogOpen, setManageDialogOpen] = useState(false);
  const [manageDialogBusy, setManageDialogBusy] = useState(false);
  const [manageDialogError, setManageDialogError] = useState("");
  const [settingsDialogOpen, setSettingsDialogOpen] = useState(false);
  const [selectedDetail, setSelectedDetail] = useState<SelectedLogDetail>();
  const { settings, shellStyle, updateSetting, updateTheme, resetSettings } = useViewSettings();
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
  const resultSearch = useMemo(
    () => buildResultSearchPreview(state.search.query),
    [state.search.query],
  );

  useEffect(() => {
    if (!state.selectedLog) {
      startTransition(() => setSelectedDetail(undefined));
    } else {
      startTransition(() => setSelectedDetail(undefined));
    }
  }, [state.selectedLog?.sourceIndex]);

  async function handleSaveFilter(draft: { name: string; packageName: string; query: string }) {
    setSaveDialogBusy(true);
    setSaveDialogError("");
    try {
      await api.saveFilter(draft);
      setSaveDialogOpen(false);
    } catch (error) {
      setSaveDialogError(String(error));
    } finally {
      setSaveDialogBusy(false);
    }
  }

  async function handleManageFilters(draft: SavedFiltersDraft) {
    setManageDialogBusy(true);
    setManageDialogError("");
    try {
      await api.replaceSavedFilters(draft);
      setManageDialogOpen(false);
    } catch (error) {
      setManageDialogError(String(error));
    } finally {
      setManageDialogBusy(false);
    }
  }

  async function openCreateFilterDialog(query: string) {
    await api.setFilterDraft(query);
    setSaveDialogError("");
    setSaveDialogOpen(true);
  }

  return (
    <div className="app-shell" data-theme={settings.theme} style={shellStyle}>
      <Toolbar
        canEditSavedFilter={state.filter.saved.length > 0}
        state={state}
        onEditSavedFilter={() => {
          if (state.filter.saved.length === 0) {
            return;
          }
          setManageDialogError("");
          setManageDialogOpen(true);
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
            onSearch={(query) => void api.setSearchQuery(query)}
            onToggleFollow={() => setAutoFollow(!autoFollow)}
            onSaveFilter={(query) => void openCreateFilterDialog(query)}
          />

          <LogTable
            fontSize={settings.tableFontSize}
            loading={loading}
            logs={state.logs}
            resultSearch={resultSearch}
            selectedCount={state.selectedCount}
            visibleCount={state.visibleCount}
            scrollTop={scrollTop}
            viewportHeight={viewportHeight}
            tableRef={tableRef}
            onScroll={handleScroll}
            onSelectLog={(index, mode) => void api.selectLogs(index, mode)}
            onCopySelected={() => void api.copySelectedLogs()}
            onCopyAll={() => void api.copyAllVisibleLogs()}
            onClearVisible={() => void api.clearVisible()}
          />
        </section>

        <DetailPanel
          state={state}
          collapsed={detailCollapsed}
          detail={selectedDetail}
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
        initialName={suggestFilterName(state.selectedPackage, state.filter.draft)}
        initialPackageName={state.selectedPackage}
        initialQuery={state.filter.draft}
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
      <SavedFiltersDialog
        defaultFilterID={state.filter.defaultFilterId}
        errorMessage={manageDialogError}
        initialFilterID={pickManagedFilterID(state)}
        open={manageDialogOpen}
        packageOptions={state.packages.map((pkg) => pkg.name)}
        savedFilters={state.filter.saved}
        saving={manageDialogBusy}
        onClose={() => {
          if (!manageDialogBusy) {
            setManageDialogOpen(false);
          }
        }}
        onSubmit={handleManageFilters}
      />
      <SettingsDialog
        open={settingsDialogOpen}
        settings={settings}
        onChange={updateSetting}
        onClose={() => setSettingsDialogOpen(false)}
        onThemeChange={updateTheme}
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

function pickManagedFilterID(state: ReturnType<typeof useAppController>["state"]) {
  return state.filter.activeFilterId || state.filter.defaultFilterId || state.filter.saved[0]?.id;
}
