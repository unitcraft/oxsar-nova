import type { AnchorHTMLAttributes } from 'react';

interface Props extends AnchorHTMLAttributes<HTMLAnchorElement> {
  href: string;
}

export function Link({ href, children, ...rest }: Props) {
  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    if (
      !e.ctrlKey && !e.metaKey && !e.shiftKey &&
      !href.startsWith('http') && !href.startsWith('//')
    ) {
      e.preventDefault();
      window.history.pushState({}, '', href);
      window.dispatchEvent(new PopStateEvent('popstate'));
    }
    rest.onClick?.(e);
  };
  return <a href={href} onClick={handleClick} {...rest}>{children}</a>;
}
