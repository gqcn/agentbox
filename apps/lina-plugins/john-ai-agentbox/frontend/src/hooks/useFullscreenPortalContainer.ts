import { useEffect, useState } from 'react';

export function useFullscreenPortalContainer() {
  const [container, setContainer] = useState<HTMLElement | null>(() => getFullscreenContainer());

  useEffect(() => {
    function syncContainer() {
      setContainer(getFullscreenContainer());
    }

    syncContainer();
    document.addEventListener('fullscreenchange', syncContainer);
    return () => document.removeEventListener('fullscreenchange', syncContainer);
  }, []);

  return container;
}

function getFullscreenContainer() {
  if (typeof document === 'undefined') {
    return null;
  }
  const element = document.fullscreenElement;
  return element instanceof HTMLElement ? element : null;
}
