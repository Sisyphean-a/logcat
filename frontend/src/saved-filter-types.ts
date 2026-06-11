import { type SaveFilterDraft } from "./filter-rule-builder";

export type SavedFilterDefinitionDraft = SaveFilterDraft & {
  existingID: string;
};

export type SavedFiltersDraft = {
  activeFilterID: string;
  defaultFilterID: string;
  filters: SavedFilterDefinitionDraft[];
};
