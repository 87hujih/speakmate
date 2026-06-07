export function scoreTone(score: number) {
  if (score >= 85) return "优秀";
  if (score >= 75) return "良好";
  if (score >= 60) return "可提升";
  return "需加强";
}

export function getSessionIdFallback(sessionId: string | undefined) {
  return sessionId?.trim() || "interview-20260607-001";
}
