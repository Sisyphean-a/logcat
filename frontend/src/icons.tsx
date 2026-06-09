import type { ReactNode } from "react";

type IconProps = {
  children: ReactNode;
};

function IconBase({ children }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" aria-hidden="true">
      {children}
    </svg>
  );
}

export function DotIcon() {
  return (
    <IconBase>
      <circle cx="8" cy="8" r="4" fill="currentColor" />
    </IconBase>
  );
}

export function DeviceIcon() {
  return (
    <IconBase>
      <rect x="3" y="2" width="10" height="12" rx="2" fill="none" stroke="currentColor" strokeWidth="1.4" />
      <circle cx="8" cy="11.5" r="0.9" fill="currentColor" />
    </IconBase>
  );
}

export function ChevronDownIcon() {
  return (
    <IconBase>
      <path d="M4 6.5 8 10l4-3.5" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
    </IconBase>
  );
}

export function CloseCircleIcon() {
  return (
    <IconBase>
      <circle cx="8" cy="8" r="5.3" fill="none" stroke="currentColor" strokeWidth="1.2" />
      <path d="M6.2 6.2 9.8 9.8M9.8 6.2 6.2 9.8" fill="none" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
    </IconBase>
  );
}

export function PauseIcon() {
  return (
    <IconBase>
      <rect x="4" y="3" width="2.5" height="10" rx="1" fill="currentColor" />
      <rect x="9.5" y="3" width="2.5" height="10" rx="1" fill="currentColor" />
    </IconBase>
  );
}

export function PlayIcon() {
  return (
    <IconBase>
      <path d="M5 3.5 12 8 5 12.5Z" fill="currentColor" />
    </IconBase>
  );
}

export function SearchIcon() {
  return (
    <IconBase>
      <circle cx="7" cy="7" r="4.2" fill="none" stroke="currentColor" strokeWidth="1.4" />
      <path d="M10.2 10.2 13 13" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
    </IconBase>
  );
}

export function SaveIcon() {
  return (
    <IconBase>
      <path d="M3 2.5h8l2 2V13.5H3Z" fill="none" stroke="currentColor" strokeWidth="1.4" />
      <path d="M5 2.5v4h5v-4" fill="none" stroke="currentColor" strokeWidth="1.4" />
      <rect x="5.2" y="9" width="5.6" height="3" rx="0.8" fill="currentColor" />
    </IconBase>
  );
}

export function ClearIcon() {
  return (
    <IconBase>
      <path d="M4 4h8" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
      <path d="M5.2 4 6 2.8h4L10.8 4" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M5 4.5 5.7 13h4.6l.7-8.5" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
    </IconBase>
  );
}

export function DownloadIcon() {
  return (
    <IconBase>
      <path d="M8 2.5v7" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
      <path d="M5.2 7.2 8 10l2.8-2.8" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M3 12.5h10" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
    </IconBase>
  );
}

export function CopyIcon() {
  return (
    <IconBase>
      <rect x="5.5" y="3.5" width="7" height="9" rx="1.4" fill="none" stroke="currentColor" strokeWidth="1.4" />
      <path d="M3.5 10.5V5a1.5 1.5 0 0 1 1.5-1.5H9" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
    </IconBase>
  );
}

export function SettingsIcon() {
  return (
    <IconBase>
      <circle cx="8" cy="8" r="2.4" fill="none" stroke="currentColor" strokeWidth="1.4" />
      <path d="M8 2.5v1.4M8 12.1v1.4M13.5 8h-1.4M3.9 8H2.5M11.9 4.1l-1 1M5.1 10.9l-1 1M11.9 11.9l-1-1M5.1 5.1l-1-1" fill="none" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
    </IconBase>
  );
}

export function DetailCollapseIcon() {
  return (
    <IconBase>
      <path d="M10 4 6 8l4 4" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
    </IconBase>
  );
}
