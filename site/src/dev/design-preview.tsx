// Dev-only styleguide for the lab design system: `npm run dev`, then open
// /design-preview.html. Not part of the production build (see vite.config.ts
// rollup input and .assetsignore).
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { ArrowRight, Play } from 'lucide-react';
import '@fontsource-variable/geist';
import '@fontsource-variable/jetbrains-mono';
import '../styles/tokens.css';
import '../styles/globals.css';
import { Badge } from '../components/system/Badge';
import { Button } from '../components/system/Button';
import { Section, SectionHeader } from '../components/system/Section';
import { Tabs } from '../components/system/Tabs';
import { Terminal } from '../components/system/Terminal';
import styles from './design-preview.module.css';

const swatches = [
  ['--surface-page', 'Page'],
  ['--surface-raised', 'Raised'],
  ['--surface-brand', 'Brand surface'],
  ['--text-strong', 'Text strong'],
  ['--text-body', 'Text body'],
  ['--text-muted', 'Text muted'],
  ['--brand-text', 'Brand text'],
  ['--signal', 'Signal'],
  ['--ok', 'OK'],
  ['--danger', 'Danger'],
] as const;

function Preview() {
  return (
    <main data-theme="lab" className={styles.page}>
      <Section spacing="tight">
        <div className={styles.hero}>
          <div className={styles.heroCopy}>
            <h1 className={styles.display}>Your tests pass. Your balances don't add up.</h1>
            <p className={styles.lead}>
              ChaosSQL forces the races between transactions in PostgreSQL, MySQL and SQLite, then hands you the
              smallest test that reproduces the bug.
            </p>
            <div className={styles.row}>
              <Button size="lg" href="#" icon={<Play />}>
                See the bug happen
              </Button>
              <Button size="lg" variant="secondary" trailingIcon={<ArrowRight />}>
                Install
              </Button>
            </div>
          </div>
          <Terminal
            title="shell"
            prompt
            code={'go install github.com/bregaldahq/chaossql/cmd/chaossql@latest\nchaossql run examples/banking_lost_update/chaos.yaml --seed 184729'}
          />
        </div>
      </Section>

      <Section labelledBy="tokens">
        <SectionHeader id="tokens" title="Color roles" lead="Purple is the brand. Yellow is the one warm accent and only marks where something breaks." />
        <div className={styles.swatches}>
          {swatches.map(([token, name]) => (
            <div key={token} className={styles.swatch}>
              <span className={styles.chip} style={{ background: `var(${token})` }} />
              <span className={styles.swatchName}>{name}</span>
              <code className={styles.swatchToken}>{token}</code>
            </div>
          ))}
        </div>
      </Section>

      <Section labelledBy="components">
        <SectionHeader id="components" title="Components" />
        <div className={styles.stack}>
          <div className={styles.row}>
            <Button>Primary</Button>
            <Button variant="secondary">Secondary</Button>
            <Button variant="ghost" trailingIcon={<ArrowRight />}>
              Ghost
            </Button>
            <Button disabled>Disabled</Button>
          </div>
          <div className={styles.row}>
            <Badge>neutral</Badge>
            <Badge tone="brand">brand</Badge>
            <Badge tone="signal">P4 lost update</Badge>
            <Badge tone="ok">invariant holds</Badge>
            <Badge tone="danger">40P01 deadlock</Badge>
          </div>
          <Tabs
            label="Scenarios"
            items={[
              {
                id: 'banking',
                label: 'Banking',
                content: (
                  <Terminal
                    title="invariant.sql"
                    language="sql"
                    code={'-- total money in the system never changes\nSELECT SUM(balance) = 2000 FROM accounts;'}
                  />
                ),
              },
              {
                id: 'hospital',
                label: 'Hospital',
                content: (
                  <Terminal
                    title="invariant.sql"
                    language="sql"
                    code={'-- at least one doctor is always on call\nSELECT COUNT(*) >= 1 FROM on_call WHERE active;'}
                  />
                ),
              },
              {
                id: 'booking',
                label: 'Ticket booking',
                content: (
                  <Terminal
                    title="invariant.sql"
                    language="sql"
                    code={'-- a seat is never sold twice\nSELECT COUNT(*) = COUNT(DISTINCT seat_id) FROM tickets;'}
                  />
                ),
              },
            ]}
          />
        </div>
      </Section>

      <Section labelledBy="type">
        <SectionHeader id="type" title="Type scale" />
        <div className={styles.stack}>
          {(['--fs-display', '--fs-3xl', '--fs-2xl', '--fs-xl', '--fs-lg', '--fs-base', '--fs-sm'] as const).map((size) => (
            <p key={size} className={styles.typeSample} style={{ fontSize: `var(${size})` }}>
              <code className={styles.swatchToken}>{size}</code> Two workers read the same balance.
            </p>
          ))}
        </div>
      </Section>
    </main>
  );
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Preview />
  </StrictMode>
);
