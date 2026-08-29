import { Alert, Button } from "@persona-apps/ui";

import { publicErrorPresentation } from "./presentation";

export function CatalogueError({
    error,
    retry,
}: {
    error: unknown;
    retry: () => void;
}) {
    const state = publicErrorPresentation(error);
    return (
        <Alert variant="destructive" role="alert">
            <p>{state.message}</p>
            {state.requestId ? (
                <p className="mt-2 text-sm">
                    Request ID: <code>{state.requestId}</code>
                </p>
            ) : null}
            {state.retryable ? (
                <Button variant="outline" className="mt-3" onClick={retry}>
                    Try again
                </Button>
            ) : null}
        </Alert>
    );
}
