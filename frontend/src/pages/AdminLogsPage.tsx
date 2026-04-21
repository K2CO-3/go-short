import { FormEvent, useState } from "react";
import { api } from "../lib/api";
import { getToken } from "../lib/auth";
import type { AdminAccessLogItem } from "../lib/types";

export function AdminLogsPage() {
  const token = getToken()!;
  const [logs, setLogs] = useState<AdminAccessLogItem[]>([]);
  const [error, setError] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endTime, setEndTime] = useState("");
  const [ipAddress, setIpAddress] = useState("");
  const [shortCode, setShortCode] = useState("");
  const [originalURL, setOriginalURL] = useState("");
  const [limit, setLimit] = useState("100");

  async function query(e: FormEvent) {
    e.preventDefault();
    setError("");
    try {
      const res = await api.adminGetLogs(token, {
        start_time: startTime || undefined,
        end_time: endTime || undefined,
        ip_address: ipAddress || undefined,
        short_code: shortCode || undefined,
        original_url: originalURL || undefined,
        limit: limit ? Number(limit) : undefined
      });
      setLogs(res.logs);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <section>
      <h2>管理员 - 访问日志</h2>
      <p className="hint">
        时间格式支持：<code>2026-04-01T00:00:00Z</code>、<code>2026-04-01 00:00:00</code>、
        <code>2026-04-01</code>
      </p>
      <form className="card grid" onSubmit={query}>
        <input
          value={startTime}
          onChange={(e) => setStartTime(e.target.value)}
          placeholder="start_time (如 2026-04-01T00:00:00Z)"
        />
        <input
          value={endTime}
          onChange={(e) => setEndTime(e.target.value)}
          placeholder="end_time (如 2026-04-20T23:59:59Z)"
        />
        <input value={ipAddress} onChange={(e) => setIpAddress(e.target.value)} placeholder="ip_address" />
        <input value={shortCode} onChange={(e) => setShortCode(e.target.value)} placeholder="short_code" />
        <input
          value={originalURL}
          onChange={(e) => setOriginalURL(e.target.value)}
          placeholder="original_url"
        />
        <input value={limit} onChange={(e) => setLimit(e.target.value)} placeholder="limit" />
        <button type="submit">查询</button>
      </form>
      {error && <p className="error">{error}</p>}
      <div className="card">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>LinkID</th>
              <th>短码</th>
              <th>IP</th>
              <th>UserAgent</th>
              <th>访问时间</th>
            </tr>
          </thead>
          <tbody>
            {logs.map((l) => (
              <tr key={l.id}>
                <td>{l.id}</td>
                <td>{l.link_id}</td>
                <td>{l.short_code}</td>
                <td>{l.ip_address}</td>
                <td>{l.user_agent}</td>
                <td>{l.visited_at}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
