import {
    MenuTrigger,
    Button as AriaButton,
    Menu,
    MenuItem,
    Popover,
} from "react-aria-components";

import { cn } from "../../lib/cn";

export type DropdownMenuItem = {
    id: string;
    label: string;
    disabled?: boolean;
};

export type DropdownMenuProps = {
    triggerLabel: string;
    label?: string;
    items: DropdownMenuItem[];
    onAction?: (id: string) => void;
};

export function DropdownMenu({
    triggerLabel,
    label = triggerLabel,
    items,
    onAction,
}: DropdownMenuProps) {
    return (
        <MenuTrigger>
            <AriaButton
                aria-label={triggerLabel}
                className="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent focus-visible:outline-none"
            >
                {triggerLabel}
            </AriaButton>
            <Popover className="min-w-44 rounded-md border bg-card p-1 text-card-foreground shadow-md outline-none">
                <Menu
                    aria-label={label}
                    onAction={(key) => onAction?.(String(key))}
                    className="outline-none"
                >
                    {items.map((item) => (
                        <MenuItem
                            key={item.id}
                            id={item.id}
                            isDisabled={item.disabled}
                            className={({ isFocused, isDisabled }) =>
                                cn(
                                    "cursor-default rounded-sm px-3 py-2 text-sm outline-none",
                                    isFocused && "bg-accent",
                                    isDisabled && "opacity-50",
                                )
                            }
                        >
                            {item.label}
                        </MenuItem>
                    ))}
                </Menu>
            </Popover>
        </MenuTrigger>
    );
}
