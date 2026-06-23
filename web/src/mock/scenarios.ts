import type { Scenario } from "../types";

export const scenarios: Scenario[] = [
  {
    id: 1,
    code: "interview",
    name: "英语面试",
    englishName: "Interview Practice",
    description: "练习自我介绍、项目经历和技术追问",
    difficulty: "medium",
    difficultyLabel: "中等",
    aiRole: "外企技术面试官",
    userGoal: "完成自我介绍、项目经历表达和技术追问回答",
    openingMessage:
      "Hello, welcome to the interview. Could you start by briefly introducing yourself and telling me about one project you are proud of?",
    goals: ["英文自我介绍", "项目经历表达", "技术追问回答"],
    stages: [
      { name: "自我介绍", description: "介绍教育背景、技术方向和当前目标。" },
      { name: "项目经历", description: "说明项目目标、个人职责和关键成果。" },
      { name: "技术追问", description: "回答实现细节、技术选择和问题排查过程。" },
      { name: "结尾提问", description: "向面试官提出关于团队、岗位或项目的问题。" },
    ],
    rubric: [
      { name: "流利度", description: "回答是否连贯，是否能自然展开说明。" },
      { name: "语法准确度", description: "时态、主谓一致和句子结构是否准确。" },
      { name: "表达自然度", description: "是否使用面试场景中自然、得体的表达。" },
      { name: "场景完成度", description: "是否完成自我介绍、项目说明和反问环节。" },
    ],
    sessionId: "interview-20260607-001",
  },
  {
    id: 2,
    code: "restaurant",
    name: "餐厅点餐",
    englishName: "Restaurant Ordering",
    description: "练习询问菜单、表达偏好和处理特殊需求",
    difficulty: "easy",
    difficultyLabel: "简单",
    aiRole: "餐厅服务员",
    userGoal: "完成菜单询问、偏好表达、特殊需求说明和订单确认。",
    openingMessage:
      "Good evening. Welcome to SpeakMate Bistro. Would you like to hear today’s recommendations?",
    goals: ["询问菜单推荐", "表达饮食偏好", "完成付款沟通"],
    stages: [
      { name: "入座问候", description: "自然回应服务员问候并说明需求。" },
      { name: "菜单询问", description: "询问菜品、饮料和推荐选择。" },
      { name: "特殊需求", description: "表达忌口、过敏或口味偏好。" },
      { name: "确认付款", description: "确认订单并完成结账沟通。" },
    ],
    rubric: [
      { name: "流利度", description: "点餐表达是否顺畅。" },
      { name: "语法准确度", description: "请求句式和数量表达是否准确。" },
      { name: "表达自然度", description: "礼貌表达是否符合餐厅场景。" },
      { name: "场景完成度", description: "是否完成点餐和确认流程。" },
    ],
    sessionId: "restaurant-20260607-001",
  },
  {
    id: 3,
    code: "meeting",
    name: "工作会议",
    englishName: "Work Meeting",
    description: "练习表达观点、澄清问题和总结结论",
    difficulty: "hard",
    difficultyLabel: "较难",
    aiRole: "项目同事",
    userGoal: "在会议中清楚表达进展、风险、观点和下一步行动。",
    openingMessage:
      "Let’s start the weekly sync. Could you give a short update on your current progress?",
    goals: ["表达观点", "汇报项目进展", "回应质疑与追问"],
    stages: [
      { name: "进展汇报", description: "说明本周完成的关键事项。" },
      { name: "问题澄清", description: "解释阻塞点和资源需求。" },
      { name: "方案讨论", description: "提出可执行的解决方案。" },
      { name: "行动确认", description: "总结下一步行动和负责人。" },
    ],
    rubric: [
      { name: "流利度", description: "能否连续表达会议观点。" },
      { name: "语法准确度", description: "时态、条件句和从句结构是否清晰。" },
      { name: "表达自然度", description: "是否使用职场会议常用表达。" },
      { name: "场景完成度", description: "是否完成汇报、讨论和总结。" },
    ],
    sessionId: "meeting-20260607-001",
  },
];

export const interviewScenario = scenarios[0];

/** getScenarioBySessionId 封装当前模块的辅助逻辑。 */
export function getScenarioBySessionId(sessionId: string) {
  return scenarios.find((scenario) => scenario.sessionId === sessionId) ?? interviewScenario;
}
