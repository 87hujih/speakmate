import { describe, expect, it } from "vitest";
import historySessionCardSource from "../components/history/HistorySessionCard.tsx?raw";
import historyPageSource from "./HistoryPage.tsx?raw";

describe("history page insights layout", () => {
  it("loads insights independently from the paginated history list", () => {
    expect(historyPageSource).toContain("loadHistoryInsights");
    expect(historyPageSource).toContain("isInsightsLoading");
    expect(historyPageSource).toContain("insightsError");
  });

  it("offers a 7 and 30 day insights control", () => {
    expect(historyPageSource).toContain("insightsDays");
    expect(historyPageSource).toContain("7");
    expect(historyPageSource).toContain("30");
  });

  it("removes current-page average score wording", () => {
    expect(historyPageSource).not.toContain("当前页平均综合评分");
    expect(historyPageSource).not.toContain("averageScore = useMemo");
  });

  it("keeps insights errors retryable without hiding the history list", () => {
    expect(historyPageSource).toContain("insightsError");
    expect(historyPageSource).toContain("retryInsights");
    expect(historyPageSource).toContain("洞察加载失败");
  });

  it("adds repeat practice actions to finished sessions", () => {
    expect(historySessionCardSource).toContain("再练一次同场景");
    expect(historySessionCardSource).toContain("onRepeat");
    expect(historySessionCardSource).toContain("isRepeating");
  });
});
