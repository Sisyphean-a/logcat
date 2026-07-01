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
  const [saveDialogInitialQuery, setSaveDialogInitialQuery] = useState("");
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
  const actionNotice = actionError ? humanizeStatus(actionError) : "";
  const statusText = displayStatus(actionError, state.filter.error, state.status);

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

  function openCreateFilterDialog(query: string) {
    setSaveDialogInitialQuery(query);
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
      {actionNotice && <div className="action-banner">{actionNotice}</div>}

      <main className="workspace">
        <section className="viewer">
          <FilterBar
            state={state}
            autoFollow={autoFollow}
            onSelectPackage={(packageName) => void api.selectPackage(packageName)}
            onApplyFilter={(query) => void api.applyFilter(query)}
            onSearch={(query) => void api.setSearchQuery(query)}
            onToggleFollow={() => setAutoFollow(!autoFollow)}
            onSaveFilter={openCreateFilterDialog}
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
        statusText={statusText}
      />
      <SaveFilterDialog
        errorMessage={saveDialogError}
        initialName={suggestFilterName(state.selectedPackage, saveDialogInitialQuery)}
        initialPackageName={state.selectedPackage}
        initialQuery={saveDialogInitialQuery}
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
    return humanizeStatus(actionError);
  }
  if (filterError) {
    return filterError;
  }
  switch (status) {
    case "":
    case "running":
    case "idle":
      return "";
    default:
      if (status.startsWith("adb ")) {
        return "";
      }
      return humanizeStatus(status);
  }
}

function humanizeStatus(status: string) {
  const normalized = status.startsWith("Error: ") ? status.slice(7) : status;
  if (normalized === "device_not_selected") {
    return "未选择设备，无法开始";
  }
  if (normalized === "foreground_package_not_found") {
    return "未找到前台应用包名";
  }
  if (normalized === "runtime_not_ready") {
    return "运行时未就绪";
  }
  return humanizeStatusWithDetail(normalized);
}

function humanizeStatusWithDetail(status: string) {
  const [code, detail = ""] = status.split(": ", 2);
  switch (code) {
    case "app_not_running":
      return `应用未运行：${detail}`;
    case "process_not_running":
      return `进程未运行：${detail}`;
    case "device_offline":
      return `设备离线：${detail}`;
    case "device_unauthorized":
      return `设备未授权：${detail}`;
    case "device_no_permission":
      return `设备无权限：${detail}`;
    case "device_unavailable":
      return `设备不可用：${detail}`;
    default:
      return status;
  }
}

function pickManagedFilterID(state: ReturnType<typeof useAppController>["state"]) {
  return state.filter.activeFilterId || state.filter.defaultFilterId || state.filter.saved[0]?.id;
}
