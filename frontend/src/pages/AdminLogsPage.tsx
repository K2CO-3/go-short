import { FormEvent, useState } from "react";
import { api } from "../lib/api";
import {
  addDays,
  buildAccessLogTimeParam,
  formatInputDate,
  formatUtcToShanghai,
  isAccessLogRangeReversed
} from "../lib/datetime";
import { getToken } from "../lib/auth";
import type { AdminAccessLogItem } from "../lib/types";

const LIMIT_MAX = 2000;
const LIMIT_MIN = 1;

export function AdminLogsPage() {
  const token = getToken()!;
  const [logs, setLogs] = useState<AdminAccessLogItem[]>([]);
  const [error, setError] = useState("");

  const [startDate, setStartDate] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endDate, setEndDate] = useState("");
  const [endTime, setEndTime] = useState("");

  const [ipAddress, setIpAddress] = useState("");
  const [shortCode, setShortCode] = useState("");
  const [originalURL, setOriginalURL] = useState("");
  const [limit, setLimit] = useState("100");
  const [loading, setLoading] = useState(false);
  /** 快捷切换时递增，强制时间输入框重新挂载，避免部分浏览器在清空 value 后仍显示旧时刻 */
  const [presetKey, setPresetKey] = useState(0);

  function applyPreset(kind: "today" | "7d" | "30d" | "clear") {
    setError("");
    if (kind === "clear") {
      setStartDate("");
      setStartTime("");
      setEndDate("");
      setEndTime("");
      setPresetKey((k) => k + 1);
      return;
    }
    const today = new Date();
    const todayStr = formatInputDate(today);
    const dayStart = "00:00";
    const dayEnd = "23:59";
    if (kind === "today") {
      setStartDate(todayStr);
      setEndDate(todayStr);
      setStartTime(dayStart);
      setEndTime(dayEnd);
      setPresetKey((k) => k + 1);
      return;
    }
    const days = kind === "7d" ? -6 : -29;
    const start = formatInputDate(addDays(today, days));
    setStartDate(start);
    setEndDate(todayStr);
    setStartTime(dayStart);
    setEndTime(dayEnd);
    setPresetKey((k) => k + 1);
  }

  async function query(e: FormEvent) {
    e.preventDefault();
    setError("");

    const a = buildAccessLogTimeParam(startDate, startTime);
    if (a.error) {
      setError(a.error);
      return;
    }
    const b = buildAccessLogTimeParam(endDate, endTime);
    if (b.error) {
      setError(b.error);
      return;
    }

    if (isAccessLogRangeReversed(a.value, b.value)) {
      setError("开始时间不能晚于结束时间");
      return;
    }

    const limitTrim = limit.trim();
    if (limitTrim !== "") {
      const n = Number(limitTrim);
      if (!Number.isInteger(n) || n < LIMIT_MIN || n > LIMIT_MAX) {
        setError(`请输入 ${LIMIT_MIN}～${LIMIT_MAX} 之间的整数`);
        return;
      }
    }

    setLoading(true);
    try {
      const res = await api.adminGetLogs(token, {
        start_time: a.value,
        end_time: b.value,
        ip_address: ipAddress.trim() || undefined,
        short_code: shortCode.trim() || undefined,
        original_url: originalURL.trim() || undefined,
        limit: limitTrim === "" ? undefined : Number(limitTrim)
      });
      setLogs(res.logs);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <section className="page">
      <header className="page-header">
        <h2>访问日志</h2>
        <p className="page-lead">按条件筛选访问记录</p>
      </header>
      <form className="card card-form" onSubmit={query}>
        <div className="log-presets" role="group" aria-label="快捷选择时间范围">
          <span className="log-presets-label">快捷</span>
          <button type="button" className="btn-ghost btn-sm" onClick={() => applyPreset("today")}>
            今天
          </button>
          <button type="button" className="btn-ghost btn-sm" onClick={() => applyPreset("7d")}>
            近 7 天
          </button>
          <button type="button" className="btn-ghost btn-sm" onClick={() => applyPreset("30d")}>
            近 30 天
          </button>
          <button type="button" className="btn-ghost btn-sm" onClick={() => applyPreset("clear")}>
            清空时间
          </button>
        </div>
        <div className="form-grid-logs form-grid-logs--wide">
          <div className="field-stack form-log-time-block">
            <span className="field-label">从哪一天起</span>
            <div className="form-log-time-pair">
              <input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                aria-label="开始日期"
                inputMode="none"
                autoComplete="off"
                className="input-calendar"
                title="点击打开日历"
              />
              <input
                key={`log-start-time-${presetKey}`}
                type="time"
                step={60}
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
                aria-label="开始时间（选填，需配合日期）"
                inputMode="none"
                autoComplete="off"
                className="input-calendar"
                title="点击打开时钟"
              />
            </div>
          </div>
          <div className="field-stack form-log-time-block">
            <span className="field-label">查到哪一天</span>
            <div className="form-log-time-pair">
              <input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                aria-label="结束日期"
                inputMode="none"
                autoComplete="off"
                className="input-calendar"
                title="点击打开日历"
              />
              <input
                key={`log-end-time-${presetKey}`}
                type="time"
                step={60}
                value={endTime}
                onChange={(e) => setEndTime(e.target.value)}
                aria-label="结束时间（选填，需配合日期）"
                inputMode="none"
                autoComplete="off"
                className="input-calendar"
                title="点击打开时钟"
              />
            </div>
          </div>
          <div className="field-stack">
            <span className="field-label">最多多少条</span>
            <input
              value={limit}
              onChange={(e) => setLimit(e.target.value)}
              inputMode="numeric"
              placeholder="默认 100"
              aria-label="返回条数上限"
            />
          </div>
        </div>
        <div className="form-grid-logs">
          <input
            value={ipAddress}
            onChange={(e) => setIpAddress(e.target.value)}
            placeholder="IP 地址（选填）"
          />
          <input value={shortCode} onChange={(e) => setShortCode(e.target.value)} placeholder="短码（选填）" />
          <input
            value={originalURL}
            onChange={(e) => setOriginalURL(e.target.value)}
            placeholder="原始链接（选填）"
          />
        </div>
        <button type="submit" className="btn-primary" disabled={loading}>
          {loading ? "查询中..." : "查询"}
        </button>
      </form>
      {error && <p className="error">{error}</p>}
      <div className="card table-wrap">
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>编号</th>
                <th>链接 ID</th>
                <th>短码</th>
                <th>IP</th>
                <th>访问环境</th>
                <th title="数据库存 UTC，此处按东八区显示">访问时间 (UTC+8)</th>
              </tr>
            </thead>
            <tbody>
              {!loading && logs.length === 0 && (
                <tr>
                  <td colSpan={6} className="empty-cell">
                    暂无日志数据
                  </td>
                </tr>
              )}
              {logs.map((l) => (
                <tr key={l.id}>
                  <td>{l.id}</td>
                  <td>{l.link_id}</td>
                  <td>
                    <code className="code-inline">{l.short_code}</code>
                  </td>
                  <td>{l.ip_address}</td>
                  <td className="cell-muted">{l.user_agent}</td>
                  <td className="cell-nowrap" title={l.visited_at}>
                    {formatUtcToShanghai(l.visited_at)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}
