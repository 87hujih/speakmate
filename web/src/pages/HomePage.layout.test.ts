import { describe, expect, it } from "vitest";
import scenarioCardSource from "../components/scenario/ScenarioCard.tsx?raw";
import homePageSource from "./HomePage.tsx?raw";

function firstClassName(source: string, pattern: RegExp) {
  const match = source.match(pattern);
  return match?.[1] ?? "";
}

describe("home page presentation", () => {
  it("pins scenario start actions to the bottom of equal-height cards", () => {
    const articleClass = firstClassName(scenarioCardSource, /<article className="([^"]+)">/);
    const goalsClass = firstClassName(scenarioCardSource, /<ul className="([^"]+)">/);
    const startButtonClass = firstClassName(scenarioCardSource, /<Button className="([^"]+)"/);
    const startLinkClass = firstClassName(scenarioCardSource, /<ButtonLink to=\{`\/training\/\$\{scenario\.sessionId\}`\} className="([^"]+)"/);

    expect(articleClass).toContain("flex");
    expect(articleClass).toContain("h-full");
    expect(articleClass).toContain("flex-col");
    expect(goalsClass).toContain("flex-1");
    expect(startButtonClass).toContain("mt-auto");
    expect(startLinkClass).toContain("mt-auto");
  });

  it("uses product copy instead of testing or implementation copy", () => {
    const forbiddenTerms = ["REST / SSE Ready", "Gin API", "Mock Agent", "联调", "测试"];

    for (const term of forbiddenTerms) {
      expect(homePageSource).not.toContain(term);
    }

    expect(homePageSource).toContain("SpeakMate 场景化英语训练");
    expect(homePageSource).toContain("训练记录闭环");
  });
});
