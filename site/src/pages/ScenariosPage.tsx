import { useEffect } from 'react';
import { SCENARIOS_DATA, SCENARIO_SLUGS, ScenarioItem, scenarioBySlug } from '../data/scenarios-data';
import { CodeBlock } from '../components/docs/CodeBlock';
import { localizePath, scenarioSlugFromPath, useLocation } from '../lib/router';
import styles from './ScenariosPage.module.css';
import '../lib/detail-meta';
import { applyRouteMeta } from '../lib/route-meta';

export interface ScenariosPageProps {
  lang?: 'pt' | 'en';
}

/**
 * Scenario catalog. /scenarios shows the first scenario; /scenarios/<slug>
 * is an indexable page per anomaly, with every section rendered (no tabs) so
 * search engines and readers see the whole write-up.
 */
export function ScenariosPage({ lang = 'en' }: ScenariosPageProps) {
  const { pathname } = useLocation();
  // App applies route metadata before this lazy chunk (and its scenario
  // titles) has loaded; apply it again once the page is here.
  useEffect(() => applyRouteMeta('scenarios', pathname), [pathname]);
  const fromUrl = scenarioBySlug(scenarioSlugFromPath(pathname));
  const currentScenario: ScenarioItem = fromUrl ?? SCENARIOS_DATA[0];
  const isDetail = fromUrl !== undefined;
  const pt = lang === 'pt';
  const name = currentScenario.name[lang] || currentScenario.name.pt;
  // Section headings sit one level below the page's scenario title.
  const SectionHeading = isDetail ? 'h2' : 'h3';

  return (
    <div className={styles.pageContainer} data-surface="light">
      <div className={styles.header}>
        <p className="technical-label" style={{ color: 'var(--purple)' }}>
          {lang === 'pt' ? 'Catálogo de Concorrência' : 'Concurrency Catalog'}
        </p>
        {isDetail ? (
          <p style={{ fontSize: 'var(--type-h3)', fontWeight: 500, letterSpacing: '-0.04em', color: 'var(--ink)' }}>
            {pt ? 'Cenários Canônicos & Mitigações' : 'Canonical Scenarios & Fixes'}
          </p>
        ) : (
          <h1 style={{ fontSize: 'var(--type-h2)', fontWeight: 500, letterSpacing: '-0.05em' }}>
            {pt ? 'Cenários Canônicos & Mitigações' : 'Canonical Scenarios & Fixes'}
          </h1>
        )}
        <p style={{ color: 'var(--text-secondary)', fontSize: 'var(--type-body-lg)', maxWidth: '42rem' }}>
          {lang === 'pt'
            ? 'Explore 10 cenários de falhas de concorrência baseados em pesquisas acadêmicas de isolamento e casos reais de sistemas financeiros e distribuídos.'
            : 'Inspect 10 concurrency failure scenarios based on academic isolation literature and real-world financial/distributed systems.'}
        </p>
      </div>

      <div className={styles.layout}>
        {/* Navegação de Cenários à Esquerda */}
        <aside className={styles.scenariosNav} aria-label={lang === 'pt' ? 'Lista de cenários' : 'Scenario list'}>
          {SCENARIOS_DATA.map((sc, idx) => {
            const isActive = sc.id === currentScenario.id;
            return (
              <a
                key={sc.id}
                href={localizePath(`/scenarios/${SCENARIO_SLUGS[sc.id]}`, lang)}
                className={`${styles.scenarioBtn} ${isActive ? styles.scenarioBtnActive : ''}`}
                aria-current={isActive ? 'page' : undefined}
              >
                <span>
                  <span className="technical-label" style={{ fontSize: '0.68rem', marginRight: '0.5rem', color: 'var(--text-secondary)' }}>
                    0{idx + 1}
                  </span>
                  {sc.name[lang] || sc.name.pt}
                </span>
                <span className={styles.anomalyBadge}>{sc.code}</span>
              </a>
            );
          })}
        </aside>

        {/* Palco do Cenário Selecionado */}
        <div className={styles.stagePane}>
          <div>
            <span className={styles.anomalyBadge}>{currentScenario.code}</span>
            {isDetail ? <h1 className={styles.stageTitle}>{name}</h1> : <h2 className={styles.stageTitle}>{name}</h2>}
            <p className={styles.summary}>{currentScenario.summary[lang] || currentScenario.summary.pt}</p>
            <p className={styles.summary}>{currentScenario.description[lang] || currentScenario.description.pt}</p>
          </div>

          <section className={styles.section} aria-labelledby="scenario-schema">
            <SectionHeading id="scenario-schema" className={styles.sectionTitle}>{pt ? 'Schema e dados iniciais' : 'Schema and seed data'}</SectionHeading>
            <CodeBlock lang={lang} code={currentScenario.schema} language="sql" filename="schema.sql & seed.sql" />
          </section>

          <section className={styles.section} aria-labelledby="scenario-workload">
            <SectionHeading id="scenario-workload" className={styles.sectionTitle}>{pt ? 'Carga concorrente (chaos.yaml)' : 'Concurrent workload (chaos.yaml)'}</SectionHeading>
            <CodeBlock lang={lang} code={currentScenario.chaos} language="yaml" filename="chaos.yaml" />
          </section>

          <section className={styles.section} aria-labelledby="scenario-analysis">
            <SectionHeading id="scenario-analysis" className={styles.sectionTitle}>{pt ? 'Por que a invariante quebra' : 'Why the invariant breaks'}</SectionHeading>
            <p className={styles.summary}>{currentScenario.analysis[lang]}</p>
            <div className={styles.infoBox} style={{ borderLeftColor: 'var(--yellow)' }}>
              <code style={{ display: 'block', fontSize: '0.95rem', color: 'var(--purple)' }}>{currentScenario.reduction.cycle}</code>
              <p style={{ margin: '0.4rem 0 0', fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
                {pt
                  ? `Delta debugging reduz ${currentScenario.reduction.originalOps} operações às ${currentScenario.reduction.minimalOps} que causam a falha.`
                  : `Delta debugging reduces ${currentScenario.reduction.originalOps} operations to the ${currentScenario.reduction.minimalOps} that cause the failure.`}
              </p>
            </div>
          </section>

          <section className={styles.section} aria-labelledby="scenario-fix">
            <SectionHeading id="scenario-fix" className={styles.sectionTitle}>{pt ? 'Correção em produção' : 'Production fix'}</SectionHeading>
            <div className={styles.infoBox} style={{ borderLeftColor: 'var(--green)' }}>
              <p style={{ margin: 0, fontWeight: 600 }}>{currentScenario.fix[lang].title}</p>
              <p style={{ margin: '0.4rem 0 0' }}>{currentScenario.fix[lang].explanation}</p>
              <div className={styles.enginesTagList}>
                {currentScenario.fix.engines.map((eng) => (
                  <span key={eng} className={styles.engineTag}>
                    ✓ {eng}
                  </span>
                ))}
              </div>
              <p style={{ margin: '0.6rem 0 0', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>{currentScenario.fix[lang].driverNotes}</p>
            </div>
            <CodeBlock lang={lang} code={currentScenario.fix[lang].code} language="sql" filename="remediation_fix.sql" />
          </section>
        </div>
      </div>
    </div>
  );
}
