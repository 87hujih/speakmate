import { Brain, LoaderCircle, Mic, Square } from "lucide-react";
import type { VoiceStatus } from "../../types";
import { cn } from "../../utils/cn";
import { Waveform } from "./Waveform";
import { voiceStatusContent, voiceStatusOrder } from "./voiceStatus";

interface VoiceRecorderProps {
  status: VoiceStatus;
  transcript?: string;
  error?: string;
  isDisabled?: boolean;
  onToggle?: () => void;
}

export function VoiceRecorder({ status, transcript, error, isDisabled = false, onToggle }: VoiceRecorderProps) {
  const copy = voiceStatusContent[status];
  const isActive = status !== "idle";
  const StatusIcon = status === "thinking" ? Brain : status === "recognizing" ? LoaderCircle : status === "recording" ? Square : Mic;
  const isButtonDisabled = isDisabled || status === "recognizing" || status === "thinking";

  return (
    <div className="border-t border-line bg-white px-5 py-4">
      <div className="grid grid-cols-1 items-center gap-4 rounded-[24px] border border-line bg-slate-50/90 p-4 shadow-soft md:grid-cols-[minmax(180px,0.9fr)_minmax(180px,1fr)_88px]">
        <div className="flex min-w-0 items-center gap-4">
          <div
            className={cn(
              "relative grid h-[58px] w-[58px] shrink-0 place-items-center rounded-[20px] text-white shadow-[0_18px_34px_rgba(37,99,235,0.22)]",
              status === "recording" ? "mic-pulse bg-rose-500" : "bg-gradient-to-br from-brand-blue to-brand-purple",
            )}
          >
            <StatusIcon className={cn("h-6 w-6", status === "recognizing" && "animate-spin")} fill={status === "recording" ? "currentColor" : "none"} />
          </div>
          <div className="min-w-0">
            <strong className="block truncate text-[15px] font-black text-ink">{copy.title}</strong>
            <span className="mt-1 block truncate text-[13px] font-semibold text-muted">{copy.description}</span>
          </div>
        </div>

        <div className="min-w-0">
          <Waveform active={isActive} />
          <div className="mt-2 grid grid-cols-4 gap-1.5">
            {voiceStatusOrder.map((item) => (
              <span
                key={item}
                className={cn(
                  "h-1.5 rounded-full bg-slate-200 transition",
                  item === status && "bg-brand-blue",
                  status === "recording" && item === status && "bg-rose-500",
                )}
              />
            ))}
          </div>
        </div>

        <button
          type="button"
          onClick={onToggle}
          disabled={isButtonDisabled}
          className={cn(
            "grid h-[72px] w-[72px] place-items-center justify-self-start rounded-full text-white shadow-glow transition duration-200 hover:-translate-y-0.5 active:translate-y-px disabled:cursor-not-allowed disabled:opacity-60 md:justify-self-end",
            status === "recording" ? "bg-rose-500 shadow-[0_20px_42px_rgba(244,63,94,0.28)]" : "bg-gradient-to-br from-brand-blue to-brand-purple",
          )}
          aria-label={`${copy.action}录音状态`}
        >
          <span className="grid gap-1 place-items-center text-[11px] font-black">
            <Mic className="h-6 w-6" />
            {copy.action}
          </span>
        </button>
      </div>
      {transcript ? <div className="mt-2 rounded-2xl bg-blue-50 px-3 py-2 text-xs font-bold text-blue-700">转写：{transcript}</div> : null}
      {error ? <div className="mt-2 rounded-2xl bg-rose-50 px-3 py-2 text-xs font-bold text-rose-700">{error}</div> : null}
    </div>
  );
}
