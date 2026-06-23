import { useCallback, useEffect, useRef, useState } from "react";
import { createTextToSpeechPlayer, type TextToSpeechPlayer, type TextToSpeechPlayerOptions } from "../utils/tts";

type TextToSpeechPlayerFactory = (options?: TextToSpeechPlayerOptions) => TextToSpeechPlayer;

interface UseTTSPlaybackOptions {
  setStreamNotice: (message: string) => void;
  onPlaybackEnd?: () => void;
  createPlayer?: TextToSpeechPlayerFactory;
}

/** useTTSPlayback 管理 AI 回复的语音播放和去重。 */
export function useTTSPlayback({
  setStreamNotice,
  onPlaybackEnd,
  createPlayer = createTextToSpeechPlayer,
}: UseTTSPlaybackOptions) {
  const [isSpeaking, setIsSpeaking] = useState(false);
  const shouldSpeakNextAIRef = useRef(false);
  const spokenAIMessageIdsRef = useRef<Set<number>>(new Set());
  const playerRef = useRef<TextToSpeechPlayer | null>(null);
  const onPlaybackEndRef = useRef(onPlaybackEnd);

  useEffect(() => {
    onPlaybackEndRef.current = onPlaybackEnd;
  }, [onPlaybackEnd]);

  const textToSpeechPlayer = useCallback(() => {
    if (!playerRef.current) {
      playerRef.current = createPlayer({
        onUnavailable: () => {
          setStreamNotice("当前浏览器不支持 AI 语音播放，已保留文本回复。");
          setIsSpeaking(false);
          onPlaybackEndRef.current?.();
        },
        onEnd: () => {
          setIsSpeaking(false);
          onPlaybackEndRef.current?.();
        },
      });
    }

    return playerRef.current;
  }, [createPlayer, setStreamNotice]);

  const requestSpeakNextAI = useCallback(() => {
    shouldSpeakNextAIRef.current = true;
  }, []);

  const cancelPendingAIReply = useCallback(() => {
    shouldSpeakNextAIRef.current = false;
    setIsSpeaking(false);
  }, []);

  const speakAIReply = useCallback(
    (messageId: number, content: string) => {
      const reply = content.trim();
      if (!shouldSpeakNextAIRef.current || !reply) {
        return;
      }
      if (messageId > 0 && spokenAIMessageIdsRef.current.has(messageId)) {
        return;
      }
      if (messageId > 0) {
        spokenAIMessageIdsRef.current.add(messageId);
      }
      shouldSpeakNextAIRef.current = false;
      setIsSpeaking(true);
      textToSpeechPlayer().speak(reply);
    },
    [textToSpeechPlayer],
  );

  const cancelPlayback = useCallback(() => {
    shouldSpeakNextAIRef.current = false;
    setIsSpeaking(false);
    playerRef.current?.cancel();
  }, []);

  useEffect(() => cancelPlayback, [cancelPlayback]);

  return {
    isSpeaking,
    requestSpeakNextAI,
    cancelPendingAIReply,
    speakAIReply,
    cancelPlayback,
  };
}
