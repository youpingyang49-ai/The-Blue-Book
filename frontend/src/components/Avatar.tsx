import { useState } from "react";

const GRADIENTS = [
  "linear-gradient(135deg,#7aa7ff,#4c7df0)",
  "linear-gradient(135deg,#ffb199,#ff7a9c)",
  "linear-gradient(135deg,#8fd3c7,#4fb3a5)",
  "linear-gradient(135deg,#ffd479,#f5a623)",
  "linear-gradient(135deg,#b8a4ff,#7c5cff)",
  "linear-gradient(135deg,#8fd0ff,#4aa3df)",
  "linear-gradient(135deg,#ffc9a3,#f0864f)",
  "linear-gradient(135deg,#a8e6a1,#5fbf6d)",
];

function hash(str: string): number {
  let h = 0;
  for (let i = 0; i < str.length; i++) h = (h * 31 + str.charCodeAt(i)) >>> 0;
  return h;
}

interface AvatarProps {
  src?: string | null;
  name?: string | null;
  size?: "xs" | "sm" | "md" | "lg" | "xl";
  className?: string;
}

export default function Avatar({ src, name, size = "sm", className = "" }: AvatarProps) {
  const [failed, setFailed] = useState(false);
  const label = (name || "书").trim();
  const initial = label.charAt(0);
  const bg = GRADIENTS[hash(label) % GRADIENTS.length];
  const showImg = Boolean(src) && !failed;

  return (
    <div className={`avatar avatar-${size} ${className}`.trim()} style={{ background: bg }} title={label}>
      {showImg ? (
        <img src={src as string} alt={label} loading="lazy" onError={() => setFailed(true)} />
      ) : (
        <span aria-hidden="true">{initial}</span>
      )}
    </div>
  );
}
