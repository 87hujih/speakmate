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
    description: "录制一段英文回答，结束后上传转写。",
    action: "开始",
  },
  recording: {
    title: "正在听你说",
    description: "保持英文回答连贯，结束后进行整段识别。",
    action: "结束",
  },
  recognizing: {
    title: "正在识别",
    description: "ASR 正在处理整段回答，完成后会接入本轮训练反馈。",
    action: "等待",
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
