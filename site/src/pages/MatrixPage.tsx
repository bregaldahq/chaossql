import { MatrixTable } from '../components/artifacts/MatrixTable';
import styles from './MatrixPage.module.css';

export interface MatrixPageProps {
  lang?: 'pt' | 'en';
}

export function MatrixPage({ lang = 'pt' }: MatrixPageProps) {
  return (
    <div className={styles.pageContainer} data-surface="light">
      <div className={styles.header}>
        <p className="technical-label" style={{ color: 'var(--purple)' }}>
          {lang === 'pt' ? 'Validação Empírica de Isolamento' : 'Empirical Isolation Verification'}
        </p>
        <h1 style={{ fontSize: 'var(--type-h2)', fontWeight: 500, letterSpacing: '-0.05em' }}>
          {lang === 'pt' ? 'Matriz Hermitage de Isolamento SQL' : 'Hermitage SQL Isolation Matrix'}
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: 'var(--type-body-lg)', maxWidth: '42rem' }}>
          {lang === 'pt'
            ? 'Comparação empírica dos limites reais de isolamento transacional entre SQLite, PostgreSQL e MySQL através de fuzzing concorrente sistemático executado com o comando `chaossql matrix`.'
            : 'Empirical comparison of transaction isolation boundaries across SQLite, PostgreSQL, and MySQL through systematic concurrency fuzzing executed with `chaossql matrix`.'}
        </p>
      </div>

      <div className={styles.legendBar}>
        <div className={styles.legendItem}>
          <span style={{ width: 10, height: 10, borderRadius: '50%', background: 'var(--green)' }} />
          <span>{lang === 'pt' ? 'PREVENIDO: Isolamento Seguro' : 'PREVENTED: Safe Isolation'}</span>
        </div>
        <div className={styles.legendItem}>
          <span style={{ width: 10, height: 10, borderRadius: '50%', background: 'var(--yellow)' }} />
          <span>{lang === 'pt' ? 'PERMITIDO: Risco de Anomalia' : 'PERMITTED: Anomaly Risk'}</span>
        </div>
      </div>

      <MatrixTable lang={lang} />
    </div>
  );
}
