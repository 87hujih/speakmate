/** cn 合并条件样式类名并过滤空值。 */
export function cn(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(" ");
}
