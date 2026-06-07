import { describe, expect, it } from "vitest";
import source from "./ErrorAnalysisCard.tsx?raw";

describe("ErrorAnalysisCard", () => {
  it("renders the correction explanation so evidence from the report is visible", () => {
    expect(source).toContain("correction.explanation");
  });
});
