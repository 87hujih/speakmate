import { describe, expect, it } from "vitest";
import { createSessionStreamUrl } from "./sessionStream";

describe("session stream helpers", () => {
  it("builds stream URLs from the API base URL", () => {
    expect(createSessionStreamUrl(7, "/api/v1")).toBe("/api/v1/sessions/7/stream");
    expect(createSessionStreamUrl(7, "http://localhost:8080/api/v1/")).toBe("http://localhost:8080/api/v1/sessions/7/stream");
  });
});
