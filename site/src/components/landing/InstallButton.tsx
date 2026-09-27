import { useEffect, useRef, useState } from 'react';
import { Check, Copy } from 'lucide-react';
import { track } from '../../lib/analytics';
import { messages, type Language } from '../../i18n';
import { Button } from '../system/Button';

export const INSTALL_CMD = 'go install github.com/bregaldahq/chaossql/cmd/chaossql@latest';

/** Copies the install command; `placement` names the spot for analytics. */
export function InstallButton({ lang, placement }: { lang: Language; placement: string }) {
  const m = messages[lang];
  const [copied, setCopied] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  useEffect(() => () => clearTimeout(timer.current), []);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(INSTALL_CMD);
    } catch {
      return;
    }
    track('install_copy', placement);
    setCopied(true);
    clearTimeout(timer.current);
    timer.current = setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Button
      variant="secondary"
      size="lg"
      onClick={copy}
      title={INSTALL_CMD}
      aria-label={`${copied ? m.cta.installed : m.cta.install}. ${m.landingUi.installTitle}: ${INSTALL_CMD}`}
      icon={copied ? <Check /> : <Copy />}
    >
      <span aria-live="polite">{copied ? m.cta.installed : m.cta.install}</span>
    </Button>
  );
}
