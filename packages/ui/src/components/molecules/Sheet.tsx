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

export type SheetProps = {
    trigger: ReactNode;
    title: ReactNode;
    children: ReactNode;
    side?: "left" | "right";
};

export function Sheet({
    trigger,
    title,
    children,
    side = "right",
}: SheetProps) {
    return (
        <DialogTrigger>
            <AriaButton className="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent focus-visible:outline-none">
                {trigger}
            </AriaButton>
            <ModalOverlay
                isDismissable
                className="fixed inset-0 z-50 bg-black/50"
            >
                <Modal
                    className={cn(
                        "fixed inset-y-0 w-full max-w-sm bg-card p-6 text-card-foreground shadow-xl outline-none",
                        side === "right" ? "right-0" : "left-0",
                    )}
                >
                    <AriaDialog
                        aria-label={
                            typeof title === "string" ? title : undefined
                        }
                        className="grid h-full content-start gap-4"
                    >
                        {({ close }) => (
                            <>
                                <div className="flex items-center justify-between gap-4">
                                    <Heading
                                        slot="title"
                                        className="text-lg font-semibold"
                                    >
                                        {title}
                                    </Heading>
                                    <AriaButton
                                        onPress={close}
                                        aria-label="Close panel"
                                        className="inline-flex size-9 items-center justify-center rounded-md border border-input hover:bg-accent focus-visible:outline-none"
                                    >
                                        ×
                                    </AriaButton>
                                </div>
                                <div>{children}</div>
                            </>
                        )}
                    </AriaDialog>
                </Modal>
            </ModalOverlay>
        </DialogTrigger>
    );
}
