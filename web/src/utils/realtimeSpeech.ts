/** SpeechRecognitionLike 描述浏览器语音识别实例的最小接口。 */
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

/** SpeechRecognitionConstructor 描述语音识别构造器。 */
type SpeechRecognitionConstructor = new () => SpeechRecognitionLike;

/** SpeechRecognitionHost 描述可能提供语音识别能力的宿主对象。 */
interface SpeechRecognitionHost {
  SpeechRecognition?: SpeechRecognitionConstructor;
  webkitSpeechRecognition?: SpeechRecognitionConstructor;
}

/** RealtimeSpeechSessionOptions 定义实时语音会话回调。 */
export interface RealtimeSpeechSessionOptions {
  host?: SpeechRecognitionHost;
  language?: string;
  onPartial: (transcript: string) => void;
  onFinal: (transcript: string) => void;
  onError?: (code: string) => void;
  onEnd?: () => void;
}

/** RealtimeSpeechSession 定义实时语音会话控制接口。 */
export interface RealtimeSpeechSession {
  start: () => void;
  stop: () => void;
  abort: () => void;
}

/** RealtimeSpeechFallbackState 描述实时语音降级判断所需状态。 */
export interface RealtimeSpeechFallbackState {
  finalReceived?: boolean;
  stopRequested?: boolean;
  fallbackAttempted?: boolean;
}

/** NormalizedSpeechResult 描述归一化后的语音识别结果。 */
interface NormalizedSpeechResult {
  transcript: string;
  isFinal: boolean;
}

const nonFallbackSpeechErrors = new Set(["no-speech", "not-allowed", "realtime_speech_unsupported"]);

/** browserSpeechHost 读取浏览器实时语音识别宿主对象。 */
function browserSpeechHost(): SpeechRecognitionHost {
  if (typeof window === "undefined") {
    return {};
  }

  return window as unknown as SpeechRecognitionHost;
}

/** speechRecognitionConstructor 获取当前浏览器可用的语音识别构造器。 */
function speechRecognitionConstructor(host: SpeechRecognitionHost = browserSpeechHost()) {
  return host.SpeechRecognition ?? host.webkitSpeechRecognition;
}

/** isRealtimeSpeechSupported 判断浏览器是否支持实时语音识别。 */
export function isRealtimeSpeechSupported(host: SpeechRecognitionHost = browserSpeechHost()) {
  return Boolean(speechRecognitionConstructor(host));
}

/** shouldFallbackToRecordedAudio 判断实时识别错误是否应降级到录音上传。 */
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

/** normalizeSpeechRecognitionResult 将浏览器识别事件归一为文本和完成状态。 */
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

/** createRealtimeSpeechSession 创建浏览器实时语音识别会话。 */
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
