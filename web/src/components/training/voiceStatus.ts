import type { VoiceStatus } from "../../types";

export const voiceStatusOrder = ["idle", "recording", "recognizing", "thinking", "speaking"] as const satisfies readonly VoiceStatus[];

const nextStatusByStatus: Record<VoiceStatus, VoiceStatus> = {
  idle: "recording",
  recording: "recognizing",
  recognizing: "thinking",
  thinking: "speaking",
  speaking: "idle",
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
    title: "正在实时听写",
    description: "系统会持续显示英文转写，结束后进入本轮对话。",
    action: "结束",
  },
  recognizing: {
    title: "正在提交转写",
    description: "最终转写已生成，正在发送给 AI 教练。",
    action: "等待",
  },
  thinking: {
    title: "AI 正在思考",
    description: "AI 面试官正在生成追问和实时反馈。",
    action: "重置",
  },
  speaking: {
    title: "AI 正在说话",
    description: "正在播放本轮 AI 回复，播放结束后可以继续练习。",
    action: "等待",
  },
};

/** getNextVoiceStatus 根据当前语音状态推导按钮的下一个状态。 */
export function getNextVoiceStatus(status: VoiceStatus) {
  return nextStatusByStatus[status];
}
