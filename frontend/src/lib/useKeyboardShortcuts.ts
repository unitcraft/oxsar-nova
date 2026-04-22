import { useEffect } from 'react';

interface ShortcutHandler {
  key: string;
  ctrl?: boolean;
  alt?: boolean;
  shift?: boolean;
  handler: () => void;
  description?: string;
}

export function useKeyboardShortcuts(shortcuts: ShortcutHandler[]) {
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      for (const shortcut of shortcuts) {
        const keyMatches = event.key.toLowerCase() === shortcut.key.toLowerCase() ||
                          event.code.toLowerCase() === shortcut.key.toLowerCase();
        const ctrlMatches = (shortcut.ctrl ?? false) === (event.ctrlKey || event.metaKey);
        const altMatches = (shortcut.alt ?? false) === event.altKey;
        const shiftMatches = (shortcut.shift ?? false) === event.shiftKey;

        if (keyMatches && ctrlMatches && altMatches && shiftMatches) {
          event.preventDefault();
          shortcut.handler();
          break;
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [shortcuts]);
}

export function useEscapeKey(handler: () => void) {
  useKeyboardShortcuts([{ key: 'Escape', handler }]);
}

export function useCtrlSKey(handler: () => void) {
  useKeyboardShortcuts([{ key: 's', ctrl: true, handler }]);
}
