export type RuleField = "level" | "tag" | "message";
export type RuleOperator = "is" | "isNot" | "contains" | "notContains";

export type RuleCondition = {
  id: string;
  field: RuleField;
  operator: RuleOperator;
  value: string;
};

export type RuleGroup = {
  id: string;
  conditions: RuleCondition[];
};

export type SaveFilterDraft = {
  name: string;
  packageName: string;
  query: string;
};

type OperatorOption = {
  label: string;
  value: RuleOperator;
};

let builderSequence = 0;

const operatorOptionsByField: Record<RuleField, OperatorOption[]> = {
  level: [
    { value: "is", label: "等于" },
    { value: "isNot", label: "不等于" },
  ],
  tag: [
    { value: "is", label: "等于" },
    { value: "isNot", label: "不等于" },
    { value: "contains", label: "包含" },
    { value: "notContains", label: "不包含" },
  ],
  message: [
    { value: "contains", label: "包含" },
    { value: "notContains", label: "不包含" },
  ],
};

export function createRuleGroups() {
  return [createRuleGroup()];
}

export function createRuleGroup(): RuleGroup {
  return { id: nextRuleID("group"), conditions: [createCondition()] };
}

export function createCondition(field: RuleField = "tag"): RuleCondition {
  return {
    id: nextRuleID("condition"),
    field,
    operator: defaultOperator(field),
    value: field === "level" ? "I" : "",
  };
}

export function defaultOperator(field: RuleField): RuleOperator {
  return operatorOptionsByField[field][0].value;
}

export function operatorOptions(field: RuleField) {
  return operatorOptionsByField[field];
}

export function normalizeOperator(field: RuleField, operator: RuleOperator) {
  return operatorOptions(field).some((option) => option.value === operator)
    ? operator
    : defaultOperator(field);
}

export function buildFilterQuery(groups: RuleGroup[]) {
  const groupQueries = groups
    .map((group) => buildGroupQuery(group))
    .filter((query) => query.length > 0);

  if (groupQueries.length === 0) {
    return "";
  }
  if (groupQueries.length === 1) {
    return groupQueries[0];
  }
  return groupQueries.map((query) => `(${query})`).join(" || ");
}

export function buildGroupQuery(group: RuleGroup) {
  const terms = group.conditions
    .map((condition) => buildConditionQuery(condition))
    .filter((query) => query.length > 0);

  if (terms.length === 0) {
    return "";
  }
  if (terms.length === 1) {
    return terms[0];
  }
  return terms.join(" && ");
}

export function buildConditionQuery(condition: RuleCondition) {
  const value = sanitizeRuleValue(condition.value);
  if (!value) {
    return "";
  }

  if (condition.field === "level") {
    return condition.operator === "isNot" ? `-level:${value}` : `level:${value}`;
  }

  if (condition.field === "tag") {
    switch (condition.operator) {
      case "is":
        return `tag:"${value}"`;
      case "isNot":
        return `-tag:"${value}"`;
      case "notContains":
        return `-tag~:"${value}"`;
      default:
        return `tag~:"${value}"`;
    }
  }

  return condition.operator === "notContains"
    ? `-message~:"${value}"`
    : `message~:"${value}"`;
}

export function suggestFilterName(packageName: string, query: string) {
  const normalizedPackage = packageName.trim();
  if (normalizedPackage) {
    return normalizedPackage;
  }

  const normalizedQuery = query.trim();
  if (!normalizedQuery) {
    return "H5 日志";
  }
  return normalizedQuery.length > 24 ? normalizedQuery.slice(0, 24) : normalizedQuery;
}

function sanitizeRuleValue(value: string) {
  return value.trim().replaceAll("\"", "'");
}

function nextRuleID(prefix: string) {
  builderSequence += 1;
  return `${prefix}-${builderSequence}`;
}
