import { useCallback, useEffect, useRef, useState } from "react";
import PostCard from "./PostCard";
import type { PostItem } from "./PostCard";

interface MasonryFeedProps {
  items: PostItem[];
  gap?: number;
}

function columnsFor(width: number): number {
  if (width >= 1300) return 5;
  if (width >= 1000) return 4;
  if (width >= 720) return 3;
  return 2;
}

/**
 * Lightweight masonry layout.
 * - real DOM order (row-major reading order)
 * - shortest-column placement
 * - re-measures when images finish loading or the container resizes
 */
export default function MasonryFeed({ items, gap = 16 }: MasonryFeedProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const itemRefs = useRef<Array<HTMLDivElement | null>>([]);
  const [width, setWidth] = useState(0);
  const [heights, setHeights] = useState<number[]>([]);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const update = () => setWidth(el.clientWidth);
    update();
    if (typeof ResizeObserver === "undefined") {
      window.addEventListener("resize", update);
      return () => window.removeEventListener("resize", update);
    }
    const observer = new ResizeObserver(update);
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  const measure = useCallback(() => {
    const next = items.map((_, index) => itemRefs.current[index]?.offsetHeight ?? 0);
    setHeights((prev) => {
      if (prev.length === next.length && prev.every((value, index) => Math.abs(value - next[index]) < 1)) {
        return prev;
      }
      return next;
    });
  }, [items]);

  useEffect(() => {
    const raf = requestAnimationFrame(measure);
    const timer = window.setTimeout(measure, 320);
    return () => {
      cancelAnimationFrame(raf);
      window.clearTimeout(timer);
    };
  }, [measure, width]);

  const columns = columnsFor(width);
  const columnWidth = width > 0 ? (width - gap * (columns - 1)) / columns : 0;

  const positions: Array<{ x: number; y: number }> = [];
  const columnHeights = new Array(columns).fill(0);
  let totalHeight = 0;

  if (columnWidth > 0) {
    items.forEach((_, index) => {
      let target = 0;
      for (let c = 1; c < columns; c += 1) {
        if (columnHeights[c] < columnHeights[target] - 0.5) target = c;
      }
      positions.push({ x: target * (columnWidth + gap), y: columnHeights[target] });
      columnHeights[target] += (heights[index] || 0) + gap;
    });
    totalHeight = Math.max(0, ...columnHeights) - gap;
  }

  const ready = columnWidth > 0 && heights.length === items.length;

  return (
    <div className="masonry" ref={containerRef} style={{ height: ready ? totalHeight : undefined }}>
      {items.map((post, index) => {
        const pos = positions[index];
        return (
          <div
            key={post.id}
            ref={(node) => {
              itemRefs.current[index] = node;
            }}
            className="masonry-item"
            style={{
              width: columnWidth || "100%",
              transform: pos ? `translate3d(${pos.x}px, ${pos.y}px, 0)` : undefined,
              opacity: pos ? 1 : 0,
              visibility: pos ? "visible" : "hidden",
            }}
          >
            <PostCard post={post} onMediaLoad={measure} />
          </div>
        );
      })}
    </div>
  );
}
