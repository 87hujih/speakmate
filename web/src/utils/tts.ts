/** SpeechSynthesisLike 描述浏览器语音合成实例的最小接口。 */
interface SpeechSynthesisLike {
  cancel: () => void;
  speak: (utterance: SpeechSynthesisUtteranceLike) => void;
}

/** SpeechSynthesisUtteranceLike 描述浏览器语音播报实例。 */
interface SpeechSynthesisUtteranceLike {
  text: string;
  lang: string;
  rate: number;
  onend: (() => void) | null;
}

/** SpeechSynthesisUtteranceConstructor 描述语音播报构造器。 */
type SpeechSynthesisUtteranceConstructor = new (text: string) => SpeechSynthesisUtteranceLike;

/** TextToSpeechHost 描述可能提供 TTS 能力的宿主对象。 */
interface TextToSpeechHost {
  speechSynthesis?: SpeechSynthesisLike;
  SpeechSynthesisUtterance?: SpeechSynthesisUtteranceConstructor;
}

/** TextToSpeechPlayerOptions 定义 TTS 播放器依赖。 */
export interface TextToSpeechPlayerOptions {
  host?: TextToSpeechHost;
  language?: string;
  rate?: number;
  onUnavailable?: () => void;
  onEnd?: () => void;
}

/** TextToSpeechPlayer 定义 TTS 播放器控制接口。 */
export interface TextToSpeechPlayer {
  speak: (text: string) => void;
  cancel: () => void;
}

/** browserTTSHost 读取浏览器 TTS 宿主对象。 */
function browserTTSHost(): TextToSpeechHost {
  if (typeof window === "undefined") {
    return {};
  }

  return window as unknown as TextToSpeechHost;
}

/** isTextToSpeechSupported 判断浏览器是否支持文本转语音。 */
export function isTextToSpeechSupported(host: TextToSpeechHost = browserTTSHost()) {
  return Boolean(host.speechSynthesis && host.SpeechSynthesisUtterance);
}

/** createTextToSpeechPlayer 创建可取消的文本转语音播放器。 */
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
