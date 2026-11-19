import type { ReactNode } from "react";
import Icon from "./Icon";
import type { IconName } from "./Icon";

export function LoadingRow({ text = "加载中…" }: { text?: string }) {
  return (
    <div className="loading-row">
      <span className="spinner" />
      {text}
    </div>
  );
}

export function PageLoading({ text = "加载中…" }: { text?: string }) {
  return (
    <div className="page-loading">
      <span className="spinner" />
      {text}
    </div>
  );
}

interface EmptyStateProps {
  icon?: IconName;
  title: string;
  description?: string;
  action?: ReactNode;
}

export function EmptyState({ icon = "compass", title, description, action }: EmptyStateProps) {
  return (
    <div className="empty-state">
      <div className="empty-icon">
        <Icon name={icon} size={24} />
      </div>
      <h3>{title}</h3>
      {description ? <p>{description}</p> : null}
      {action}
    </div>
  );
}

export function ErrorState({ message, onRetry }: { message?: string; onRetry?: () => void }) {
  return (
    <div className="error-state">
      <div className="empty-icon" style={{ background: "#fff1f0", color: "#f0453a" }}>
        <Icon name="close" size={22} />
      </div>
      <p>{message || "加载失败了，请稍后重试"}</p>
      {onRetry ? (
        <button className="btn btn-ghost btn-sm" onClick={onRetry}>
          重新加载
        </button>
      ) : null}
    </div>
  );
}

export function SkeletonCard({ ratio = 1.33 }: { ratio?: number }) {
  return (
    <div style={{ borderRadius: 16, overflow: "hidden" }}>
      <div className="skeleton" style={{ width: "100%", aspectRatio: `1 / ${ratio}` }} />
      <div style={{ padding: "10px 2px" }}>
        <div className="skeleton" style={{ height: 13, width: "86%", marginBottom: 8 }} />
        <div className="skeleton" style={{ height: 13, width: "52%" }} />
      </div>
    </div>
  );
}

export function FeedSkeleton({ count = 8 }: { count?: number }) {
  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(auto-fill, minmax(220px, 1fr))",
        gap: 16,
      }}
    >
      {Array.from({ length: count }).map((_, i) => (
        <SkeletonCard key={i} ratio={i % 3 === 0 ? 1.33 : i % 3 === 1 ? 1.2 : 1.5} />
      ))}
    </div>
  );
}
