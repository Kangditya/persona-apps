import { describe, expect, it } from "vitest";

import { Badge, Dialog, DropdownMenu, Sheet } from "@persona-apps/ui";

describe("shared operations UI foundation", () => {
    it("uses shared components without application-to-application imports", () => {
        expect(Badge({ children: "Ready" }).props["data-slot"]).toBe("badge");
        expect(
            Dialog({ trigger: "Open", title: "Dialog", children: "Content" })
                .props.children,
        ).toBeTruthy();
        expect(
            DropdownMenu({ triggerLabel: "Actions", items: [] }).props.children,
        ).toBeTruthy();
        expect(
            Sheet({ trigger: "Open", title: "Panel", children: "Content" })
                .props.children,
        ).toBeTruthy();
    });
});
