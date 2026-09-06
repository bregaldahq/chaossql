import { ReactNode } from 'react';
import styles from './ProjectCycle.module.css';

export interface EvidenceItem {
  label: string;
  value: string;
}

export interface ProjectCycleProps {
  id: string;
  sequence: string;
  name: string;
  headline: string;
  summary: string;
  primaryAction?: {
    label: string;
    href: string;
  };
  secondaryAction?: {
    label: string;
    href: string;
  };
  technologies: readonly string[];
  evidence: readonly EvidenceItem[];
  artifact: ReactNode;
}

export function ProjectCycle({
  id,
  sequence,
  name,
  headline,
  summary,
  primaryAction,
  secondaryAction,
  technologies,
  evidence,
  artifact,
}: ProjectCycleProps) {
  return (
    <article
      id={`project-${id}`}
      className={styles.chapter}
      data-surface="light"
      aria-labelledby={`${id}-title`}
    >
      <div className={styles.chapterGrid}>
        {/* Coluna Esquerda: Narrativa */}
        <div className={styles.chapterCopy}>
          <div className={styles.projectName}>
            <span className={styles.sequence}>{sequence} /</span>
            <span className={styles.titleText}>{name}</span>
          </div>

          <h2 id={`${id}-title`} className={styles.headline}>
            {headline}
          </h2>

          <p className={styles.summary}>{summary}</p>

          {(primaryAction || secondaryAction) && (
            <div className={styles.actionsRow}>
              {primaryAction && (
                <a
                  href={primaryAction.href}
                  className={styles.primaryLink}
                >
                  {primaryAction.label}
                  <span aria-hidden="true">↗</span>
                </a>
              )}
              {secondaryAction && (
                <a
                  href={secondaryAction.href}
                  className={styles.secondaryLink}
                >
                  {secondaryAction.label}
                  <span aria-hidden="true">→</span>
                </a>
              )}
            </div>
          )}
        </div>

        {/* Coluna Direita: Artefato Sticky */}
        <div className={styles.projectDetails}>
          <div>{artifact}</div>
        </div>

        {/* Lista de Evidências */}
        <dl className={styles.evidence}>
          {evidence.map((item) => (
            <div key={item.label} className={styles.evidenceItem}>
              <dt className={styles.evidenceDt}>{item.label}</dt>
              <dd className={styles.evidenceDd}>{item.value}</dd>
            </div>
          ))}
        </dl>
      </div>

      {/* Tags de Tecnologias */}
      <div className={styles.chapterFoot}>
        <ul aria-label={`${name} tecnologias`} className={styles.techList}>
          {technologies.map((tech) => (
            <li key={tech}>{tech}</li>
          ))}
        </ul>
      </div>
    </article>
  );
}
