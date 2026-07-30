import { describe, expect, it } from "vitest";

import { paths } from "./paths";

describe("operations paths", () => {
  it("keeps the bootstrap routes stable", () => {
    expect(Object.values(paths)).toEqual(["/login", "/management", "/pos"]);
  });
});
