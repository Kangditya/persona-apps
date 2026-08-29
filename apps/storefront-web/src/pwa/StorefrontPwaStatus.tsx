"use client";

import { useEffect, useRef, useState } from "react";

import { PwaNotice } from "@persona-apps/ui";

export function StorefrontPwaStatus() {
    const updateRef = useRef<(() => Promise<void>) | undefined>(undefined);
    const reloadForUpdate = useRef(false);
    const [offline, setOffline] = useState(false);
    const [updateAvailable, setUpdateAvailable] = useState(false);
    const [offlineReady, setOfflineReady] = useState(false);

    useEffect(() => {
        setOffline(!navigator.onLine);
        const goOffline = () => setOffline(true);
        const goOnline = () => setOffline(false);
        window.addEventListener("offline", goOffline);
        window.addEventListener("online", goOnline);

        let registration: ServiceWorkerRegistration | undefined;
        let installing: ServiceWorker | null = null;

        const installed = () => {
            if (installing?.state !== "installed") return;
            if (navigator.serviceWorker.controller) {
                setUpdateAvailable(true);
            } else {
                setOfflineReady(true);
            }
        };
        const watchInstalling = () => {
            installing?.removeEventListener("statechange", installed);
            installing = registration?.installing ?? null;
            installing?.addEventListener("statechange", installed);
        };
        const controllerChanged = () => {
            if (reloadForUpdate.current) window.location.reload();
        };

        if (
            process.env.NODE_ENV === "production" &&
            "serviceWorker" in navigator
        ) {
            navigator.serviceWorker.addEventListener(
                "controllerchange",
                controllerChanged,
            );
            void navigator.serviceWorker
                .register("/sw.js", { scope: "/", updateViaCache: "none" })
                .then((nextRegistration) => {
                    registration = nextRegistration;
                    if (registration.waiting) setUpdateAvailable(true);
                    if (
                        registration.active &&
                        !navigator.serviceWorker.controller
                    ) {
                        setOfflineReady(true);
                    }
                    registration.addEventListener(
                        "updatefound",
                        watchInstalling,
                    );
                    watchInstalling();
                    updateRef.current = async () => {
                        if (!registration?.waiting) return;
                        reloadForUpdate.current = true;
                        registration.waiting.postMessage({
                            type: "SKIP_WAITING",
                        });
                    };
                })
                .catch(() => undefined);
        }

        return () => {
            updateRef.current = undefined;
            installing?.removeEventListener("statechange", installed);
            registration?.removeEventListener("updatefound", watchInstalling);
            navigator.serviceWorker?.removeEventListener(
                "controllerchange",
                controllerChanged,
            );
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
