import { describe, expect, it } from "vitest";
import { extensionForAudioMimeType, selectSupportedAudioMimeType } from "./audioMime";

describe("audio mime helpers", () => {
  it("prefers Tencent-compatible recording formats before webm", () => {
    const supported = new Set(["audio/webm;codecs=opus", "audio/webm", "audio/ogg;codecs=opus"]);

    expect(selectSupportedAudioMimeType((candidate) => supported.has(candidate))).toBe("audio/ogg;codecs=opus");
  });

  it("falls back to webm only when Tencent-compatible formats are unavailable", () => {
    const supported = new Set(["audio/webm"]);

    expect(selectSupportedAudioMimeType((candidate) => supported.has(candidate))).toBe("audio/webm");
  });

  it("maps Tencent-compatible mime types to upload extensions", () => {
    expect(extensionForAudioMimeType("audio/ogg;codecs=opus")).toBe("ogg");
    expect(extensionForAudioMimeType("audio/mp4")).toBe("m4a");
    expect(extensionForAudioMimeType("audio/wav")).toBe("wav");
  });
});
