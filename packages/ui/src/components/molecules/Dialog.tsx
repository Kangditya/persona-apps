import {
    DialogTrigger,
    Button as AriaButton,
    Dialog as AriaDialog,
    Heading,
    Modal,
    ModalOverlay,
} from "react-aria-components";
import type { ReactNode } from "react";

import { cn } from "../../lib/cn";

export type DialogProps = {
    trigger: ReactNode;
    title: ReactNode;
    description?: ReactNode;
    children: ReactNode;
    closeLabel?: string;
    className?: string;
};

export function Dialog({
    trigger,
    title,
    description,
    children,
    closeLabel = "Close",
    className,
}: DialogProps) {
    return (
        <DialogTrigger>
            <AriaButton className="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50">
                {trigger}
            </AriaButton>
            <ModalOverlay
                isDismissable
                className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
            >
                <Modal className="w-full max-w-lg rounded-xl border bg-card text-card-foreground shadow-lg outline-none">
                    <AriaDialog
                        aria-label={
                            typeof title === "string" ? title : undefined
                        }
                        className={cn("grid gap-4 p-6", className)}
                    >
                        {({ close }) => (
                            <>
                                <div className="grid gap-1">
                                    <Heading
                                        slot="title"
                                        className="text-lg font-semibold"
                                    >
                                        {title}
                                    </Heading>
                                    {description ? (
                                        <p className="text-sm text-muted-foreground">
                                            {description}
                                        </p>
                                    ) : null}
                                </div>
                                <div>{children}</div>
                                <AriaButton
                                    onPress={close}
                                    className="ml-auto inline-flex h-9 items-center justify-center rounded-md border border-input px-3 text-sm font-medium hover:bg-accent focus-visible:outline-none"
                                >
                                    {closeLabel}
                                </AriaButton>
                            </>
                        )}
                    </AriaDialog>
                </Modal>
            </ModalOverlay>
        </DialogTrigger>
    );
}
