import type { ReactNode } from 'react';
import styles from './Section.module.css';

export interface SectionProps {
  id?: string;
  /** Id of the heading that names this section (for landmarks). */
  labelledBy?: string;
  width?: 'content' | 'wide';
  /** Vertical rhythm: 'default' for normal sections, 'tight' for strips. */
  spacing?: 'default' | 'tight';
  className?: string;
  children: ReactNode;
}

export function Section({ id, labelledBy, width = 'content', spacing = 'default', className, children }: SectionProps) {
  return (
    <section
      id={id}
      aria-labelledby={labelledBy}
      className={[styles.section, styles[spacing], className].filter(Boolean).join(' ')}
    >
      <div className={[styles.inner, styles[width]].join(' ')}>{children}</div>
    </section>
  );
}

export interface SectionHeaderProps {
  id?: string;
  title: ReactNode;
  lead?: ReactNode;
  align?: 'start' | 'center';
}

/** Headline stacked over an optional lead paragraph (no split headers). */
export function SectionHeader({ id, title, lead, align = 'start' }: SectionHeaderProps) {
  return (
    <header className={[styles.header, align === 'center' ? styles.center : ''].join(' ')}>
      <h2 id={id} className={styles.title}>
        {title}
      </h2>
      {lead && <p className={styles.lead}>{lead}</p>}
    </header>
  );
}
