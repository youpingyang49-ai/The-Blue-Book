interface FeedTabsProps {
  value: string;
  onChange: (value: string) => void;
}

const TABS = [
  { key: "recommend", label: "推荐" },
  { key: "latest", label: "最新" },
  { key: "hot", label: "热门" },
  { key: "following", label: "关注" },
];

export default function FeedTabs({ value, onChange }: FeedTabsProps) {
  return (
    <div className="tabs" role="tablist" aria-label="内容分类">
      {TABS.map((tab) => (
        <button
          key={tab.key}
          type="button"
          role="tab"
          aria-selected={value === tab.key}
          className={`tab ${value === tab.key ? "active" : ""}`}
          onClick={() => onChange(tab.key)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
