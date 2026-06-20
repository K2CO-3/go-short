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

/** 将 `datetime-local` 字符串拆成日历日期与时刻（供分开使用 date + time 控件） */
export function splitLocalDatetime(combined: string): { date: string; time: string } {
  const parts = combined.split("T");
  const date = parts[0] ?? "";
  const timeRaw = parts[1] ?? "";
  const time = timeRaw.length >= 5 ? timeRaw.slice(0, 5) : "00:00";
  return { date, time };
}

/** 日历日期 + 时刻（HH:mm）合并为 RFC3339，供 JSON 提交 */
export function dateAndTimeToIso(date: string, time: string): string | undefined {
  const d = date.trim();
  const t = time.trim();
  if (!d) return undefined;
  const local = t.length >= 5 ? `${d}T${t.slice(0, 5)}` : `${d}T00:00`;
  const dt = new Date(local);
  if (Number.isNaN(dt.getTime())) return undefined;
  return dt.toISOString();
}

/** 短链等场景：有过期时间且已早于当前时刻（无时区则按 Date 解析规则） */
export function isExpiresAtInPast(iso: string | null | undefined): boolean {
  if (!iso || !String(iso).trim()) return false;
  const t = new Date(iso).getTime();
  if (Number.isNaN(t)) return false;
  return t < Date.now();
}

/** 列表展示用：ISO 时间或「—」 */
export function formatExpiresAtDisplay(iso: string | null | undefined): string {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

/** 访问日志等：数据库存储为 UTC 时刻，API 通常带 Z；列表按东八区（中国）展示 */
export function formatUtcToShanghai(utcInput: string | null | undefined): string {
  if (utcInput == null || !String(utcInput).trim()) return "—";
  const d = new Date(utcInput);
  if (Number.isNaN(d.getTime())) return String(utcInput);
  return d.toLocaleString("zh-CN", {
    timeZone: "Asia/Shanghai",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false
  });
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
 * - 日期+时间：`2006-01-02 15:04:05`（与 admin `parseAccessLogTimeForFilter` 中第二种 layout 一致）
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
