import { useState } from 'react';
import { Check, Copy, Cpu, GitBranch, Terminal } from 'lucide-react';
import { ProjectCycle } from '../components/ui/ProjectCycle';
import { ChaosSqlArtifact } from '../components/artifacts/ChaosSqlArtifact';
import { ContactSection } from '../components/ui/ContactSection';
import styles from './LandingPage.module.css';

export interface LandingPageProps {
  lang?: 'pt' | 'en';
}

export function LandingPage({ lang = 'pt' }: LandingPageProps) {
  const [copied, setCopied] = useState(false);

  const installCmd = 'go install github.com/bregaldahq/chaossql/cmd/chaossql@latest';

  const handleCopy = () => {
    navigator.clipboard.writeText(installCmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const cycleData = {
    id: 'chaossql',
    sequence: '01',
    name: 'ChaosSQL',
    headline:
      lang === 'pt'
        ? 'Concorrência imprevisível transformada em evidência reproduzível.'
        : 'Concurrency bugs become evidence you can inspect.',
    summary:
      lang === 'pt'
        ? 'Testes de estresse em cargas de trabalho SQL concorrentes, avaliação formal de invariantes de isolamento e algoritmo causal delta-debugging para sintetizar falhas em testes mínimos.'
        : 'Stress SQL workloads, check invariants, and shrink failing execution traces into focused reproductions. Concurrency bugs become evidence you can inspect.',
    primaryAction: {
      label: lang === 'pt' ? 'Explorar Documentação' : 'Explore Documentation',
      href: '#/docs',
    },
    secondaryAction: {
      label: lang === 'pt' ? 'Testar no Playground WASM' : 'Test in WASM Playground',
      href: '#/playground',
    },
    technologies: [
      'Go 1.25',
      'SQLite',
      'PostgreSQL',
      'MySQL',
      'Seeded scheduling',
      'Causal delta debugging',
      'Go testing SDK',
      'GitHub Actions',
    ] as const,
    evidence: [
      {
        label: lang === 'pt' ? 'Detecção' : 'Detection',
        value:
          lang === 'pt'
            ? 'Cargas de trabalho concorrentes, invariantes SQL e classificação formal de Adya (P4, A3, A5B, G1, G2).'
            : 'Concurrent workloads, SQL invariants, and formal Adya anomaly taxonomy (P4, A3, A5B, G1, G2).',
      },
      {
        label: lang === 'pt' ? 'Redução' : 'Reduction',
        value:
          lang === 'pt'
            ? 'Algoritmo causal ddmin que encolhe rastros de dezenas de operações em reproduções 1-minimal focadas.'
            : 'Causal delta debugging (ddmin) narrows failing traces into focused, minimal reproductions.',
      },
      {
        label: lang === 'pt' ? 'Entrega' : 'Delivery',
        value:
          lang === 'pt'
            ? 'SDK Go (pkg/chaostest), testes standalone reproduzíveis (repro_test.go), exportação JUnit e SARIF.'
            : 'Go testing SDK, standalone repro tests, HTML interactive reports, and JUnit/SARIF exports.',
      },
    ],
  };

  const pillars = [
    {
      id: '01',
      icon: <Terminal size={18} />,
      title: lang === 'pt' ? 'Fuzzing Determinístico & Jitter' : 'Deterministic Fuzzing & Jitter',
      desc:
        lang === 'pt'
          ? 'Escalonamento baseado em PRNG com sementes e injeção de atraso micro-temporal (jitter), permitindo reprodução idêntica de deadlocks e corridas críticas.'
          : 'Seeded PRNG scheduling with micro-temporal jitter injection, delivering bit-level deterministic replays of critical transaction races and deadlocks.',
    },
    {
      id: '02',
      icon: <GitBranch size={18} />,
      title: lang === 'pt' ? 'Grafos de Dependência de Adya' : 'Adya Direct Dependency Graphs',
      desc:
        lang === 'pt'
          ? 'Construção matemática de Directed Serialization Graphs (DSG) com arestas de leitura e escrita (wr, ww, rw) para detectar ciclos de anomalias com prova formal.'
          : 'Mathematical construction of Direct Dependency Serialization Graphs (DSG) tracking write/read dependencies to formally classify isolation anomalies.',
    },
    {
      id: '03',
      icon: <Cpu size={18} />,
      title: lang === 'pt' ? 'Causal Delta-Debugging (ddmin)' : 'Causal Delta-Debugging (ddmin)',
      desc:
        lang === 'pt'
          ? 'Elimina ruído de dezenas de queries irrelevantes, isolando estritamente as poucas operações causais necessárias para disparar a quebra da invariante.'
          : 'Eliminates interleaving noise from dozens of parallel queries, shrinking traces down to the 1-minimal subset that triggers the exact invariant failure.',
    },
  ];

  return (
    <div data-surface="light">
      {/* 1. Hero Section */}
      <section className={styles.hero}>
        <div className={styles.monogramWrapper}>
          <img
            src="/brand/bregalda_monogram.svg"
            alt="Bregalda Emblem"
            width="64"
            height="64"
          />
        </div>

        <div className={styles.heroEyebrow}>
          <span className="technical-label">Studio Bregalda</span>
          <span style={{ color: 'var(--purple)' }}>·</span>
          <span className="technical-label" style={{ color: 'var(--purple)' }}>
            Database Systems Tooling
          </span>
        </div>

        <h1 className={styles.heroTitle}>
          {lang === 'pt' ? 'Transforme o caos em teste.' : 'Turn chaos into a test.'}
        </h1>

        <p className={styles.heroLead}>
          {lang === 'pt'
            ? 'Testes de estresse em cargas concorrentes SQL, verificação matemática de invariantes e encolhimento de falhas em reproduções focadas.'
            : 'Stress SQL workloads, check invariants, and shrink failing execution traces into focused reproductions. Concurrency bugs become evidence you can inspect.'}
        </p>

        <div>
          <div className={styles.installBox}>
            <code>{installCmd}</code>
            <button
              type="button"
              className={styles.copyBtn}
              onClick={handleCopy}
              aria-label="Copiar comando de instalação"
            >
              {copied ? (
                <>
                  <Check size={12} /> Copiado!
                </>
              ) : (
                <>
                  <Copy size={12} /> Copiar
                </>
              )}
            </button>
          </div>
        </div>
      </section>

      {/* 2. Signature Chapter Grid */}
      <ProjectCycle
        id={cycleData.id}
        sequence={cycleData.sequence}
        name={cycleData.name}
        headline={cycleData.headline}
        summary={cycleData.summary}
        primaryAction={cycleData.primaryAction}
        secondaryAction={cycleData.secondaryAction}
        technologies={cycleData.technologies}
        evidence={cycleData.evidence}
        artifact={<ChaosSqlArtifact />}
      />

      {/* 3. Three Pillars Section */}
      <section className={styles.pillarsSection}>
        <div style={{ textAlign: 'center', marginBottom: 'var(--space-3)' }}>
          <p className="technical-label">Fundamentos de Engenharia</p>
          <h2
            style={{
              fontSize: 'var(--type-h3)',
              fontWeight: 500,
              letterSpacing: '-0.05em',
              marginBlock: 'var(--space-1) var(--space-2)',
            }}
          >
            {lang === 'pt' ? 'Três pilares de rigor transacional' : 'Three pillars of transactional rigor'}
          </h2>
        </div>

        <div className={styles.pillarsGrid}>
          {pillars.map((pillar) => (
            <div key={pillar.id} className={styles.pillarCard}>
              <div className={styles.pillarHeader}>
                {pillar.icon}
                <span className="technical-label">{pillar.id} /</span>
              </div>
              <h3 className={styles.pillarTitle}>{pillar.title}</h3>
              <p className={styles.pillarCopy}>{pillar.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* 4. Brand Divider */}
      <div className={styles.brandDivider}>
        <img
          src="/brand/bregalda_primary_lockup.svg"
          alt="Bregalda · Build · Learn · Ship"
          width="360"
        />
      </div>

      {/* 5. Contact Section */}
      <ContactSection lang={lang} />
    </div>
  );
}
