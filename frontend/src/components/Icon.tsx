import type { CSSProperties, ReactNode } from "react";

export type IconName =
  | "home"
  | "compass"
  | "plus"
  | "bell"
  | "chat"
  | "user"
  | "search"
  | "heart"
  | "star"
  | "comment"
  | "share"
  | "more"
  | "back"
  | "send"
  | "image"
  | "close"
  | "check"
  | "logout"
  | "grid"
  | "trash"
  | "edit";

const paths: Record<IconName, ReactNode> = {
  home: (
    <>
      <path d="M3.5 10.4 12 3.5l8.5 6.9" />
      <path d="M5.8 9.4V19a1.5 1.5 0 0 0 1.5 1.5h9.4A1.5 1.5 0 0 0 18.2 19V9.4" />
    </>
  ),
  compass: (
    <>
      <circle cx="12" cy="12" r="8.6" />
      <path d="m14.9 9.1-1.6 4.2-4.2 1.6 1.6-4.2z" />
    </>
  ),
  plus: (
    <>
      <path d="M12 5.5v13" />
      <path d="M5.5 12h13" />
    </>
  ),
  bell: (
    <>
      <path d="M18 8.8a6 6 0 1 0-12 0c0 5.2-2 6.6-2 6.6h16s-2-1.4-2-6.6" />
      <path d="M10.3 19.2a2 2 0 0 0 3.4 0" />
    </>
  ),
  chat: (
    <>
      <path d="M20.5 11.6c0 4-3.8 7.2-8.5 7.2-1 0-2-.15-2.9-.42L4 20.5l1.3-3.6A6.9 6.9 0 0 1 3.5 11.6c0-4 3.8-7.2 8.5-7.2s8.5 3.2 8.5 7.2Z" />
    </>
  ),
  user: (
    <>
      <circle cx="12" cy="8.2" r="3.7" />
      <path d="M4.8 20.2c.7-3.5 3.7-5.6 7.2-5.6s6.5 2.1 7.2 5.6" />
    </>
  ),
  search: (
    <>
      <circle cx="11" cy="11" r="6.6" />
      <path d="m16 16 4 4" />
    </>
  ),
  heart: <path d="M12 20s-7.6-4.6-7.6-9.6A4.4 4.4 0 0 1 12 7.4a4.4 4.4 0 0 1 7.6 3C19.6 15.4 12 20 12 20Z" />,
  star: <path d="m12 3.8 2.6 5.3 5.8.85-4.2 4.1 1 5.75L12 17l-5.2 2.8 1-5.75-4.2-4.1 5.8-.85z" />,
  comment: <path d="M20.4 11.7c0 3.9-3.8 7-8.4 7-1 0-1.9-.13-2.8-.38L4.2 20.4l1.3-3.5a6.7 6.7 0 0 1-1.9-4.8c0-3.9 3.8-7 8.4-7s8.4 3.1 8.4 7Z" />,
  share: (
    <>
      <path d="M14 4.5 20 10l-6 5.5v-3.2c-5 0-7.4 1.9-8.6 5.2-.4-6.6 3-9.9 8.6-10.3Z" />
    </>
  ),
  more: (
    <>
      <circle cx="5.5" cy="12" r="1.3" fill="currentColor" stroke="none" />
      <circle cx="12" cy="12" r="1.3" fill="currentColor" stroke="none" />
      <circle cx="18.5" cy="12" r="1.3" fill="currentColor" stroke="none" />
    </>
  ),
  back: <path d="M14.5 5.5 8 12l6.5 6.5" />,
  send: <path d="M20.5 3.5 3.9 10.2l6.7 2.7 2.7 6.7z" />,
  image: (
    <>
      <rect x="3.6" y="4.6" width="16.8" height="14.8" rx="2.6" />
      <circle cx="9" cy="10" r="1.6" />
      <path d="m4.6 17.4 4.5-4.2 3.3 3 2.6-2.4 4.4 4" />
    </>
  ),
  close: (
    <>
      <path d="m6 6 12 12" />
      <path d="M18 6 6 18" />
    </>
  ),
  check: <path d="m5 12.6 4.5 4.4L19 6.8" />,
  logout: (
    <>
      <path d="M14.5 4.5h3.2a1.8 1.8 0 0 1 1.8 1.8v11.4a1.8 1.8 0 0 1-1.8 1.8h-3.2" />
      <path d="M10 8.5 6 12l4 3.5" />
      <path d="M6 12h9.5" />
    </>
  ),
  grid: (
    <>
      <rect x="4" y="4" width="7" height="7" rx="2" />
      <rect x="13" y="4" width="7" height="7" rx="2" />
      <rect x="4" y="13" width="7" height="7" rx="2" />
      <rect x="13" y="13" width="7" height="7" rx="2" />
    </>
  ),
  trash: (
    <>
      <path d="M4.8 7h14.4" />
      <path d="M9.5 7V5.2A1.2 1.2 0 0 1 10.7 4h2.6a1.2 1.2 0 0 1 1.2 1.2V7" />
      <path d="M6.6 7 7.5 19a1.4 1.4 0 0 0 1.4 1.3h6.2A1.4 1.4 0 0 0 16.5 19L17.4 7" />
    </>
  ),
  edit: (
    <>
      <path d="M4.5 19.5h4L19 9a2.1 2.1 0 0 0-3-3L5.5 16.5z" />
      <path d="M14.5 7.5 17 10" />
    </>
  ),
};

interface IconProps {
  name: IconName;
  size?: number;
  strokeWidth?: number;
  className?: string;
  style?: CSSProperties;
}

export default function Icon({ name, size = 20, strokeWidth = 1.7, className, style }: IconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={strokeWidth}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      style={style}
      aria-hidden="true"
    >
      {paths[name]}
    </svg>
  );
}
