import { useCallback, useEffect, useMemo, useState } from "react";
import { main } from "../wailsjs/go/models";
import { type SaveFilterDraft } from "./filter-rule-builder";
import { SavedFilterForm } from "./saved-filter-form";
import { type SavedFilterDefinitionDraft, type SavedFiltersDraft } from "./saved-filter-types";

type SavedFiltersDialogProps = {
  defaultFilterID: string;
  errorMessage: string;
  initialFilterID?: string;
  open: boolean;
  packageOptions: string[];
  savedFilters: main.SavedFilterView[];
  saving: boolean;
  onClose: () => void;
  onSubmit: (draft: SavedFiltersDraft) => Promise<void>;
};

export function SavedFiltersDialog({
  defaultFilterID,
  errorMessage,
  initialFilterID,
  open,
  packageOptions,
  savedFilters,
  saving,
  onClose,
  onSubmit,
}: SavedFiltersDialogProps) {
  const [drafts, setDrafts] = useState<SavedFilterDefinitionDraft[]>([]);
  const [selectedFilterID, setSelectedFilterID] = useState("");
  const [selectedDefaultFilterID, setSelectedDefaultFilterID] = useState("");
  const [validationError, setValidationError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }

    const nextDrafts = savedFilters.map((filter) => ({
      existingID: filter.id,
      name: filter.name,
      packageName: filter.packageName,
      query: filter.query,
    }));
    const nextSelectedFilterID = pickSelectedFilterID(initialFilterID, nextDrafts);
    setDrafts(nextDrafts);
    setSelectedFilterID(nextSelectedFilterID);
    setSelectedDefaultFilterID(normalizeFilterID(defaultFilterID, nextDrafts));
    setValidationError("");
  }, [open]);

  useEffect(() => {
    if (!open) {
      return;
    }

    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }

    window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [onClose, open]);

  const selectedDraft = useMemo(
    () => drafts.find((draft) => draft.existingID === selectedFilterID),
    [drafts, selectedFilterID],
  );

  const handleSelectedDraftChange = useCallback((draft: SaveFilterDraft) => {
    if (!selectedFilterID) {
      return;
    }
    setDrafts((current) => current.map((item) => (
      item.existingID === selectedFilterID
        ? { existingID: selectedFilterID, ...draft }
        : item
    )));
  }, [selectedFilterID]);

  if (!open) {
    return null;
  }

  async function handleSubmit() {
    const trimmedDrafts = drafts.map((draft) => ({
      ...draft,
      name: draft.name.trim(),
      packageName: draft.packageName.trim(),
      query: draft.query.trim(),
    }));
    const nextDefaultFilterID = normalizeFilterID(selectedDefaultFilterID, trimmedDrafts);
    const nextActiveFilterID = normalizeFilterID(selectedFilterID, trimmedDrafts);

    if (trimmedDrafts.some((draft) => !draft.name)) {
      setValidationError("过滤器名称不能为空");
      return;
    }
    if (trimmedDrafts.some((draft) => !draft.packageName && !draft.query)) {
      setValidationError("每个过滤器至少配置一个包名或筛选规则");
      return;
    }

    setValidationError("");
    await onSubmit({
      activeFilterID: nextActiveFilterID,
      defaultFilterID: nextDefaultFilterID,
      filters: trimmedDrafts,
    });
  }

  function moveSelected(offset: -1 | 1) {
    setDrafts((current) => moveDraft(current, selectedFilterID, offset));
  }

  function deleteSelected() {
    setDrafts((current) => {
      const index = current.findIndex((draft) => draft.existingID === selectedFilterID);
      if (index === -1) {
        return current;
      }

      const next = current.filter((draft) => draft.existingID !== selectedFilterID);
      const nextSelectedFilterID = next[Math.min(index, next.length - 1)]?.existingID || "";
      setSelectedFilterID(nextSelectedFilterID);
      if (selectedDefaultFilterID === selectedFilterID) {
        setSelectedDefaultFilterID("");
      }
      return next;
    });
  }

  return (
    <div
      className="dialog-overlay"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <section className="dialog-card saved-filters-dialog">
        <header className="dialog-header">
          <div>
            <div className="dialog-title">管理过滤器</div>
            <div className="dialog-subtitle">切换、排序、删除并指定默认过滤器。</div>
          </div>
          <button className="ghost-button dialog-close" type="button" onClick={onClose}>
            取消
          </button>
        </header>

        <div className="saved-filters-layout">
          <aside className="saved-filters-sidebar">
            <div className="saved-filters-sidebar-title">已保存筛选</div>
            <div className="saved-filters-list">
              {drafts.map((draft) => {
                const active = draft.existingID === selectedFilterID;
                const isDefault = draft.existingID === selectedDefaultFilterID;
                return (
                  <button
                    key={draft.existingID}
                    className={`saved-filter-nav-item ${active ? "active" : ""}`}
                    type="button"
                    onClick={() => setSelectedFilterID(draft.existingID)}
                  >
                    <span className="saved-filter-nav-name">{draft.name.trim() || "未命名过滤器"}</span>
                    <span className="saved-filter-nav-meta">
                      {isDefault ? <span className="saved-filter-default-badge">默认</span> : null}
                      <span className="saved-filter-nav-package">{draft.packageName.trim() || "全部包"}</span>
                    </span>
                  </button>
                );
              })}
            </div>
          </aside>

          <div className="saved-filters-main">
            {selectedDraft ? (
              <>
                <div className="saved-filters-actions">
                  <div className="saved-filters-current">
                    <span className="saved-filters-current-name">{selectedDraft.name.trim() || "未命名过滤器"}</span>
                    {selectedDraft.existingID === selectedDefaultFilterID ? (
                      <span className="saved-filter-default-badge">默认</span>
                    ) : null}
                  </div>
                  <div className="saved-filters-action-buttons">
                    <button
                      className="ghost-button mini-button"
                      type="button"
                      disabled={saving || drafts[0]?.existingID === selectedFilterID}
                      onClick={() => moveSelected(-1)}
                    >
                      上移
                    </button>
                    <button
                      className="ghost-button mini-button"
                      type="button"
                      disabled={saving || drafts[drafts.length - 1]?.existingID === selectedFilterID}
                      onClick={() => moveSelected(1)}
                    >
                      下移
                    </button>
                    <button
                      className="ghost-button mini-button"
                      type="button"
                      disabled={saving || selectedDraft.existingID === selectedDefaultFilterID}
                      onClick={() => setSelectedDefaultFilterID(selectedDraft.existingID)}
                    >
                      设为默认
                    </button>
                    <button
                      className="ghost-button mini-button"
                      type="button"
                      disabled={saving}
                      onClick={deleteSelected}
                    >
                      删除
                    </button>
                  </div>
                </div>

                <div className="dialog-body saved-filters-form-body">
                  <SavedFilterForm
                    draftKey={selectedDraft.existingID}
                    initialName={selectedDraft.name}
                    initialPackageName={selectedDraft.packageName}
                    initialQuery={selectedDraft.query}
                    packageOptions={packageOptions}
                    onChange={handleSelectedDraftChange}
                  />
                </div>
              </>
            ) : (
              <div className="saved-filters-empty">暂无已保存筛选</div>
            )}
          </div>
        </div>

        <footer className="dialog-footer">
          {validationError || errorMessage ? (
            <div className="saved-filters-footer-error">{validationError || errorMessage}</div>
          ) : <div />}
          <div className="saved-filters-footer-actions">
            <button className="text-button secondary" type="button" onClick={onClose} disabled={saving}>
              关闭
            </button>
            <button className="text-button primary" type="button" onClick={() => void handleSubmit()} disabled={saving}>
              {saving ? "保存中…" : "保存修改"}
            </button>
          </div>
        </footer>
      </section>
    </div>
  );
}

function pickSelectedFilterID(
  initialFilterID: string | undefined,
  drafts: SavedFilterDefinitionDraft[],
) {
  return normalizeFilterID(initialFilterID, drafts) || drafts[0]?.existingID || "";
}

function normalizeFilterID(
  filterID: string | undefined,
  drafts: SavedFilterDefinitionDraft[],
) {
  if (!filterID) {
    return "";
  }
  return drafts.some((draft) => draft.existingID === filterID) ? filterID : "";
}

function moveDraft(
  drafts: SavedFilterDefinitionDraft[],
  selectedFilterID: string,
  offset: -1 | 1,
) {
  const index = drafts.findIndex((draft) => draft.existingID === selectedFilterID);
  const nextIndex = index + offset;
  if (index === -1 || nextIndex < 0 || nextIndex >= drafts.length) {
    return drafts;
  }

  const next = [...drafts];
  const [selectedDraft] = next.splice(index, 1);
  next.splice(nextIndex, 0, selectedDraft);
  return next;
}
