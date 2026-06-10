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

export function parseFilterQuery(query: string) {
  const normalizedQuery = query.trim();
  if (!normalizedQuery) {
    return createRuleGroups();
  }

  const groupQueries = splitFilterQuery(normalizedQuery, "||");
  const groups = groupQueries.map((groupQuery) => parseGroupQuery(groupQuery.trim()));
  if (groups.some((group) => group === null)) {
    return null;
  }
  return groups as RuleGroup[];
}

function parseGroupQuery(query: string) {
  const normalizedGroup = trimGroupWrapper(query);
  const terms = splitFilterQuery(normalizedGroup, "&&");
  if (terms.length === 0) {
    return null;
  }

  const conditions = terms.map((term) => parseConditionQuery(term.trim()));
  if (conditions.some((condition) => condition === null)) {
    return null;
  }
  return {
    id: nextRuleID("group"),
    conditions: conditions as RuleCondition[],
  };
}

function parseConditionQuery(query: string) {
  const levelMatch = query.match(/^(-)?level[:=]([A-Z])$/);
  if (levelMatch) {
    return createParsedCondition(
      "level",
      levelMatch[1] ? "isNot" : "is",
      levelMatch[2],
    );
  }

  const tagMatch = query.match(/^(-)?tag(~?)(?::|=)(?:"([^"]*)"|(.+))$/);
  if (tagMatch) {
    return createParsedCondition(
      "tag",
      parseStringOperator(tagMatch[1], tagMatch[2] === "~"),
      (tagMatch[3] ?? tagMatch[4]).trim(),
    );
  }

  const messageMatch = query.match(/^(-)?message(~?)(?::|=)(?:"([^"]*)"|(.+))$/);
  if (messageMatch) {
    return createParsedCondition(
      "message",
      messageMatch[1] ? "notContains" : "contains",
      (messageMatch[3] ?? messageMatch[4]).trim(),
    );
  }

  return null;
}

function splitFilterQuery(query: string, operator: "&&" | "||") {
  const parts: string[] = [];
  let depth = 0;
  let quoted = false;
  let cursor = 0;

  for (let index = 0; index < query.length; index += 1) {
    const char = query[index];
    if (char === "\"" && query[index - 1] !== "\\") {
      quoted = !quoted;
      continue;
    }
    if (quoted) {
      continue;
    }
    if (char === "(") {
      depth += 1;
      continue;
    }
    if (char === ")") {
      depth -= 1;
      continue;
    }
    if (depth === 0 && query.slice(index, index + operator.length) === operator) {
      parts.push(query.slice(cursor, index).trim());
      cursor = index + operator.length;
      index += operator.length - 1;
    }
  }

  parts.push(query.slice(cursor).trim());
  return parts.filter((item) => item.length > 0);
}

function trimGroupWrapper(query: string) {
  const trimmed = query.trim();
  if (!trimmed.startsWith("(") || !trimmed.endsWith(")")) {
    return trimmed;
  }
  return hasBalancedGroupWrapper(trimmed) ? trimmed.slice(1, -1).trim() : trimmed;
}

function hasBalancedGroupWrapper(query: string) {
  let depth = 0;
  let quoted = false;

  for (let index = 0; index < query.length; index += 1) {
    const char = query[index];
    if (char === "\"" && query[index - 1] !== "\\") {
      quoted = !quoted;
      continue;
    }
    if (quoted) {
      continue;
    }
    if (char === "(") {
      depth += 1;
    }
    if (char === ")") {
      depth -= 1;
    }
    if (depth === 0 && index < query.length - 1) {
      return false;
    }
  }

  return depth === 0;
}

function createParsedCondition(
  field: RuleField,
  operator: RuleOperator,
  value: string,
): RuleCondition {
  return {
    id: nextRuleID("condition"),
    field,
    operator,
    value,
  };
}

function parseStringOperator(negated: string | undefined, fuzzy: boolean): RuleOperator {
  if (!negated && !fuzzy) {
    return "is";
  }
  if (negated && !fuzzy) {
    return "isNot";
  }
  if (!negated && fuzzy) {
    return "contains";
  }
  return "notContains";
}

function nextRuleID(prefix: string) {
  builderSequence += 1;
  return `${prefix}-${builderSequence}`;
}
