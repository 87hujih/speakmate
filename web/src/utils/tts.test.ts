import { describe, expect, it, vi } from "vitest";
import { createTextToSpeechPlayer, isTextToSpeechSupported } from "./tts";

class FakeUtterance {
  text: string;
  lang = "";
  rate = 1;
  onend: (() => void) | null = null;

  constructor(text: string) {
    this.text = text;
  }
}

function makeHost() {
  return {
    speechSynthesis: {
      cancel: vi.fn(),
      speak: vi.fn(),
    },
    SpeechSynthesisUtterance: FakeUtterance,
  };
}

describe("tts helper", () => {
  it("detects browser speech synthesis support", () => {
    expect(isTextToSpeechSupported(makeHost())).toBe(true);
    expect(isTextToSpeechSupported({})).toBe(false);
  });

  it("plays a whole sentence with English settings", () => {
    const host = makeHost();
    const player = createTextToSpeechPlayer({ host });

    player.speak("  Could you explain your role?  ");

    expect(host.speechSynthesis.cancel).toHaveBeenCalledOnce();
    expect(host.speechSynthesis.speak).toHaveBeenCalledOnce();
    expect(host.speechSynthesis.speak).toHaveBeenCalledWith(
      expect.objectContaining({
        text: "Could you explain your role?",
        lang: "en-US",
        rate: 1,
      }),
    );
  });

  it("does not speak empty text", () => {
    const host = makeHost();
    const player = createTextToSpeechPlayer({ host });

    player.speak("   ");

    expect(host.speechSynthesis.cancel).not.toHaveBeenCalled();
    expect(host.speechSynthesis.speak).not.toHaveBeenCalled();
  });

  it("reports unavailable synthesis without throwing", () => {
    const onUnavailable = vi.fn();
    const player = createTextToSpeechPlayer({ host: {}, onUnavailable });

    player.speak("Hello");

    expect(onUnavailable).toHaveBeenCalledOnce();
  });

  it("calls onEnd when the utterance ends", () => {
    const host = makeHost();
    const onEnd = vi.fn();
    const player = createTextToSpeechPlayer({ host, onEnd });

    player.speak("Hello");
    const utterance = host.speechSynthesis.speak.mock.calls[0][0] as FakeUtterance;
    utterance.onend?.();

    expect(onEnd).toHaveBeenCalledOnce();
  });
});
