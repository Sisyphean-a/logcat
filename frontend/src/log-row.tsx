import { memo, type MouseEvent } from "react";
import { main } from "../wailsjs/go/models";
import { buildPlainTextTokens, TokenText, chipTone, getLogSemanticTone, timeOnly, tokenizeLogText } from "./log-text";
import { type LogSelectionMode, type ResultSearchPreview } from "./use-app-controller";

type LogRowProps = {
  log: LogItemView;
  index: number;
  onSelect: (index: number, mode: LogSelectionMode) => void;
  onContextMenu: (event: MouseEvent<HTMLButtonElement>, log: LogItemView) => void;
  resultSearch: ResultSearchPreview;
};

function LogRowComponent({ log, index, onSelect, onContextMenu, resultSearch }: LogRowProps) {
  const isCurrent = resultSearch.query.trim().length > 0 && log.isFocused;
  const tone = getLogSemanticTone(log);
  const messageTokens = tokenizeLogText(log.message);
  return (
    <button
      className={[
        "table-row",
        `tone-${tone}`,
        log.isSelected ? "selected" : "",
        log.isFocused ? "focused" : "",
        isCurrent ? "current" : "",
      ].join(" ")}
      onClick={(event) => onSelect(index, resolveSelectionMode(event))}
      onContextMenu={(event) => onContextMenu(event, log)}
    >
      <span className="time-cell">{timeOnly(log.timeText)}</span>
      <span className={`level-chip ${chipTone(tone)}`}>{log.level}</span>
      <span className="tag-cell">
        <TokenText highlightTerms={resultSearch.highlightTerms} tokens={buildPlainTextTokens(log.tag)} />
      </span>
      <span className="message-cell">
        <TokenText highlightTerms={resultSearch.highlightTerms} tokens={messageTokens} />
      </span>
    </button>
  );
}

function areEqual(prev: LogRowProps, next: LogRowProps) {
  return (
    prev.index === next.index &&
    prev.onSelect === next.onSelect &&
    prev.onContextMenu === next.onContextMenu &&
    prev.resultSearch.query === next.resultSearch.query &&
    prev.log.sourceIndex === next.log.sourceIndex &&
    prev.log.message === next.log.message &&
    prev.log.tag === next.log.tag &&
    prev.log.level === next.log.level &&
    prev.log.timeText === next.log.timeText &&
    prev.log.isFocused === next.log.isFocused &&
    prev.log.isSelected === next.log.isSelected
  );
}

function resolveSelectionMode(event: MouseEvent<HTMLButtonElement>): LogSelectionMode {
  if (event.shiftKey) {
    return "range";
  }
  if (event.ctrlKey || event.metaKey) {
    return "add";
  }
  return "replace";
}

export const LogRow = memo(LogRowComponent, areEqual);

export type LogItemView = main.LogItemView;
