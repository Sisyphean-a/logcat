import { useEffect, useMemo, useState } from "react";
import {
  buildFilterQuery,
  createCondition,
  createRuleGroup,
  createRuleGroups,
  parseFilterQuery,
  type RuleCondition,
  type RuleGroup,
  type SaveFilterDraft,
} from "./filter-rule-builder";
import { RuleGroupsEditor } from "./filter-rule-editor";

export type FilterDialogDraft = SaveFilterDraft & {
  existingID?: string;
};

type SaveFilterDialogProps = {
  errorMessage: string;
  initialFilterID?: string;
  initialName: string;
  initialPackageName: string;
  initialQuery: string;
  mode: "create" | "edit";
  open: boolean;
  packageOptions: string[];
  saving: boolean;
  onClose: () => void;
  onSubmit: (draft: FilterDialogDraft) => Promise<void>;
};

export function SaveFilterDialog({
  errorMessage,
  initialFilterID,
  initialName,
  initialPackageName,
  initialQuery,
  mode,
  open,
  packageOptions,
  saving,
  onClose,
  onSubmit,
}: SaveFilterDialogProps) {
  const [name, setName] = useState(initialName);
  const [packageName, setPackageName] = useState(initialPackageName);
  const [groups, setGroups] = useState<RuleGroup[]>(createRuleGroups);
  const [queryOverride, setQueryOverride] = useState("");
  const [validationError, setValidationError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }

    const parsedGroups = parseFilterQuery(initialQuery);
    setName(initialName);
    setPackageName(initialPackageName);
    setGroups(parsedGroups ?? createRuleGroups());
    setQueryOverride(parsedGroups ? "" : initialQuery.trim());
    setValidationError("");
  }, [initialName, initialPackageName, initialQuery, open]);

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

  const previewQuery = useMemo(() => {
    const generated = buildFilterQuery(groups);
    return queryOverride.trim() || generated;
  }, [groups, queryOverride]);

  if (!open) {
    return null;
  }

  async function handleSubmit() {
    const trimmedName = name.trim();
    const trimmedPackage = packageName.trim();
    const trimmedQuery = previewQuery.trim();

    if (!trimmedName) {
      setValidationError("请填写过滤器名称");
      return;
    }
    if (!trimmedPackage && !trimmedQuery) {
      setValidationError("至少配置一个包名或筛选规则");
      return;
    }

    setValidationError("");
    await onSubmit({
      existingID: mode === "edit" ? initialFilterID : undefined,
      name: trimmedName,
      packageName: trimmedPackage,
      query: trimmedQuery,
    });
  }

  function updateGroups(updater: (current: RuleGroup[]) => RuleGroup[]) {
    setGroups((current) => updater(current));
    setQueryOverride("");
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
      <section className="dialog-card">
        <header className="dialog-header">
          <div>
            <div className="dialog-title">{mode === "edit" ? "编辑过滤器" : "保存过滤器"}</div>
            <div className="dialog-subtitle">组内全部满足（&&），组间满足任一组（||）。</div>
          </div>
          <button className="ghost-button dialog-close" type="button" onClick={onClose}>
            取消
          </button>
        </header>

        <div className="dialog-body">
          <label className="dialog-field">
            <span>过滤器名称</span>
            <input
              className="dialog-input"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="例如：bridge h5 调试"
            />
          </label>

          <label className="dialog-field">
            <span>包名</span>
            <input
              className="dialog-input"
              list="saved-filter-package-options"
              value={packageName}
              onChange={(event) => setPackageName(event.target.value)}
              placeholder="留空表示不绑定包名"
            />
            <datalist id="saved-filter-package-options">
              {packageOptions.map((option) => <option key={option} value={option} />)}
            </datalist>
          </label>

          <div className="dialog-field">
            <div className="dialog-field-header">
              <span>筛选规则</span>
              {initialQuery.trim() ? (
                <button
                  className="ghost-button mini-button"
                  type="button"
                  onClick={() => {
                    setQueryOverride(initialQuery.trim());
                    setValidationError("");
                  }}
                >
                  重置为当前规则
                </button>
              ) : null}
            </div>

            <RuleGroupsEditor
              groups={groups}
              onAddCondition={(groupID) => {
                updateGroups((current) => current.map((group) => {
                  if (group.id !== groupID) {
                    return group;
                  }
                  return {
                    ...group,
                    conditions: [...group.conditions, createCondition()],
                  };
                }));
              }}
              onAddGroup={() => {
                updateGroups((current) => [...current, createRuleGroup()]);
              }}
              onChangeCondition={(groupID, conditionID, patch) => {
                updateGroups((current) => current.map((group) => {
                  if (group.id !== groupID) {
                    return group;
                  }
                  return {
                    ...group,
                    conditions: group.conditions.map((condition) => {
                      if (condition.id !== conditionID) {
                        return condition;
                      }
                      return patchCondition(condition, patch);
                    }),
                  };
                }));
              }}
              onDeleteCondition={(groupID, conditionID) => {
                updateGroups((current) => current.map((group) => {
                  if (group.id !== groupID || group.conditions.length === 1) {
                    return group;
                  }
                  return {
                    ...group,
                    conditions: group.conditions.filter((condition) => condition.id !== conditionID),
                  };
                }));
              }}
              onDeleteGroup={(groupID) => {
                updateGroups((current) => current.filter((group) => group.id !== groupID));
              }}
            />
          </div>

          <label className="dialog-field">
            <span>规则预览 / 手工编辑</span>
            <textarea
              className="dialog-preview"
              value={previewQuery}
              onChange={(event) => setQueryOverride(event.target.value)}
              placeholder="规则会自动生成到这里，也可以手工补充或覆盖。"
            />
          </label>

          {validationError || errorMessage ? (
            <div className="dialog-error">{validationError || errorMessage}</div>
          ) : null}
        </div>

        <footer className="dialog-footer">
          <button className="text-button secondary" type="button" onClick={onClose} disabled={saving}>
            关闭
          </button>
          <button className="text-button primary" type="button" onClick={() => void handleSubmit()} disabled={saving}>
            {saving ? "保存中…" : mode === "edit" ? "更新过滤器" : "写入并保存"}
          </button>
        </footer>
      </section>
    </div>
  );
}

function patchCondition(condition: RuleCondition, patch: Partial<RuleCondition>) {
  return {
    ...condition,
    ...patch,
  };
}
