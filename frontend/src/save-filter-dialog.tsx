import { useEffect, useMemo, useState } from "react";
import {
  buildFilterQuery,
  createCondition,
  createRuleGroup,
  createRuleGroups,
  normalizeOperator,
  operatorOptions,
  type RuleCondition,
  type RuleField,
  type RuleGroup,
  type SaveFilterDraft,
} from "./filter-rule-builder";

type SaveFilterDialogProps = {
  errorMessage: string;
  existingQuery: string;
  initialName: string;
  initialPackageName: string;
  open: boolean;
  packageOptions: string[];
  saving: boolean;
  onClose: () => void;
  onSubmit: (draft: SaveFilterDraft) => Promise<void>;
};

const levelOptions = ["V", "D", "I", "W", "E", "F"];

export function SaveFilterDialog({
  errorMessage,
  existingQuery,
  initialName,
  initialPackageName,
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
    setName(initialName);
    setPackageName(initialPackageName);
    setGroups(createRuleGroups());
    setQueryOverride("");
    setValidationError("");
  }, [initialName, initialPackageName, open]);

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
            <div className="dialog-title">保存过滤器</div>
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
              {existingQuery.trim() && (
                <button
                  className="ghost-button mini-button"
                  type="button"
                  onClick={() => {
                    setQueryOverride(existingQuery.trim());
                    setValidationError("");
                  }}
                >
                  沿用当前筛选框
                </button>
              )}
            </div>

            <div className="rule-groups">
              {groups.map((group, groupIndex) => (
                <RuleGroupCard
                  key={group.id}
                  group={group}
                  groupIndex={groupIndex}
                  removable={groups.length > 1}
                  onAddCondition={() => {
                    updateGroups((current) => current.map((item) => {
                      if (item.id !== group.id) {
                        return item;
                      }
                      return {
                        ...item,
                        conditions: [...item.conditions, createCondition()],
                      };
                    }));
                  }}
                  onChangeCondition={(conditionID, patch) => {
                    updateGroups((current) => current.map((item) => {
                      if (item.id !== group.id) {
                        return item;
                      }
                      return {
                        ...item,
                        conditions: item.conditions.map((condition) => {
                          if (condition.id !== conditionID) {
                            return condition;
                          }
                          return patchCondition(condition, patch);
                        }),
                      };
                    }));
                  }}
                  onDeleteCondition={(conditionID) => {
                    updateGroups((current) => current.map((item) => {
                      if (item.id !== group.id || item.conditions.length === 1) {
                        return item;
                      }
                      return {
                        ...item,
                        conditions: item.conditions.filter((condition) => condition.id !== conditionID),
                      };
                    }));
                  }}
                  onDeleteGroup={() => {
                    updateGroups((current) => current.filter((item) => item.id !== group.id));
                  }}
                />
              ))}
            </div>

            <button
              className="text-button secondary"
              type="button"
              onClick={() => {
                updateGroups((current) => [...current, createRuleGroup()]);
              }}
            >
              新增 OR 条件组
            </button>
          </div>

          <label className="dialog-field">
            <span>规则预览</span>
            <textarea
              className="dialog-preview"
              readOnly
              value={previewQuery}
              placeholder="规则会自动生成到这里，并在保存后写回顶部筛选输入框"
            />
          </label>

          {(validationError || errorMessage) && (
            <div className="dialog-error">{validationError || errorMessage}</div>
          )}
        </div>

        <footer className="dialog-footer">
          <button className="text-button secondary" type="button" onClick={onClose} disabled={saving}>
            关闭
          </button>
          <button className="text-button primary" type="button" onClick={() => void handleSubmit()} disabled={saving}>
            {saving ? "保存中…" : "写入并保存"}
          </button>
        </footer>
      </section>
    </div>
  );
}

type RuleGroupCardProps = {
  group: RuleGroup;
  groupIndex: number;
  removable: boolean;
  onAddCondition: () => void;
  onChangeCondition: (conditionID: string, patch: Partial<RuleCondition>) => void;
  onDeleteCondition: (conditionID: string) => void;
  onDeleteGroup: () => void;
};

function RuleGroupCard({
  group,
  groupIndex,
  removable,
  onAddCondition,
  onChangeCondition,
  onDeleteCondition,
  onDeleteGroup,
}: RuleGroupCardProps) {
  return (
    <section className="rule-group-card">
      <header className="rule-group-header">
        <div className="rule-group-title">
          {groupIndex === 0 ? "条件组 1" : `或条件组 ${groupIndex + 1}`}
        </div>
        {removable && (
          <button className="ghost-button mini-button" type="button" onClick={onDeleteGroup}>
            删除组
          </button>
        )}
      </header>

      <div className="rule-group-hint">组内条件全部满足</div>

      <div className="rule-condition-list">
        {group.conditions.map((condition, conditionIndex) => (
          <RuleConditionRow
            key={condition.id}
            condition={condition}
            removable={group.conditions.length > 1}
            showJoin={conditionIndex > 0}
            onChange={(patch) => onChangeCondition(condition.id, patch)}
            onDelete={() => onDeleteCondition(condition.id)}
          />
        ))}
      </div>

      <button className="ghost-button mini-button" type="button" onClick={onAddCondition}>
        添加 AND 条件
      </button>
    </section>
  );
}

type RuleConditionRowProps = {
  condition: RuleCondition;
  removable: boolean;
  showJoin: boolean;
  onChange: (patch: Partial<RuleCondition>) => void;
  onDelete: () => void;
};

function RuleConditionRow({
  condition,
  removable,
  showJoin,
  onChange,
  onDelete,
}: RuleConditionRowProps) {
  const operators = operatorOptions(condition.field);

  return (
    <div className="rule-condition-row">
      <span className={`rule-join ${showJoin ? "" : "muted"}`}>{showJoin ? "&&" : "起始"}</span>
      <select
        className="dialog-select"
        value={condition.field}
        onChange={(event) => {
          const field = event.target.value as RuleField;
          onChange({
            field,
            operator: defaultRuleOperator(field),
            value: field === "level" ? "I" : "",
          });
        }}
      >
        <option value="level">等级</option>
        <option value="tag">标签</option>
        <option value="message">信息</option>
      </select>

      <select
        className="dialog-select"
        value={condition.operator}
        onChange={(event) => onChange({
          operator: normalizeOperator(condition.field, event.target.value as RuleCondition["operator"]),
        })}
      >
        {operators.map((option) => (
          <option key={option.value} value={option.value}>{option.label}</option>
        ))}
      </select>

      {condition.field === "level" ? (
        <select
          className="dialog-select level-select"
          value={condition.value}
          onChange={(event) => onChange({ value: event.target.value })}
        >
          {levelOptions.map((option) => <option key={option} value={option}>{option}</option>)}
        </select>
      ) : (
        <input
          className="dialog-input rule-value-input"
          value={condition.value}
          onChange={(event) => onChange({ value: event.target.value })}
          placeholder={condition.field === "tag" ? "例如：bridge" : "例如：h5"}
        />
      )}

      {removable && (
        <button className="ghost-button mini-button" type="button" onClick={onDelete}>
          删除
        </button>
      )}
    </div>
  );
}

function patchCondition(condition: RuleCondition, patch: Partial<RuleCondition>) {
  const nextField = patch.field ?? condition.field;
  const nextOperator = patch.operator
    ? normalizeOperator(nextField, patch.operator)
    : condition.operator;

  return {
    ...condition,
    ...patch,
    field: nextField,
    operator: nextOperator,
  };
}

function defaultRuleOperator(field: RuleField) {
  return operatorOptions(field)[0].value;
}
