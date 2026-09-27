import type { AnchorHTMLAttributes, ButtonHTMLAttributes, ReactNode } from 'react';
import styles from './Button.module.css';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost';
export type ButtonSize = 'md' | 'lg';

interface CommonProps {
  variant?: ButtonVariant;
  size?: ButtonSize;
  /** Icon rendered before the label. Decorative: give the label the meaning. */
  icon?: ReactNode;
  /** Icon rendered after the label (e.g. an arrow on navigation CTAs). */
  trailingIcon?: ReactNode;
  children: ReactNode;
}

type AsButton = CommonProps & ButtonHTMLAttributes<HTMLButtonElement> & { href?: undefined };
type AsLink = CommonProps & AnchorHTMLAttributes<HTMLAnchorElement> & { href: string };

export type ButtonProps = AsButton | AsLink;

export function buttonClassName(variant: ButtonVariant = 'primary', size: ButtonSize = 'md', extra?: string): string {
  return [styles.button, styles[variant], styles[size], extra].filter(Boolean).join(' ');
}

/** The one CTA primitive. Renders an <a> when given an href, a <button> otherwise. */
export function Button(props: ButtonProps) {
  const { variant = 'primary', size = 'md', icon, trailingIcon, children, className, ...rest } = props;
  const content = (
    <>
      {icon && <span className={styles.icon} aria-hidden="true">{icon}</span>}
      <span>{children}</span>
      {trailingIcon && <span className={styles.icon} aria-hidden="true">{trailingIcon}</span>}
    </>
  );

  if (typeof rest.href === 'string') {
    const anchorProps = rest as AnchorHTMLAttributes<HTMLAnchorElement>;
    return (
      <a {...anchorProps} className={buttonClassName(variant, size, className)}>
        {content}
      </a>
    );
  }

  const buttonProps = rest as ButtonHTMLAttributes<HTMLButtonElement>;
  return (
    <button type="button" {...buttonProps} className={buttonClassName(variant, size, className)}>
      {content}
    </button>
  );
}
