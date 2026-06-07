INSERT INTO scenarios (
  id, code, name, description, difficulty, ai_role, user_goal, opening_message, stages_json, rubric_json
) VALUES
(
  1,
  'interview',
  '英语面试',
  '练习自我介绍、项目经历和技术追问',
  'medium',
  '技术面试官',
  '清晰介绍自己的背景、项目经历和协作方式，并能回答常见技术追问。',
  'Hello, welcome to the interview. Could you start by briefly introducing yourself and telling me about one project you are proud of?',
  '[{"name":"自我介绍","description":"介绍教育背景、技术方向和当前目标。"},{"name":"项目经历","description":"说明项目目标、个人职责和关键成果。"},{"name":"技术追问","description":"回答实现细节、技术选择和问题排查过程。"},{"name":"结尾提问","description":"向面试官提出关于团队、岗位或项目的问题。"}]',
  '[{"name":"流利度","description":"回答是否连贯，是否能自然展开说明。"},{"name":"语法准确度","description":"时态、主谓一致和句子结构是否准确。"},{"name":"表达自然度","description":"是否使用面试场景中自然、得体的表达。"},{"name":"场景完成度","description":"是否完成自我介绍、项目说明和反问环节。"}]'
),
(
  2,
  'restaurant',
  '餐厅点餐',
  '练习询问菜单、表达偏好和处理特殊需求',
  'easy',
  '餐厅服务员',
  '完成点餐流程，清楚表达口味偏好、忌口和额外需求。',
  'Good evening. Welcome to our restaurant. Would you like to hear today''s specials before you order?',
  '[{"name":"询问菜单","description":"询问推荐菜、套餐或今日特色。"},{"name":"表达偏好","description":"说明口味、饮食限制或过敏信息。"},{"name":"确认订单","description":"确认菜品、饮品和额外要求。"}]',
  '[{"name":"清晰度","description":"点餐需求是否表达明确。"},{"name":"礼貌程度","description":"是否使用礼貌、自然的服务场景表达。"},{"name":"词汇匹配度","description":"是否使用菜单、口味和数量相关词汇。"}]'
),
(
  3,
  'meeting',
  '工作会议',
  '练习表达观点、澄清问题和总结结论',
  'medium',
  '项目同事',
  '在会议中清楚表达观点、回应问题，并推动形成下一步行动。',
  'Thanks for joining the meeting. Could you give us a quick update on your progress and any blockers you are seeing?',
  '[{"name":"进度同步","description":"说明当前进展、已完成事项和风险。"},{"name":"观点表达","description":"提出建议并解释理由。"},{"name":"澄清确认","description":"澄清问题、确认分工和下一步。"}]',
  '[{"name":"结构清晰度","description":"表达是否有清楚的先后顺序和重点。"},{"name":"互动能力","description":"是否能回应他人问题并主动澄清。"},{"name":"场景词汇","description":"是否使用会议、进度和协作相关表达。"},{"name":"结论完整度","description":"是否明确下一步行动和负责人。"}]'
)
ON DUPLICATE KEY UPDATE
  code = VALUES(code),
  name = VALUES(name),
  description = VALUES(description),
  difficulty = VALUES(difficulty),
  ai_role = VALUES(ai_role),
  user_goal = VALUES(user_goal),
  opening_message = VALUES(opening_message),
  stages_json = VALUES(stages_json),
  rubric_json = VALUES(rubric_json);
