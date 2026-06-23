import { FormEvent, KeyboardEvent, useEffect, useRef } from "react";
import { LoaderCircle, MessageSquareText, SendHorizontal, TimerReset } from "lucide-react";
import type { ChatMessage as ChatMessageType, VoiceStatus } from "../../types";
import { ChatMessage } from "./ChatMessage";
import { VoiceRecorder } from "./VoiceRecorder";

/** ConversationPanelProps 定义对应组件接收的属性。 */
interface ConversationPanelProps {
  currentStage: string;
  messages: ChatMessageType[];
  turnCount: number;
  draft: string;
  isSending: boolean;
  isDisabled: boolean;
  error?: string;
  streamNotice?: string;
  voiceStatus: VoiceStatus;
  voiceTranscript?: string;
  voiceError?: string;
  isVoiceDisabled: boolean;
  onDraftChange: (value: string) => void;
  onSend: () => void;
  onVoiceToggle: () => void;
}

/** ConversationPanel 渲染对应的页面或界面组件。 */
export function ConversationPanel({
  currentStage,
  messages,
  turnCount,
  draft,
  isSending,
  isDisabled,
  error,
  streamNotice,
  voiceStatus,
  voiceTranscript,
  voiceError,
  isVoiceDisabled,
  onDraftChange,
  onSend,
  onVoiceToggle,
}: ConversationPanelProps) {
  const scrollRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const element = scrollRef.current;
    if (element) {
      element.scrollTop = element.scrollHeight;
    }
  }, [messages]);

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    onSend();
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
      event.preventDefault();
      onSend();
    }
  }

  return (
    <section className="flex h-[calc(100vh-120px)] max-h-[760px] min-h-[520px] flex-col overflow-hidden rounded-panel border border-line bg-white/95 shadow-panel xl:h-full xl:min-h-0 xl:max-h-none">
      <div className="flex min-h-[62px] shrink-0 flex-wrap items-center justify-between gap-3 border-b border-line px-4 py-3 md:px-5">
        <div className="flex items-center gap-3">
          <span className="grid h-10 w-10 place-items-center rounded-[16px] bg-blue-50 text-brand-blue">
            <MessageSquareText className="h-5 w-5" />
          </span>
          <div>
            <h2 className="m-0 text-[17px] font-black tracking-[-0.01em] text-ink">实时对话</h2>
            <p className="m-0 text-xs font-bold text-muted">AI 正在围绕 {currentStage} 推进训练</p>
          </div>
        </div>
        <div className="inline-flex items-center gap-2 rounded-full border border-line bg-slate-50 px-3 py-2 text-xs font-black text-muted">
          <TimerReset className="h-4 w-4 text-brand-blue" />
          {turnCount} 轮对话
        </div>
      </div>

      <div ref={scrollRef} className="chat-scroll min-h-0 flex-1 overflow-y-auto px-4 py-5 md:px-6">
        {messages.length ? (
          messages.map((message) => <ChatMessage key={message.id} message={message} />)
        ) : (
          <div className="grid h-full min-h-[220px] place-items-center text-center">
            <div>
              <h3 className="m-0 text-xl font-black text-ink">训练即将开始</h3>
              <p className="mt-2 text-sm font-semibold text-muted">输入第一句英文，AI 会继续追问并生成反馈。</p>
            </div>
          </div>
        )}
      </div>

      <form onSubmit={handleSubmit} className="shrink-0 border-t border-line bg-white px-4 py-4 md:px-5">
        {streamNotice ? <div className="mb-2 rounded-2xl bg-blue-50 px-3 py-2 text-xs font-bold text-blue-700">{streamNotice}</div> : null}
        {error ? <div className="mb-2 rounded-2xl bg-rose-50 px-3 py-2 text-xs font-bold text-rose-700">{error}</div> : null}
        <div className="grid grid-cols-[minmax(0,1fr)_52px] items-end gap-3 rounded-[24px] border border-line bg-slate-50/90 p-3 shadow-soft">
          <textarea
            value={draft}
            disabled={isDisabled || isSending}
            onChange={(event) => onDraftChange(event.target.value)}
            onKeyDown={handleKeyDown}
            rows={3}
            placeholder={isDisabled ? "本次训练已结束" : "输入你的英文回答..."}
            className="max-h-40 min-h-[78px] resize-y border-0 bg-transparent px-2 py-1 text-[15px] font-semibold leading-7 text-ink outline-none placeholder:text-slate-400 disabled:cursor-not-allowed"
          />
          <button
            type="submit"
            disabled={isDisabled || isSending}
            className="grid h-[52px] w-[52px] place-items-center rounded-[18px] bg-gradient-to-br from-brand-blue to-brand-purple text-white shadow-glow transition hover:-translate-y-0.5 active:translate-y-px disabled:cursor-not-allowed disabled:opacity-60"
            aria-label="发送消息"
          >
            {isSending ? <LoaderCircle className="h-5 w-5 animate-spin" /> : <SendHorizontal className="h-5 w-5" />}
          </button>
        </div>
      </form>
      <VoiceRecorder
        status={voiceStatus}
        transcript={voiceTranscript}
        error={voiceError}
        isDisabled={isVoiceDisabled}
        onToggle={onVoiceToggle}
      />
    </section>
  );
}
