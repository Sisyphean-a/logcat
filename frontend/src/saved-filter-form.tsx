import { useEffect, useMemo, useRef, useState } from "react";
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

type SavedFilterFormProps = {
  draftKey: string;
  initialName: string;
  initialPackageName: string;
  initialQuery: string;
  packageOptions: string[];
  resetQueryLabel?: string;
  resetQueryValue?: string;
  onChange: (draft: SaveFilterDraft) => void;
};

export function SavedFilterForm({
  draftKey,
  initialName,
  initialPackageName,
  initialQuery,
  packageOptions,
  resetQueryLabel,
  resetQueryValue,
  onChange,
}: SavedFilterFormProps) {
  const skipSyncRef = useRef(true);
  const [name, setName] = useState(initialName);
  const [packageName, setPackageName] = useState(initialPackageName);
  const [groups, setGroups] = useState<RuleGroup[]>(createRuleGroups);
  const [queryOverride, setQueryOverride] = useState("");

  useEffect(() => {
    skipSyncRef.current = true;
    applyQueryDraft(initialQuery, setGroups, setQueryOverride);
    setName(initialName);
    setPackageName(initialPackageName);
  }, [draftKey]);

  const previewQuery = useMemo(() => {
    const generated = buildFilterQuery(groups);
    return queryOverride.trim() || generated;
  }, [groups, queryOverride]);

  useEffect(() => {
    if (skipSyncRef.current) {
      skipSyncRef.current = false;
      return;
    }
    onChange({
      name,
      packageName,
      query: previewQuery.trim(),
    });
  }, [name, onChange, packageName, previewQuery]);

  function updateGroups(updater: (current: RuleGroup[]) => RuleGroup[]) {
    setGroups((current) => updater(current));
    setQueryOverride("");
  }

  return (
    <>
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
          {resetQueryLabel && resetQueryValue?.trim() ? (
            <button
              className="ghost-button mini-button"
              type="button"
              onClick={() => applyQueryDraft(resetQueryValue, setGroups, setQueryOverride)}
            >
              {resetQueryLabel}
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
    </>
  );
}

function applyQueryDraft(
  query: string,
  setGroups: (value: RuleGroup[]) => void,
  setQueryOverride: (value: string) => void,
) {
  const parsedGroups = parseFilterQuery(query);
  setGroups(parsedGroups ?? createRuleGroups());
  setQueryOverride(parsedGroups ? "" : query.trim());
}

function patchCondition(condition: RuleCondition, patch: Partial<RuleCondition>) {
  return {
    ...condition,
    ...patch,
  };
}
