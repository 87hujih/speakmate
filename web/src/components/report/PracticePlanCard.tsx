import type { PracticePlanItem } from "../../types";

/** PracticePlanCardProps 定义对应组件接收的属性。 */
interface PracticePlanCardProps {
  item: PracticePlanItem;
  index: number;
}

/** PracticePlanCard 渲染对应的页面或界面组件。 */
export function PracticePlanCard({ item, index }: PracticePlanCardProps) {
  return (
    <div className="rounded-[18px] border border-line bg-gradient-to-br from-slate-50 to-white p-4">
      <strong className="mb-2 block text-sm text-ink">
        {index + 1}. {item.title}
      </strong>
      <span className="text-[13px] leading-6 text-muted">{item.description}</span>
    </div>
  );
}
