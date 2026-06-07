import type { VoiceStatus } from "../../types";

export const voiceStatusOrder = ["idle", "recording", "recognizing", "thinking"] as const satisfies readonly VoiceStatus[];

const nextStatusByStatus: Record<VoiceStatus, VoiceStatus> = {
  idle: "recording",
  recording: "recognizing",
  recognizing: "thinking",
  thinking: "idle",
};

export const voiceStatusContent: Record<
  VoiceStatus,
  {
    title: string;
    description: string;
    action: string;
  }
> = {
  idle: {
    title: "点击开始说英语",
    description: "准备回答面试官的问题，系统会实时记录你的表达。",
    action: "开始",
  },
  recording: {
    title: "正在听你说",
    description: "保持英文回答连贯，结束后进入语音识别。",
    action: "结束",
  },
  recognizing: {
    title: "正在识别",
    description: "ASR 正在转写你的回答，并提取语法问题。",
    action: "继续",
  },
  thinking: {
    title: "AI 正在思考",
    description: "AI 面试官正在生成追问和实时反馈。",
    action: "重置",
  },
};

export function getNextVoiceStatus(status: VoiceStatus) {
  return nextStatusByStatus[status];
}
