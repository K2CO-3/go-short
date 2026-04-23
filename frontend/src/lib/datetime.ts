/** 用于 datetime-local 的默认值：从现在起若干天后的本地时间字符串 */
export function defaultLocalDatetime(daysFromNow: number): string {
  const t = new Date();
  t.setDate(t.getDate() + daysFromNow);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${t.getFullYear()}-${pad(t.getMonth() + 1)}-${pad(t.getDate())}T${pad(t.getHours())}:${pad(t.getMinutes())}`;
}

/** datetime-local 值转 RFC3339，供 JSON 提交 */
export function localDatetimeToIso(local: string): string | undefined {
  if (!local.trim()) return undefined;
  const d = new Date(local);
  if (Number.isNaN(d.getTime())) return undefined;
  return d.toISOString();
}

/** 本地日历日期，供 `input[type=date]` 使用 */
export function formatInputDate(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export function addDays(d: Date, delta: number): Date {
  const next = new Date(d);
  next.setDate(next.getDate() + delta);
  return next;
}

/**
 * 将访问日志筛选项中的「日期 + 时间」合并为后端接受的字符串：
 * - 仅日期：`2006-01-02`
 * - 日期+时间：`2006-01-02 15:04:05`（与 admin parseAccessLogTime 中第二种 layout 一致）
 */
export function buildAccessLogTimeParam(
  date: string,
  time: string
): { value?: string; error?: string } {
  const d = date.trim();
  const t = time.trim();
  if (!d && !t) {
    return { value: undefined };
  }
  if (!d && t) {
    return { error: "选了时刻就要先选左边的日期" };
  }
  if (d && !/^\d{4}-\d{2}-\d{2}$/.test(d)) {
    return { error: "请检查日期是否选完整" };
  }
  if (d && !t) {
    return { value: d };
  }
  // d && t
  let timeNorm = t;
  if (/^\d{2}:\d{2}$/.test(t)) {
    timeNorm = `${t}:00`;
  }
  if (!/^\d{2}:\d{2}:\d{2}$/.test(timeNorm)) {
    return { error: "时间请用 时:分 或 时:分:秒" };
  }
  return { value: `${d} ${timeNorm}` };
}

/**
 * 起止时间是否颠倒。仅日期时：开始按当天 0 点，结束按当天最后一刻，避免误报。
 */
export function isAccessLogRangeReversed(
  start?: string,
  end?: string
): boolean {
  if (!start || !end) {
    return false;
  }
  const startMs = rangeBoundaryMs(start, "start");
  const endMs = rangeBoundaryMs(end, "end");
  if (Number.isNaN(startMs) || Number.isNaN(endMs)) {
    return false;
  }
  return startMs > endMs;
}

function rangeBoundaryMs(s: string, kind: "start" | "end"): number {
  if (/^\d{4}-\d{2}-\d{2}$/.test(s)) {
    if (kind === "start") {
      return new Date(s + "T00:00:00").getTime();
    }
    return new Date(s + "T23:59:59.999").getTime();
  }
  if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(s)) {
    return new Date(s.replace(" ", "T")).getTime();
  }
  return Number.NaN;
}
