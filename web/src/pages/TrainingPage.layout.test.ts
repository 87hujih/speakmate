import { describe, expect, it } from "vitest";
import conversationPanelSource from "../components/training/ConversationPanel.tsx?raw";
import trainingPageSource from "./TrainingPage.tsx?raw";

function firstClassName(source: string, pattern: RegExp) {
  const match = source.match(pattern);
  return match?.[1] ?? "";
}

describe("training page layout", () => {
  it("bounds the training workspace to the viewport instead of growing with messages", () => {
    const workspaceClass = firstClassName(trainingPageSource, /<PageContainer size="full" className="([^"]+)">\s*<TrainingHeader/s);
    const gridClass = firstClassName(trainingPageSource, /<div className="([^"]+)">\s*<TaskPanel/s);

    expect(workspaceClass).toContain("min-h-[calc(100vh-72px)]");
    expect(workspaceClass).toContain("xl:h-[calc(100vh-72px)]");
    expect(workspaceClass).toContain("xl:min-h-0");
    expect(workspaceClass).toContain("xl:overflow-hidden");
    expect(workspaceClass).not.toContain("flex h-[calc(100vh-72px)]");

    expect(gridClass).toContain("min-h-0");
    expect(gridClass).toContain("xl:overflow-hidden");
  });

  it("makes the conversation message list the scroll container", () => {
    const panelClass = firstClassName(conversationPanelSource, /<section className="([^"]+)">/);
    const scrollClass = firstClassName(conversationPanelSource, /<div ref=\{scrollRef\} className="([^"]+)">/);
    const formClass = firstClassName(conversationPanelSource, /<form onSubmit=\{handleSubmit\} className="([^"]+)">/);

    expect(panelClass).toContain("h-[calc(100vh-120px)]");
    expect(panelClass).toContain("max-h-[760px]");
    expect(panelClass).toContain("min-h-[520px]");
    expect(panelClass).toContain("xl:h-full");
    expect(panelClass).toContain("xl:min-h-0");
    expect(panelClass).toContain("overflow-hidden");

    expect(scrollClass).toContain("min-h-0");
    expect(scrollClass).toContain("flex-1");
    expect(scrollClass).toContain("overflow-y-auto");

    expect(formClass).toContain("shrink-0");
  });
});
