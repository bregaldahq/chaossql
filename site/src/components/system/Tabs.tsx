import { useId, useRef, useState, type KeyboardEvent, type ReactNode } from 'react';
import styles from './Tabs.module.css';

export interface TabItem {
  id: string;
  label: ReactNode;
  content: ReactNode;
}

export interface TabsProps {
  /** Accessible name of the tab list. */
  label: string;
  items: readonly TabItem[];
  /** Controlled selection. Omit to let the component manage it. */
  value?: string;
  defaultValue?: string;
  onChange?: (id: string) => void;
  className?: string;
}

/**
 * WAI-ARIA tabs with automatic activation: arrow keys move and select,
 * Home/End jump to the ends, and only the selected tab is in the tab order.
 */
export function Tabs({ label, items, value, defaultValue, onChange, className }: TabsProps) {
  const baseId = useId();
  const [internal, setInternal] = useState(defaultValue ?? items[0]?.id);
  const selected = value ?? internal;
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([]);

  const select = (index: number, focus: boolean) => {
    const item = items[index];
    if (!item) return;
    if (value === undefined) setInternal(item.id);
    if (item.id !== selected) onChange?.(item.id);
    if (focus) tabRefs.current[index]?.focus();
  };

  const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>, index: number) => {
    const last = items.length - 1;
    const next =
      event.key === 'ArrowRight' ? (index === last ? 0 : index + 1)
      : event.key === 'ArrowLeft' ? (index === 0 ? last : index - 1)
      : event.key === 'Home' ? 0
      : event.key === 'End' ? last
      : -1;
    if (next < 0) return;
    event.preventDefault();
    select(next, true);
  };

  const tabId = (id: string) => `${baseId}-tab-${id}`;
  const panelId = (id: string) => `${baseId}-panel-${id}`;

  return (
    <div className={[styles.root, className].filter(Boolean).join(' ')}>
      <div role="tablist" aria-label={label} className={styles.list}>
        {items.map((item, index) => {
          const isSelected = item.id === selected;
          return (
            <button
              key={item.id}
              ref={(el) => {
                tabRefs.current[index] = el;
              }}
              type="button"
              role="tab"
              id={tabId(item.id)}
              aria-selected={isSelected}
              aria-controls={panelId(item.id)}
              tabIndex={isSelected ? 0 : -1}
              className={styles.tab}
              onClick={() => select(index, false)}
              onKeyDown={(event) => onKeyDown(event, index)}
            >
              {item.label}
            </button>
          );
        })}
      </div>
      {items.map((item) =>
        item.id === selected ? (
          <div
            key={item.id}
            role="tabpanel"
            id={panelId(item.id)}
            aria-labelledby={tabId(item.id)}
            tabIndex={0}
            className={styles.panel}
          >
            {item.content}
          </div>
        ) : null
      )}
    </div>
  );
}
