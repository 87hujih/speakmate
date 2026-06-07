import { describe, expect, it, vi } from "vitest";
import {
  createRealtimeSpeechSession,
  isRealtimeSpeechSupported,
  normalizeSpeechRecognitionResult,
  type SpeechRecognitionLike,
} from "./realtimeSpeech";

function makeHost(recognition?: new () => SpeechRecognitionLike) {
  return recognition ? { SpeechRecognition: recognition } : {};
}

class FakeRecognition implements SpeechRecognitionLike {
  continuous = false;
  interimResults = false;
  lang = "";
  onresult: ((event: unknown) => void) | null = null;
  onerror: ((event: unknown) => void) | null = null;
  onend: (() => void) | null = null;
  start = vi.fn();
  stop = vi.fn();
  abort = vi.fn();
}

function resultEvent(transcripts: string[], isFinal: boolean) {
  return {
    results: transcripts.map((transcript) => ({
      isFinal,
      0: { transcript },
    })),
  };
}

describe("realtime speech helper", () => {
  it("detects browser speech recognition support", () => {
    expect(isRealtimeSpeechSupported(makeHost(FakeRecognition))).toBe(true);
    expect(isRealtimeSpeechSupported({ webkitSpeechRecognition: FakeRecognition })).toBe(true);
    expect(isRealtimeSpeechSupported({})).toBe(false);
  });

  it("normalizes partial and final transcript events", () => {
    expect(normalizeSpeechRecognitionResult(resultEvent([" I am ", " speaking"], false))).toEqual({
      transcript: "I am speaking",
      isFinal: false,
    });

    expect(normalizeSpeechRecognitionResult(resultEvent([" I am finished. "], true))).toEqual({
      transcript: "I am finished.",
      isFinal: true,
    });
  });

  it("starts recognition with continuous interim English settings", () => {
    const instances: FakeRecognition[] = [];
    class CapturingRecognition extends FakeRecognition {
      constructor() {
        super();
        instances.push(this);
      }
    }
    const onPartial = vi.fn();
    const onFinal = vi.fn();
    const session = createRealtimeSpeechSession({
      host: makeHost(CapturingRecognition),
      onPartial,
      onFinal,
    });

    session.start();
    const instance = instances[0];
    instance?.onresult?.(resultEvent(["This is partial"], false));
    instance?.onresult?.(resultEvent(["This is final"], true));

    expect(instance?.continuous).toBe(true);
    expect(instance?.interimResults).toBe(true);
    expect(instance?.lang).toBe("en-US");
    expect(instance?.start).toHaveBeenCalledOnce();
    expect(onPartial).toHaveBeenCalledWith("This is partial");
    expect(onFinal).toHaveBeenCalledWith("This is final");
  });

  it("reports unsupported browsers as startup errors", () => {
    const onError = vi.fn();
    const session = createRealtimeSpeechSession({
      host: {},
      onPartial: vi.fn(),
      onFinal: vi.fn(),
      onError,
    });

    session.start();

    expect(onError).toHaveBeenCalledWith("realtime_speech_unsupported");
  });
});
