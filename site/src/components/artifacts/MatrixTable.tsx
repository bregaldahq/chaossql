import { useState } from 'react';
import { HERMITAGE_MATRIX, MatrixRow } from '../../data/scenarios-data';
import styles from './MatrixTable.module.css';

export interface MatrixTableProps {
  lang?: 'pt' | 'en';
}

export function MatrixTable({ lang = 'pt' }: MatrixTableProps) {
  const [selectedRow, setSelectedRow] = useState<MatrixRow>(HERMITAGE_MATRIX[0]);

  const renderBadge = (status: string, label?: string) => {
    if (status === 'prevented') {
      return <span className={styles.badgePrevented}>{lang === 'pt' ? 'PREVENIDO' : 'PREVENTED'}</span>;
    }
    if (status === 'permitted') {
      return <span className={styles.badgePermitted}>{lang === 'pt' ? 'PERMITIDO' : 'PERMITTED'}</span>;
    }
    return <span className={styles.badgeCycle}>{label || 'DETECTED'}</span>;
  };

  return (
    <div>
      <div className={styles.tableCard}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th>{lang === 'pt' ? 'Anomalia / Código' : 'Anomaly / Code'}</th>
              <th>{lang === 'pt' ? 'Grafo Formal' : 'Formal Cycle'}</th>
              <th>SQLite (WAL)</th>
              <th>Postgres (RC)</th>
              <th>Postgres (SSI)</th>
              <th>MySQL (InnoDB RR)</th>
            </tr>
          </thead>
          <tbody>
            {HERMITAGE_MATRIX.map((row) => {
              const isSelected = selectedRow.code === row.code;
              return (
                <tr
                  key={row.code}
                  className={`${styles.row} ${isSelected ? styles.rowSelected : ''}`}
                  onClick={() => setSelectedRow(row)}
                >
                  <td>
                    <strong>{row.code}</strong> — {row.name}
                  </td>
                  <td>
                    <code style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                      {row.formalCycle}
                    </code>
                  </td>
                  <td>{renderBadge(row.sqlite, row.sqliteLabel)}</td>
                  <td>{renderBadge(row.postgresRc, row.postgresRcLabel)}</td>
                  <td>{renderBadge(row.postgresSsi, row.postgresSsiLabel)}</td>
                  <td>{renderBadge(row.mysqlRr, row.mysqlRrLabel)}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {selectedRow && (
        <div className={styles.detailDrawer}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.8rem' }}>
            <span className={styles.badgeCycle}>{selectedRow.code}</span>
            <h3 style={{ margin: 0, fontSize: '1.2rem', color: 'var(--ink)' }}>
              {selectedRow.name}
            </h3>
          </div>

          <p style={{ margin: 0, fontSize: '0.92rem', lineHeight: 1.7, color: 'var(--ink)' }}>
            {selectedRow.adyaDetails[lang] || selectedRow.adyaDetails.pt}
          </p>

          <div style={{ marginTop: 'var(--space-2)' }}>
            <span className="technical-label" style={{ marginRight: '0.5rem', color: 'var(--purple)' }}>
              {lang === 'pt' ? 'Comando para reproduzir:' : 'Reproduction command:'}
            </span>
            <code
              style={{
                fontFamily: 'var(--font-jetbrains-mono), monospace',
                fontSize: '0.82rem',
                padding: '0.3rem 0.6rem',
                background: 'color-mix(in srgb, var(--purple) 8%, var(--cream))',
                border: '1px solid var(--border-subtle)',
                borderRadius: 'var(--radius-control)',
              }}
            >
              {selectedRow.cliCommand}
            </code>
          </div>
        </div>
      )}
    </div>
  );
}
