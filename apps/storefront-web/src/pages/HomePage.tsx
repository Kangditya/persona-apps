import { useStorefrontHealth } from "../api/diagnostics";
import { useState } from "react";
import {
    Alert,
    Badge,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Dialog,
    DropdownMenu,
    Field,
    Input,
    Separator,
    Sheet,
    Skeleton,
    Textarea,
} from "@persona-apps/ui";

export function HomePage() {
    const health = useStorefrontHealth();
    const [selectedAction, setSelectedAction] = useState("None");

    return (
        <section>
            <p className="font-medium text-amber-700">Qurban Event</p>
            <h1 className="mt-2 text-4xl font-semibold">Event landing</h1>
            <p className="mt-4 text-stone-600">
                Qurban event discovery and purchasing journeys are not
                implemented yet.
            </p>
            <p className="mt-4 text-sm text-stone-500" role="status">
                {health.isLoading && "Checking API availability…"}
                {health.isSuccess && `API status: ${health.data.status}.`}
                {health.isError && "API is currently unavailable."}
            </p>
            <Card className="mt-10 max-w-2xl">
                <CardHeader>
                    <div className="flex items-center gap-2">
                        <CardTitle>Shared UI foundation</CardTitle>
                        <Badge variant="secondary">Reusable</Badge>
                    </div>
                    <CardDescription>
                        A small, non-business showcase proving the shared
                        package boundary.
                    </CardDescription>
                </CardHeader>
                <CardContent className="grid gap-5">
                    <Field
                        id="showcase-note"
                        label="Accessible field"
                        description="Descriptions and validation feedback are linked automatically."
                        required
                    >
                        <Input placeholder="Type a note" />
                    </Field>
                    <Textarea
                        aria-label="Optional details"
                        placeholder="Optional details"
                    />
                    <Separator />
                    <div className="flex flex-wrap gap-2">
                        <Button onClick={() => setSelectedAction("Button")}>
                            Native button
                        </Button>
                        <Dialog
                            trigger="Open dialog"
                            title="Shared dialog"
                            description="React Aria owns focus, Escape, and naming."
                        >
                            <p className="text-sm text-muted-foreground">
                                This is presentational foundation content.
                            </p>
                        </Dialog>
                        <DropdownMenu
                            triggerLabel="Open menu"
                            items={[
                                { id: "first", label: "First action" },
                                { id: "second", label: "Second action" },
                            ]}
                            onAction={(id) => setSelectedAction(id)}
                        />
                        <Sheet trigger="Open sheet" title="Shared sheet">
                            <p className="text-sm text-muted-foreground">
                                The sheet is keyboard-dismissible and
                                focus-managed.
                            </p>
                        </Sheet>
                    </div>
                    <Alert>Last selected action: {selectedAction}</Alert>
                    <Skeleton className="h-3 w-40" />
                </CardContent>
            </Card>
        </section>
    );
}
