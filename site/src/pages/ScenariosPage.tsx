import { useState } from 'react';
import { SCENARIOS_DATA, ScenarioItem } from '../data/scenarios-data';
import { CodeBlock } from '../components/docs/CodeBlock';
import styles from './ScenariosPage.module.css';

export interface ScenariosPageProps {
  lang?: 'pt' | 'en';
}

export function ScenariosPage({ lang = 'pt' }: ScenariosPageProps) {
  const [selectedId, setSelectedId] = useState<string>(SCENARIOS_DATA[0]?.id || 'banking');
  const [activeTab, setActiveTab] = useState<'schema' | 'chaos' | 'invariant' | 'fix'>('schema');

  const currentScenario: ScenarioItem =
    SCENARIOS_DATA.find((s) => s.id === selectedId) || SCENARIOS_DATA[0];

  const tabs = [
    { id: 'schema', labelPt: 'Schema & Seed SQL', labelEn: 'Schema & Seed SQL' },
    { id: 'chaos', labelPt: 'Carga de Caos (YAML)', labelEn: 'Chaos Workload (YAML)' },
    { id: 'invariant', labelPt: 'Invariante & Teoria', labelEn: 'Invariant & Theory' },
    { id: 'fix', labelPt: 'Mitigação em Produção', labelEn: 'Production Fix' },
  ];

  return (
    <div className={styles.pageContainer} data-surface="light">
      <div className={styles.header}>
        <p className="technical-label" style={{ color: 'var(--purple)' }}>
          {lang === 'pt' ? 'Catálogo de Concorrência' : 'Concurrency Catalog'}
        </p>
        <h1 style={{ fontSize: 'var(--type-h2)', fontWeight: 500, letterSpacing: '-0.05em' }}>
          {lang === 'pt' ? 'Cenários Canônicos & Mitigações' : 'Canonical Scenarios & Fixes'}
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: 'var(--type-body-lg)', maxWidth: '42rem' }}>
          {lang === 'pt'
            ? 'Explore 10 cenários de falhas de concorrência baseados em pesquisas acadêmicas de isolamento e casos reais de sistemas financeiros e distribuídos.'
            : 'Inspect 10 concurrency failure scenarios based on academic isolation literature and real-world financial/distributed systems.'}
        </p>
      </div>

      <div className={styles.layout}>
        {/* Navegação de Cenários à Esquerda */}
        <aside className={styles.scenariosNav} aria-label="Lista de cenários">
          {SCENARIOS_DATA.map((sc, idx) => {
            const isActive = sc.id === selectedId;
            return (
              <button
                key={sc.id}
                type="button"
                className={`${styles.scenarioBtn} ${isActive ? styles.scenarioBtnActive : ''}`}
                onClick={() => setSelectedId(sc.id)}
              >
                <span>
                  <span className="technical-label" style={{ fontSize: '0.68rem', marginRight: '0.5rem', color: 'var(--text-secondary)' }}>
                    0{idx + 1}
                  </span>
                  {sc.name[lang] || sc.name.pt}
                </span>
                <span className={styles.anomalyBadge}>{sc.code}</span>
              </button>
            );
          })}
        </aside>

        {/* Palco do Cenário Selecionado */}
        <main className={styles.stagePane}>
          <div>
            <span className={styles.anomalyBadge}>Adya {currentScenario.code}</span>
            <h2 className={styles.stageTitle}>
              {currentScenario.name[lang] || currentScenario.name.pt}
            </h2>
            <p className={styles.summary}>
              {currentScenario.summary[lang] || currentScenario.summary.pt}
            </p>
          </div>

          {/* Abas do Cenário */}
          <div className={styles.tabsBar}>
            {tabs.map((tab) => (
              <button
                key={tab.id}
                type="button"
                className={`${styles.tabBtn} ${activeTab === tab.id ? styles.tabBtnActive : ''}`}
                onClick={() => setActiveTab(tab.id as typeof activeTab)}
              >
                {lang === 'pt' ? tab.labelPt : tab.labelEn}
              </button>
            ))}
          </div>

          {/* Conteúdo da Aba Ativa */}
          {activeTab === 'schema' && (
            <div>
              <CodeBlock
                code={currentScenario.schema}
                language="sql"
                filename="schema.sql & seed.sql"
              />
            </div>
          )}

          {activeTab === 'chaos' && (
            <div>
              <CodeBlock
                code={currentScenario.chaos}
                language="yaml"
                filename="chaos.yaml"
              />
            </div>
          )}

          {activeTab === 'invariant' && (
            <div>
              <div className={styles.infoBox}>
                <p className="technical-label">Asserção de Consistência</p>
                <code style={{ fontSize: '1rem', color: 'var(--purple)', display: 'block', marginBlock: '0.4rem' }}>
                  {currentScenario.invariant?.assert || 'total_completed >= 0'}
                </code>
                <p style={{ margin: 0, fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
                  Query: <code>{currentScenario.invariant?.query || 'SELECT 1;'}</code>
                </p>
              </div>

              {currentScenario.adyaGraph?.cycleDescription && (
                <div className={styles.infoBox} style={{ borderLeftColor: 'var(--yellow)' }}>
                  <p className="technical-label" style={{ color: 'var(--yellow)' }}>
                    {lang === 'pt' ? 'Ciclo de Conflito de Adya' : 'Adya Conflict Cycle'}
                  </p>
                  <p style={{ margin: 0 }}>
                    {currentScenario.adyaGraph.cycleDescription[lang] ||
                      currentScenario.adyaGraph.cycleDescription.pt}
                  </p>
                </div>
              )}
            </div>
          )}

          {activeTab === 'fix' && (
            <div>
              <div className={styles.infoBox} style={{ borderLeftColor: 'var(--green)' }}>
                <p className="technical-label" style={{ color: 'var(--green)' }}>
                  {lang === 'pt' ? 'Recomendação de Produção' : 'Production Recommendation'}
                </p>
                <p style={{ margin: 0 }}>
                  {currentScenario.fix?.recommendation[lang] ||
                    currentScenario.fix?.recommendation.pt ||
                    'Elevar o nível de isolamento para SERIALIZABLE ou utilizar locks pessimistas.'}
                </p>
                {currentScenario.fix?.validatedEngines && (
                  <div className={styles.enginesTagList}>
                    {currentScenario.fix.validatedEngines.map((eng) => (
                      <span key={eng} className={styles.engineTag}>
                        ✓ {eng}
                      </span>
                    ))}
                  </div>
                )}
              </div>

              {currentScenario.fix?.sql && (
                <CodeBlock
                  code={currentScenario.fix.sql}
                  language="sql"
                  filename="remediation_fix.sql"
                />
              )}
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
