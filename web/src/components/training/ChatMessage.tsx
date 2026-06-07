import { Bot, UserRound } from "lucide-react";
import type { ChatMessage as ChatMessageType } from "../../types";
import { cn } from "../../utils/cn";

interface ChatMessageProps {
  message: ChatMessageType;
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isUser = message.role === "user";
  const Icon = isUser ? UserRound : Bot;

  return (
    <div className={cn("mb-5 flex items-start gap-3", isUser && "flex-row-reverse")}>
      <div
        className={cn(
          "grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-gradient-to-br from-blue-50 to-indigo-100 text-brand-blue",
          isUser && "from-violet-50 to-violet-100 text-brand-purple",
        )}
      >
        <Icon className="h-5 w-5" />
      </div>

      <div className="max-w-[78%]">
        <div className={cn("mb-1.5 flex items-center gap-2 text-xs font-extrabold text-slate-400", isUser && "justify-end")}>
          <span>{message.speaker}</span>
          <span>{message.createdAt}</span>
        </div>
        <div
          className={cn(
            "rounded-[20px] px-4 py-3 text-[15px] leading-7",
            isUser ? "rounded-tr-lg bg-violet-50 text-purple-950" : "rounded-tl-lg bg-blue-50 text-blue-950",
          )}
        >
          {message.content}
          {message.isTyping ? (
            <span className="ml-2 inline-flex items-center gap-1">
              {[0, 1, 2].map((item) => (
                <i
                  key={item}
                  className="typing-dot h-1.5 w-1.5 rounded-full bg-blue-400"
                  style={{ animationDelay: `${item * 0.15}s` }}
                />
              ))}
            </span>
          ) : null}
        </div>
        {message.meta ? (
          <div className="mt-2 flex justify-end gap-2">
            <span className="rounded-full border border-line bg-slate-50 px-2.5 py-1 text-[11px] font-extrabold text-muted">
              ASR {message.meta.asrConfidence}%
            </span>
            <span className="rounded-full border border-line bg-slate-50 px-2.5 py-1 text-[11px] font-extrabold text-muted">
              {message.meta.wpm} WPM
            </span>
            <span className="rounded-full border border-line bg-slate-50 px-2.5 py-1 text-[11px] font-extrabold text-muted">
              停顿 {message.meta.pauses} 次
            </span>
          </div>
        ) : null}
      </div>
    </div>
  );
}
