import type { Correction, ScoreDimension, TrainingSession } from "../types";
import { interviewScenario, scenarios } from "./scenarios";

export const scoreDimensions: ScoreDimension[] = [
  {
    key: "fluency",
    name: "流利度",
    score: 75,
    description: "整体表达连贯，但回答技术细节时有轻微停顿。",
  },
  {
    key: "grammar",
    name: "语法准确度",
    score: 72,
    description: "主要问题集中在进行时和完成时动词形式。",
  },
  {
    key: "expression",
    name: "表达自然度",
    score: 80,
    description: "能表达核心意思，部分句子可以更符合面试语境。",
  },
  {
    key: "vocabulary",
    name: "词汇丰富度",
    score: 76,
    description: "使用了 backend、Redis、MySQL 等场景词，但动词略单一。",
  },
  {
    key: "completion",
    name: "场景完成度",
    score: 85,
    description: "已完成自我介绍和项目说明，技术追问仍在推进中。",
  },
];

export const corrections: Correction[] = [
  {
    title: "动词时态和完成时错误",
    category: "grammar",
    original: "I am study computer science and I have did a project.",
    suggestion: "I am studying computer science, and I have done a project.",
    explanation: "be 动词后需要使用现在分词，have 后需要使用过去分词。",
    issues: ["am study → am studying", "have did → have done"],
  },
  {
    title: "项目经历表达不够自然",
    category: "grammar",
    original: "My work is make the backend service.",
    suggestion: "I was responsible for building the backend service.",
    explanation: "面试中说明个人职责时，I was responsible for 更自然。",
  },
  {
    title: "项目关键词可以更具体",
    category: "vocabulary",
    original: "a project about robot control",
    suggestion: "a robotics project focused on motion control",
    explanation: "用 focused on motion control 能更清楚说明项目方向。",
  },
];

export const interviewSession: TrainingSession = {
  sessionId: interviewScenario.sessionId,
  sessionNo: "S202606070001",
  scenario: interviewScenario,
  status: "running",
  liveStatus: "WebSocket 已连接",
  durationLabel: "04:32",
  turnCount: 8,
  progress: 60,
  currentStage: "技术追问",
  voiceStatus: "idle",
  tasks: [
    { label: "英文自我介绍", status: "done" },
    { label: "项目经历表达", status: "done" },
    { label: "技术追问回答", status: "active" },
    { label: "向面试官提问", status: "pending" },
  ],
  focusTags: ["回答结构", "语法准确性", "项目表达", "临场反应"],
  messages: [
    {
      id: 1,
      role: "ai",
      speaker: "AI 面试官",
      content: "Could you briefly introduce yourself and tell me about your technical background?",
      stage: "自我介绍",
      createdAt: "04:28",
    },
    {
      id: 2,
      role: "user",
      speaker: "You",
      content: "I am study computer science and I have did a project.",
      stage: "自我介绍",
      createdAt: "04:29",
      meta: { asrConfidence: 86, wpm: 124, pauses: 2 },
    },
    {
      id: 3,
      role: "ai",
      speaker: "AI 面试官",
      content: "That sounds interesting. Could you explain your main responsibility in the robot control project?",
      stage: "项目经历",
      createdAt: "04:30",
    },
    {
      id: 4,
      role: "user",
      speaker: "You",
      content: "My work is make the backend service and connect it with Redis and MySQL.",
      stage: "项目经历",
      createdAt: "04:31",
      meta: { asrConfidence: 91, wpm: 118, pauses: 1 },
    },
    {
      id: 5,
      role: "ai",
      speaker: "AI 面试官",
      content: "Great. What challenge did you meet when designing the backend service?",
      stage: "技术追问",
      createdAt: "正在回复",
      isTyping: true,
    },
  ],
  currentScore: 78,
  coachSummary: "表达能被理解，但语法准确性和项目贡献描述还可以进一步优化。",
  scores: scoreDimensions,
  corrections,
  naturalExpression:
    "I’m currently studying computer science, and I recently worked on a robotics project focused on motion control.",
};

export const trainingSessions: Record<string, TrainingSession> = {
  [interviewSession.sessionId]: interviewSession,
  [scenarios[1].sessionId]: {
    ...interviewSession,
    sessionId: scenarios[1].sessionId,
    scenario: scenarios[1],
    sessionNo: "S202606070002",
    currentStage: "菜单询问",
    progress: 45,
    currentScore: 84,
    coachSummary: "点餐流程推进顺利，礼貌请求比较自然，特殊需求表达可以再具体一些。",
    tasks: [
      { label: "回应服务员问候", status: "done" },
      { label: "询问菜单推荐", status: "active" },
      { label: "表达饮食偏好", status: "pending" },
      { label: "确认订单和付款", status: "pending" },
    ],
    focusTags: ["礼貌请求", "饮食偏好", "订单确认", "自然回应"],
    messages: [
      {
        id: 1,
        role: "ai",
        speaker: "AI 服务员",
        content: "Good evening. Would you like to hear today’s recommendations?",
        stage: "入座问候",
        createdAt: "02:14",
      },
      {
        id: 2,
        role: "user",
        speaker: "You",
        content: "Yes, I want something not spicy and maybe with chicken.",
        stage: "菜单询问",
        createdAt: "02:15",
        meta: { asrConfidence: 93, wpm: 112, pauses: 1 },
      },
      {
        id: 3,
        role: "ai",
        speaker: "AI 服务员",
        content: "Sure. We have grilled chicken salad and chicken risotto. Do you have any allergies?",
        stage: "特殊需求",
        createdAt: "正在回复",
        isTyping: true,
      },
    ],
    naturalExpression: "I’d like something mild, preferably with chicken. Do you have any recommendations?",
  },
  [scenarios[2].sessionId]: {
    ...interviewSession,
    sessionId: scenarios[2].sessionId,
    scenario: scenarios[2],
    sessionNo: "S202606070003",
    currentStage: "方案讨论",
    progress: 55,
    currentScore: 81,
    coachSummary: "会议表达主线清楚，但对阻塞点和下一步行动的说明还可以更完整。",
    tasks: [
      { label: "汇报当前进展", status: "done" },
      { label: "说明阻塞问题", status: "done" },
      { label: "讨论解决方案", status: "active" },
      { label: "确认下一步行动", status: "pending" },
    ],
    focusTags: ["观点展开", "澄清问题", "会议总结", "行动项表达"],
    messages: [
      {
        id: 1,
        role: "ai",
        speaker: "AI 同事",
        content: "Could you give a quick update on the backend integration work?",
        stage: "进展汇报",
        createdAt: "05:42",
      },
      {
        id: 2,
        role: "user",
        speaker: "You",
        content: "I finished most API, but there is one blocking problem about cache data.",
        stage: "问题澄清",
        createdAt: "05:43",
        meta: { asrConfidence: 89, wpm: 116, pauses: 2 },
      },
      {
        id: 3,
        role: "ai",
        speaker: "AI 同事",
        content: "Thanks. What options do we have to solve the cache consistency issue?",
        stage: "方案讨论",
        createdAt: "正在回复",
        isTyping: true,
      },
    ],
    naturalExpression: "I’ve finished most of the API work, but cache consistency is still blocking the integration.",
  },
};

export function getTrainingSession(sessionId: string) {
  return trainingSessions[sessionId] ?? interviewSession;
}
