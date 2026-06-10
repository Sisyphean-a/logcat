import {
  normalizeOperator,
  operatorOptions,
  type RuleCondition,
  type RuleField,
  type RuleGroup,
} from "./filter-rule-builder";

const levelOptions = ["V", "D", "I", "W", "E", "F"];

type RuleGroupsEditorProps = {
  groups: RuleGroup[];
  onAddCondition: (groupID: string) => void;
  onAddGroup: () => void;
  onChangeCondition: (groupID: string, conditionID: string, patch: Partial<RuleCondition>) => void;
  onDeleteCondition: (groupID: string, conditionID: string) => void;
  onDeleteGroup: (groupID: string) => void;
};

export function RuleGroupsEditor({
  groups,
  onAddCondition,
  onAddGroup,
  onChangeCondition,
  onDeleteCondition,
  onDeleteGroup,
}: RuleGroupsEditorProps) {
  return (
    <div className="rule-groups">
      {groups.map((group, groupIndex) => (
        <RuleGroupCard
          key={group.id}
          group={group}
          groupIndex={groupIndex}
          removable={groups.length > 1}
          onAddCondition={() => onAddCondition(group.id)}
          onChangeCondition={(conditionID, patch) => onChangeCondition(group.id, conditionID, patch)}
          onDeleteCondition={(conditionID) => onDeleteCondition(group.id, conditionID)}
          onDeleteGroup={() => onDeleteGroup(group.id)}
        />
      ))}
      <button className="text-button secondary" type="button" onClick={onAddGroup}>
        新增 OR 条件组
      </button>
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
        {removable ? (
          <button className="ghost-button mini-button" type="button" onClick={onDeleteGroup}>
            删除组
          </button>
        ) : null}
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

      {removable ? (
        <button className="ghost-button mini-button" type="button" onClick={onDelete}>
          删除
        </button>
      ) : null}
    </div>
  );
}

function defaultRuleOperator(field: RuleField) {
  return operatorOptions(field)[0].value;
}
