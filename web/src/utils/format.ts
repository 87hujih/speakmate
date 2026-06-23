/** scoreTone 根据分数返回对应的视觉语气。 */
export function scoreTone(score: number) {
  if (score >= 85) return "优秀";
  if (score >= 75) return "良好";
  if (score >= 60) return "可提升";
  return "需加强";
}

/** getSessionIdFallback 返回路由缺失时使用的默认 Session ID。 */
export function getSessionIdFallback(sessionId: string | undefined) {
  return sessionId?.trim() || "interview-20260607-001";
}
