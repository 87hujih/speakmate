import { describe, expect, it } from "vitest";
import { initialVoiceInputState, voiceInputStateReducer } from "./useVoiceInput";

describe("voice input state machine", () => {
  it("clears stale transcript and error when recording starts", () => {
    const previous = {
      status: "idle" as const,
      transcript: "old transcript",
      error: "old error",
    };

    expect(voiceInputStateReducer(previous, { type: "recording_started" })).toEqual({
      status: "recording",
      transcript: "",
      error: "",
    });
  });

  it("tracks transcript progress through recognition and thinking", () => {
    const recording = voiceInputStateReducer(initialVoiceInputState, { type: "recording_started" });
    const partial = voiceInputStateReducer(recording, { type: "partial_transcript", transcript: "I would like" });
    const recognizing = voiceInputStateReducer(partial, { type: "recognizing", transcript: "I would like a refund" });
    const thinking = voiceInputStateReducer(recognizing, { type: "thinking" });

    expect(thinking).toEqual({
      status: "thinking",
      transcript: "I would like a refund",
      error: "",
    });
  });

  it("returns to idle with an error message when voice input fails", () => {
    const recording = voiceInputStateReducer(initialVoiceInputState, { type: "recording_started" });

    expect(voiceInputStateReducer(recording, { type: "error", message: "麦克风不可用" })).toEqual({
      status: "idle",
      transcript: "",
      error: "麦克风不可用",
    });
  });
});
