import { useEffect, useRef, useState } from "react";
import { registerSW } from "virtual:pwa-register";

import { PwaNotice } from "@persona-apps/ui";

export function OperationsPwaStatus() {
  const updateRef = useRef<(() => Promise<void>) | undefined>(undefined);
  const [offline, setOffline] = useState(() => !navigator.onLine);
  const [updateAvailable, setUpdateAvailable] = useState(false);
  const [offlineReady, setOfflineReady] = useState(false);

  useEffect(() => {
    const goOffline = () => setOffline(true);
    const goOnline = () => setOffline(false);
    window.addEventListener("offline", goOffline);
    window.addEventListener("online", goOnline);

    if (import.meta.env.PROD) {
      updateRef.current = registerSW({
        immediate: true,
        onNeedRefresh: () => setUpdateAvailable(true),
        onOfflineReady: () => setOfflineReady(true),
      });
    }

    return () => {
      window.removeEventListener("offline", goOffline);
      window.removeEventListener("online", goOnline);
    };
  }, []);

  return (
    <PwaNotice
      offline={offline}
      updateAvailable={updateAvailable}
      offlineReady={offlineReady}
      onUpdate={() => void updateRef.current?.()}
      onDismiss={() => {
        setUpdateAvailable(false);
        setOfflineReady(false);
      }}
      className="mx-auto mb-6 max-w-5xl"
    />
  );
}
