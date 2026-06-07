export interface SpeechRecognitionLike {
  continuous: boolean;
  interimResults: boolean;
  lang: string;
  onresult: ((event: unknown) => void) | null;
  onerror: ((event: unknown) => void) | null;
  onend: (() => void) | null;
  start: () => void;
  stop: () => void;
  abort: () => void;
}

type SpeechRecognitionConstructor = new () => SpeechRecognitionLike;

interface SpeechRecognitionHost {
  SpeechRecognition?: SpeechRecognitionConstructor;
  webkitSpeechRecognition?: SpeechRecognitionConstructor;
}

export interface RealtimeSpeechSessionOptions {
  host?: SpeechRecognitionHost;
  language?: string;
  onPartial: (transcript: string) => void;
  onFinal: (transcript: string) => void;
  onError?: (code: string) => void;
  onEnd?: () => void;
}

export interface RealtimeSpeechSession {
  start: () => void;
  stop: () => void;
  abort: () => void;
}

export interface RealtimeSpeechFallbackState {
  finalReceived?: boolean;
  stopRequested?: boolean;
  fallbackAttempted?: boolean;
}

interface NormalizedSpeechResult {
  transcript: string;
  isFinal: boolean;
}

const nonFallbackSpeechErrors = new Set(["no-speech", "not-allowed", "realtime_speech_unsupported"]);

function browserSpeechHost(): SpeechRecognitionHost {
  if (typeof window === "undefined") {
    return {};
  }

  return window as unknown as SpeechRecognitionHost;
}

function speechRecognitionConstructor(host: SpeechRecognitionHost = browserSpeechHost()) {
  return host.SpeechRecognition ?? host.webkitSpeechRecognition;
}

export function isRealtimeSpeechSupported(host: SpeechRecognitionHost = browserSpeechHost()) {
  return Boolean(speechRecognitionConstructor(host));
}

export function shouldFallbackToRecordedAudio(code: string, state: RealtimeSpeechFallbackState = {}) {
  if (state.finalReceived || state.stopRequested || state.fallbackAttempted) {
    return false;
  }

  const normalizedCode = code.trim().toLowerCase();
  if (!normalizedCode) {
    return true;
  }

  return !nonFallbackSpeechErrors.has(normalizedCode);
}

export function normalizeSpeechRecognitionResult(event: unknown): NormalizedSpeechResult {
  const results = (event as { results?: ArrayLike<unknown> }).results;
  if (!results || results.length === 0) {
    return { transcript: "", isFinal: false };
  }

  const parts: string[] = [];
  let isFinal = false;
  for (let index = 0; index < results.length; index += 1) {
    const result = results[index] as { isFinal?: boolean; 0?: { transcript?: string } };
    const transcript = result?.[0]?.transcript?.trim();
    if (transcript) {
      parts.push(transcript);
    }
    isFinal = Boolean(result?.isFinal);
  }

  return {
    transcript: parts.join(" ").trim(),
    isFinal,
  };
}

export function createRealtimeSpeechSession({
  host = browserSpeechHost(),
  language = "en-US",
  onPartial,
  onFinal,
  onError,
  onEnd,
}: RealtimeSpeechSessionOptions): RealtimeSpeechSession {
  const Recognition = speechRecognitionConstructor(host);
  let recognition: SpeechRecognitionLike | null = null;

  function ensureRecognition() {
    if (!Recognition) {
      onError?.("realtime_speech_unsupported");
      return null;
    }
    if (recognition) {
      return recognition;
    }

    recognition = new Recognition();
    recognition.continuous = true;
    recognition.interimResults = true;
    recognition.lang = language;
    recognition.onresult = (event) => {
      const result = normalizeSpeechRecognitionResult(event);
      if (!result.transcript) {
        return;
      }
      if (result.isFinal) {
        onFinal(result.transcript);
        return;
      }
      onPartial(result.transcript);
    };
    recognition.onerror = (event) => {
      const code = (event as { error?: string }).error || "realtime_speech_error";
      onError?.(code);
    };
    recognition.onend = () => {
      onEnd?.();
    };

    return recognition;
  }

  return {
    start() {
      ensureRecognition()?.start();
    },
    stop() {
      recognition?.stop();
    },
    abort() {
      recognition?.abort();
    },
  };
}
