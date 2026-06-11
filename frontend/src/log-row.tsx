import { main } from "../wailsjs/go/models";
import { buildPlainTextTokens, TokenText, chipTone, getLogSemanticTone, timeOnly, tokenizeLogText } from "./log-text";

export function LogRow({
  log,
  onClick,
  searchQuery,
}: {
  log: LogItemView;
  onClick: () => void;
  searchQuery: string;
}) {
  const tone = getLogSemanticTone(log);
  const messageTokens = tokenizeLogText(log.message);
  return (
    <button
      className={[
        "table-row",
        `tone-${tone}`,
        log.isSelected ? "selected" : "",
        log.isCurrent ? "current" : "",
      ].join(" ")}
      onClick={onClick}
    >
      <span className="time-cell">{timeOnly(log.timeText)}</span>
      <span className={`level-chip ${chipTone(tone)}`}>{log.level}</span>
      <span className="tag-cell">
        <TokenText query={searchQuery} tokens={buildPlainTextTokens(log.tag)} />
      </span>
      <span className="message-cell">
        <TokenText query={searchQuery} tokens={messageTokens} />
      </span>
    </button>
  );
}

export type LogItemView = main.LogItemView;
