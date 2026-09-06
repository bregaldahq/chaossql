import scenariosJson from './scenarios.json';

export interface ScenarioItem {
  id: string;
  name: {
    pt: string;
    en: string;
  };
  code: string;
  description: {
    pt: string;
    en: string;
  };
  summary: {
    pt: string;
    en: string;
  };
  schema: string;
  chaos: string;
  invariant?: {
    name: string;
    query: string;
    assert: string;
    explanation?: {
      pt: string;
      en: string;
    };
  };
  adyaGraph?: {
    nodes: string[];
    edges: { from: string; to: string; type: string }[];
    cycleDescription?: {
      pt: string;
      en: string;
    };
  };
  fix?: {
    recommendation: {
      pt: string;
      en: string;
    };
    sql: string;
    validatedEngines: string[];
    driverNotes?: {
      pt: string;
      en: string;
    };
  };
}

export const SCENARIOS_DATA = scenariosJson as unknown as ScenarioItem[];

export interface MatrixRow {
  code: string;
  name: string;
  formalCycle: string;
  sqlite: 'permitted' | 'prevented' | 'detected';
  sqliteLabel?: string;
  postgresRc: 'permitted' | 'prevented' | 'detected';
  postgresRcLabel?: string;
  postgresSsi: 'permitted' | 'prevented' | 'detected';
  postgresSsiLabel?: string;
  mysqlRr: 'permitted' | 'prevented' | 'detected';
  mysqlRrLabel?: string;
  adyaDetails: {
    pt: string;
    en: string;
  };
  cliCommand: string;
}

export const HERMITAGE_MATRIX: MatrixRow[] = [
  {
    code: 'P4',
    name: 'Lost Update (Perda de Atualização)',
    formalCycle: 'T1 ──(rw)──► T2 ──(ww)──► T1',
    sqlite: 'permitted',
    postgresRc: 'permitted',
    postgresSsi: 'prevented',
    mysqlRr: 'prevented',
    adyaDetails: {
      pt: 'Duas transações leem o mesmo registro e calculam novas versões com base em leituras defasadas. A segunda escrita comita e sobrescreve cegamente a primeira.',
      en: 'Two concurrent transactions read the same row and overwrite mutations blindly without transactional conflict serialization.',
    },
    cliCommand: 'chaossql demo banking',
  },
  {
    code: 'A5B',
    name: 'Write Skew (Distorção de Escrita)',
    formalCycle: 'T1 ──(rw)──► T2 ──(rw)──► T1',
    sqlite: 'permitted',
    postgresRc: 'permitted',
    postgresSsi: 'prevented',
    mysqlRr: 'permitted',
    adyaDetails: {
      pt: 'Transações com conjuntos de escrita disjuntos violam uma invariante global sobre conjuntos de leitura sobrepostos (ex: médicos de plantão).',
      en: 'Disjoint write sets violate a global invariant over overlapping read sets (e.g. hospital doctor on-call rotation).',
    },
    cliCommand: 'chaossql demo hospital',
  },
  {
    code: 'A5A',
    name: 'Read Skew (Distorção de Leitura)',
    formalCycle: 'T1 ──(rw)──► T2 ──(wr)──► T1',
    sqlite: 'permitted',
    postgresRc: 'permitted',
    postgresSsi: 'prevented',
    mysqlRr: 'prevented',
    adyaDetails: {
      pt: 'Transação lê registro X, outra atualiza X e Y de forma atômica, mas a primeira transação lê o novo estado de Y inconsistente.',
      en: 'Transaction reads X, concurrent tx updates X and Y atomically, but first transaction subsequently reads inconsistent Y state.',
    },
    cliCommand: 'chaossql demo financial',
  },
  {
    code: 'G0',
    name: 'Dirty Write (Escrita Suja)',
    formalCycle: 'T1 ──(ww)──► T2 ──(ww)──► T1',
    sqlite: 'prevented',
    postgresRc: 'prevented',
    postgresSsi: 'prevented',
    mysqlRr: 'prevented',
    adyaDetails: {
      pt: 'Transação modifica dado já alterado por outra transação ainda não confirmada. Prevenido em todos os motores por 2PL estrito.',
      en: 'Transaction modifies data already updated by another uncommitted transaction. Prevented by Strict 2PL write locks.',
    },
    cliCommand: 'chaossql demo auction',
  },
  {
    code: 'G1a',
    name: 'Dirty Read / Aborted Read (Leitura Suja)',
    formalCycle: 'w1(x) ... r2(x) ... a1',
    sqlite: 'prevented',
    postgresRc: 'prevented',
    postgresSsi: 'prevented',
    mysqlRr: 'prevented',
    adyaDetails: {
      pt: 'Transação observa valores gerados por outra transação que sofreu rollback posterior.',
      en: 'Transaction reads intermediate data from another transaction that subsequently rolled back.',
    },
    cliCommand: 'chaossql demo flash_crash',
  },
  {
    code: 'G2',
    name: 'Anti-Dependency Cycle (Ciclo de Anti-Dependência)',
    formalCycle: 'T1 ──(rw)──► T2 ──(rw)──► T3 ──(rw)──► T1',
    sqlite: 'permitted',
    postgresRc: 'permitted',
    postgresSsi: 'prevented',
    mysqlRr: 'permitted',
    adyaDetails: {
      pt: 'Ciclo direcionado de anti-dependências (rw), gerando inconsistência de serializabilidade estrita.',
      en: 'Directed cycle containing anti-dependency edges (rw), violating strict conflict serializability.',
    },
    cliCommand: 'chaossql demo ticket',
  },
  {
    code: 'G-DL',
    name: 'Deadlock Cycle & Recovery (Bloqueio Mútuo)',
    formalCycle: 'T1 ──(waits)──► T2 ──(waits)──► T1',
    sqlite: 'detected',
    sqliteLabel: 'Timeout',
    postgresRc: 'detected',
    postgresRcLabel: '40P01',
    postgresSsi: 'detected',
    postgresSsiLabel: '40P01',
    mysqlRr: 'detected',
    mysqlRrLabel: '1213',
    adyaDetails: {
      pt: 'Bloqueio mútuo circular entre transações concorrentes na disputa por locks de linha ou tabela.',
      en: 'Circular mutual waiting between transactions competing for exclusive row/table locks.',
    },
    cliCommand: 'chaossql demo deadlock',
  },
];
