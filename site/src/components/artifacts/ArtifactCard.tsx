import { ReactNode } from 'react';
import styles from './ArtifactCard.module.css';

export interface ArtifactCardProps {
  title: string;
  icon: ReactNode;
  tag?: string;
  disclaimer?: string;
  children: ReactNode;
}

export function ArtifactCard({
  title,
  icon,
  tag = 'chaossql_demo',
  disclaimer = 'Synthetic demonstration data · architecture grounded in repository',
  children,
}: ArtifactCardProps) {
  return (
    <div
      role="group"
      aria-label={`${title} demonstration`}
      className={styles.surface}
      data-surface="dark"
    >
      <div className={styles.toolbar}>
        <span className={styles.toolTitle}>
          {icon}
          {title}
        </span>
        <span className={styles.toolbarMeta}>
          <span className={styles.exampleLabel}>Illustrative example</span>
          <span className={styles.mono}>{tag}</span>
        </span>
      </div>

      <div className={styles.appBody}>{children}</div>

      <p className={styles.disclaimer}>{disclaimer}</p>
    </div>
  );
}
