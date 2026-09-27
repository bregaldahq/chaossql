// Canonical copy (English first: global audience). pt.ts must mirror these keys;
// the Messages type makes a missing or extra key a compile error.
export const en = {
  common: {
    copy: 'Copy',
    copied: 'Copied',
    close: 'Close',
  },
  nav: {
    howItWorks: 'How it works',
    homeLabel: 'ChaosSQL by Studio Bregalda, home',
    mainLabel: 'Main navigation',
    mobileLabel: 'Mobile navigation',
    languageLabel: 'Language selector',
    openMenu: 'Open navigation menu',
    closeMenu: 'Close navigation menu',
    githubRepository: 'GitHub Repository',
    items: {
      landing: 'Home',
      dashboard: 'Cloud Dashboard',
      docs: 'Docs',
      scenarios: 'Scenarios',
      visualizer: 'Trace Visualizer',
      matrix: 'Hermitage Matrix',
      playground: 'WASM Playground',
      pricing: 'Pricing',
    },
  },
  footer: {
    product: 'Product',
    tools: 'Tools',
    resources: 'Resources',
    pricing: 'Pricing',
    visualizer: 'Trace Visualizer',
    matrix: 'Hermitage Matrix',
    dashboard: 'Cloud Dashboard',
    releases: 'Releases',
    builtBy: 'Built by Studio Bregalda.',
    label: 'Footer links',
    docs: 'Documentation',
    scenarios: 'Scenarios',
    playground: 'WASM Playground',
    rights: 'All rights reserved.',
  },
  // Landing copy for the redesign (phase 4 renders it). Placeholders in {braces}
  // are filled from the recorded runs in src/data/traces (see format()).
  // Every factual claim is listed with its source in site/COPY.md.
  landingUi: {
    traceLabel: 'Recorded run of the banking example: two transactions update the same balance and one write is lost.',
    balance: 'Stored balance',
    expected: 'Expected',
    invariant: 'Invariant',
    invariantPending: 'checked when both commit',
    invariantFails: 'fails',
    pause: 'Pause the animation',
    play: 'Play the animation',
    installTitle: 'Copy the install command',
    termsTitle: 'Words used here',
    termInvariant: 'Invariant',
    termIsolation: 'Isolation level',
    termSeed: 'Seed',
    anomalyP4: 'Lost update',
    anomalyA5B: 'Write skew',
    scenarioInvariant: 'The rule that breaks',
    scenarioTrace: 'Minimal failing trace',
    scenarioSource: 'Recorded with seed {seed} from {source}',
    shrinkFromTo: '{original} to {minimal} transactions',
    shrinkLabel: 'Each square is one transaction of the failing run. Delta debugging keeps the {minimal} highlighted ones and proves nothing else is needed.',
    reportCaption: 'The same run in the HTML report, exported with --export-html.',
    ciFile: '.github/workflows/concurrency.yml',
    auditCta: 'See what the audit includes',
    waitlistTitle: 'Join the Cloud waitlist',
    waitlistLead: 'We will write when your invite is ready. No spam, no card.',
    nameLabel: 'Name',
    emailLabel: 'Work email',
    submit: 'Join the waitlist',
    submitting: 'Sending',
    requiredError: 'Please enter your name and a valid email.',
    sendError: 'We could not send this right now. Please try again.',
    successTitle: 'You are on the list',
    successBody: 'Thanks, {name}. We will write to {email} when your invite is ready.',
  },
  cta: {
    primary: 'See the bug happen',
    install: 'Install',
    installed: 'Command copied',
    waitlist: 'Join the Cloud waitlist',
    audit: 'Book an audit',
    playground: 'Open in the playground',
    allScenarios: 'See all scenarios',
    howWeMeasure: 'How we measure',
  },
  hero: {
    title: "Your tests pass. Your balances don't add up.",
    lead: 'ChaosSQL forces the races between transactions in PostgreSQL, MySQL and SQLite, then hands you the smallest test that reproduces the bug.',
  },
  proof: {
    engines: 'Works with PostgreSQL, MySQL and SQLite',
    license: 'Open source, MIT licensed',
    binary: 'One static Go binary, no CGO',
    stars: '{stars} stars on GitHub',
  },
  story: {
    title: 'A lost update in four steps',
    lead: 'This is a real run of the banking example, recorded with seed {seed}. Nothing on this page is mocked.',
    steps: {
      read: {
        title: 'Two withdrawals read the same balance',
        body: 'Both transactions read {start} before either one writes. Each is correct on its own.',
      },
      write: {
        title: 'Each one writes its own result',
        body: 'The first debits {amountA} and saves {afterA}. The second debits {amountB} and saves {afterB}.',
      },
      lost: {
        title: 'The second write erases the first',
        body: 'The account ends at {afterB} instead of {expected}. A {lost} debit vanished and no error reaches the logs.',
      },
      repro: {
        title: 'ChaosSQL proves it and hands you the test',
        body: 'The invariant fails, the engine replays the exact order, and delta debugging keeps only the {minimal} transactions that cause it, in repro_test.go.',
      },
    },
    glossary: {
      invariant: 'A rule your data must always obey, written as a SQL query. For example: the balance equals the opening balance minus every debit.',
      isolation: 'How much one transaction can see of another that is still running. Weaker levels are faster and allow more of these bugs.',
      seed: 'A number that fixes the order of every operation, so the same seed replays the same run.',
    },
  },
  scenarios: {
    title: 'It happens in more than banking',
    lead: 'Each example ships with its schema, its invariant and a recorded failing run.',
    banking: {
      name: 'Banking',
      pain: 'Two withdrawals at once and one of them disappears from the balance.',
      invariant: 'The balance always equals the opening balance minus every recorded debit.',
    },
    inventory: {
      name: 'Inventory',
      pain: 'Two buyers take the last units and the stock count only drops once.',
      invariant: 'Units in stock plus units sold always equals what you started with.',
    },
    hospital: {
      name: 'Hospital on-call',
      pain: 'Two doctors both see a colleague on call and both go home.',
      invariant: 'At least one doctor is on call at all times.',
    },
  },
  shrink: {
    title: 'From {original} transactions to the {minimal} that matter',
    lead: 'A failing run is noisy. Delta debugging removes transactions until nothing more can be removed without the bug going away, and checks that result against the database.',
    statReduction: 'smaller',
    statTrials: 'replays to prove it',
    statTime: 'to shrink and verify',
    note: 'Numbers from the recorded banking run. Our evaluation suite requires at least 85% reduction on every failing trace.',
  },
  ci: {
    title: 'Run it on every pull request',
    lead: 'The GitHub Action runs your scenarios, writes a JUnit report and a job summary, and attaches the reproduction.',
    gate: 'A violation does not fail the job by itself: gate merges on the JUnit report, or on the is-regression output from ChaosSQL Cloud, which compares each pull request against main and comments the minimal trace.',
  },
  plans: {
    title: 'Start free. Pay when it protects production.',
    oss: {
      name: 'Open source',
      price: 'Free',
      body: 'The full engine, every exporter and the GitHub Action. MIT licensed.',
    },
    cloud: {
      name: 'Cloud',
      price: 'From {price} per month',
      body: 'History, baselines against main, pull request comments and alerts. In early access.',
    },
    audit: {
      name: 'Concurrency audit',
      price: '{price}, one week',
      body: 'We map your five most critical transactions, explore more than 100,000 schedules and deliver the failing tests with fixes.',
    },
    compare: 'Compare plans',
  },
  faq: {
    title: 'Questions engineers ask first',
    items: {
      production: {
        q: 'Does it run against my production database?',
        a: 'No, and it must not. Every run resets the schema and seed data, so point it at a disposable database: a CI service container or a local copy.',
      },
      engines: {
        q: 'Which databases does it support?',
        a: 'PostgreSQL, MySQL and SQLite, plus an in-memory driver for the browser playground. SQLite serializes writes at its default level, so most anomalies show up on PostgreSQL, MySQL or weaker isolation levels.',
      },
      loadTest: {
        q: 'How is this different from a load test?',
        a: 'A load test measures throughput and hopes a race happens. ChaosSQL chooses the interleaving from a seed, checks a business rule after every run and replays the exact order that broke it.',
      },
      deterministic: {
        q: 'Is it really deterministic?',
        a: 'The same spec and seed always produce the same schedule. Databases still decide some timing on their own, and ChaosSQL reports that difference instead of hiding it.',
      },
      ci: {
        q: 'Will a violation fail my build?',
        a: 'Not by itself: the command exits successfully and records the violation in the JUnit report and job summary. Gate on those, or on the is-regression output from ChaosSQL Cloud.',
      },
      runtime: {
        q: 'What do I need to install?',
        a: 'Only the binary: go install gives you a single static Go program with no CGO. The Python and TypeScript SDKs call the same binary.',
      },
      audit: {
        q: 'What does the audit include?',
        a: 'A 45 minute kickoff, the five most critical transactions mapped, more than 100,000 schedules explored, Go reproduction tests, a final report presented to your team and three months of Cloud Team.',
      },
    },
  },
  final: {
    title: 'Find the race before your customers do.',
    lead: 'Install the CLI in one command, or book an audit before your next launch.',
  },
};

type DeepStrings<T> = { [K in keyof T]: T[K] extends string ? string : DeepStrings<T[K]> };
export type Messages = DeepStrings<typeof en>;
