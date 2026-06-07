interface SpeechSynthesisLike {
  cancel: () => void;
  speak: (utterance: SpeechSynthesisUtteranceLike) => void;
}

interface SpeechSynthesisUtteranceLike {
  text: string;
  lang: string;
  rate: number;
  onend: (() => void) | null;
}

type SpeechSynthesisUtteranceConstructor = new (text: string) => SpeechSynthesisUtteranceLike;

interface TextToSpeechHost {
  speechSynthesis?: SpeechSynthesisLike;
  SpeechSynthesisUtterance?: SpeechSynthesisUtteranceConstructor;
}

export interface TextToSpeechPlayerOptions {
  host?: TextToSpeechHost;
  language?: string;
  rate?: number;
  onUnavailable?: () => void;
  onEnd?: () => void;
}

export interface TextToSpeechPlayer {
  speak: (text: string) => void;
  cancel: () => void;
}

function browserTTSHost(): TextToSpeechHost {
  if (typeof window === "undefined") {
    return {};
  }

  return window as unknown as TextToSpeechHost;
}

export function isTextToSpeechSupported(host: TextToSpeechHost = browserTTSHost()) {
  return Boolean(host.speechSynthesis && host.SpeechSynthesisUtterance);
}

export function createTextToSpeechPlayer({
  host = browserTTSHost(),
  language = "en-US",
  rate = 1,
  onUnavailable,
  onEnd,
}: TextToSpeechPlayerOptions = {}): TextToSpeechPlayer {
  return {
    speak(text: string) {
      const content = text.trim();
      if (!content) {
        return;
      }
      if (!host.speechSynthesis || !host.SpeechSynthesisUtterance) {
        onUnavailable?.();
        return;
      }

      const utterance = new host.SpeechSynthesisUtterance(content);
      utterance.lang = language;
      utterance.rate = rate;
      utterance.onend = () => {
        onEnd?.();
      };
      host.speechSynthesis.cancel();
      host.speechSynthesis.speak(utterance);
    },
    cancel() {
      host.speechSynthesis?.cancel();
    },
  };
}
