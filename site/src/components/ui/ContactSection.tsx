import styles from './ContactSection.module.css';

export interface ContactSectionProps {
  lang?: 'pt' | 'en';
}

export function ContactSection({ lang = 'pt' }: ContactSectionProps) {
  return (
    <section id="contact" className={styles.contact} data-surface="dark">
      <p className={styles.eyebrow}>
        {lang === 'pt' ? 'Studio Bregalda · Engenharia de Software' : 'Studio Bregalda · Software Engineering'}
      </p>

      <h2 className={styles.title}>
        {lang === 'pt' ? (
          <>
            Construa sistemas.
            <br />
            Com confiança comprovada.
          </>
        ) : (
          <>
            Let’s build
            <br />
            something thoughtful.
          </>
        )}
      </h2>

      <p className={styles.copy}>
        {lang === 'pt'
          ? 'ChaosSQL foi desenvolvido para transformar bugs concorrentes imprevisíveis em evidências reproduzíveis. Explore o código ou converse sobre sistemas críticos.'
          : 'ChaosSQL was crafted to turn unpredictable concurrency bugs into reproducible evidence. Inspect the repository or discuss resilient distributed architectures.'}
      </p>

      <div className={styles.buttonRow}>
        <a
          href="https://www.linkedin.com/in/ricardomeneguzzibregalda"
          target="_blank"
          rel="noreferrer"
          className={styles.contactCta}
        >
          {lang === 'pt' ? 'Conectar no LinkedIn' : 'Connect on LinkedIn'}
          <span aria-hidden="true">↗</span>
        </a>

        <a
          href="https://github.com/bregaldahq/chaossql"
          target="_blank"
          rel="noreferrer"
          className={styles.githubCta}
        >
          GitHub Repository
          <span aria-hidden="true">↗</span>
        </a>
      </div>
    </section>
  );
}
