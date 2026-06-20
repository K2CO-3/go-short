import { FormEvent, useEffect, useState } from "react";
import { api } from "../lib/api";
import {
  dateAndTimeToIso,
  defaultLocalDatetime,
  formatExpiresAtDisplay,
  isExpiresAtInPast,
  splitLocalDatetime
} from "../lib/datetime";
import { getToken } from "../lib/auth";
import type { LinkItem } from "../lib/types";

export function LinksPage() {
  const token = getToken()!;
  const [items, setItems] = useState<LinkItem[]>([]);
  const [url, setUrl] = useState("");
  const [alias, setAlias] = useState("");
  const [shortCode, setShortCode] = useState("");
  const [expiresParts, setExpiresParts] = useState(() => splitLocalDatetime(defaultLocalDatetime(30)));
  const [noExpiry, setNoExpiry] = useState(true);
  const [linkActive, setLinkActive] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [togglingId, setTogglingId] = useState<number | null>(null);

  async function load() {
    setLoading(true);
    try {
      const res = await api.getLinks(token, 1);
      setItems(res.links);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function create(e: FormEvent) {
    e.preventDefault();
    setError("");
    setMessage("");
    if (!noExpiry) {
      const chosen = new Date(`${expiresParts.date}T${expiresParts.time}`);
      if (Number.isNaN(chosen.getTime()) || chosen.getTime() <= Date.now()) {
        setError("过期时间须晚于当前时间，请用日历与时钟重新选择");
        return;
      }
    }
    setCreating(true);
    try {
      const expiresIso = noExpiry ? undefined : dateAndTimeToIso(expiresParts.date, expiresParts.time);
      const res = await api.createLink(token, {
        url,
        alias: alias.trim() || undefined,
        short_code: shortCode.trim() || undefined,
        expires_at: expiresIso,
        status: linkActive
      });
      setMessage(res.message || "创建成功");
      setUrl("");
      setAlias("");
      setShortCode("");
      setExpiresParts(splitLocalDatetime(defaultLocalDatetime(30)));
      setNoExpiry(true);
      setLinkActive(true);
      await load();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setCreating(false);
    }
  }

  async function remove(id: number) {
    try {
      await api.deleteLink(token, id);
      await load();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function setActiveForRow(id: number, active: boolean) {
    setError("");
    setTogglingId(id);
    try {
      await api.setLinkActive(token, id, active);
      await load();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setTogglingId(null);
    }
  }

  function displayShortUrl(code: string): string {
    return `${window.location.origin}/code/${code}`;
  }

  return (
    <section className="page">
      <header className="page-header">
        <h2>短链</h2>
        <p className="page-lead">创建与管理你的短链接</p>
      </header>
      <form className="card card-form" onSubmit={create}>
        <div className="form-row two">
          <input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="原始 URL（必填）"
            required
          />
          <input
            value={alias}
            onChange={(e) => setAlias(e.target.value)}
            placeholder="别名（可选，不填则自动生成）"
          />
        </div>
        <div className="form-row two form-row--shortcode-expiry">
          <div className="link-form-side link-form-side--shortcode">
            <input
              value={shortCode}
              onChange={(e) => setShortCode(e.target.value)}
              placeholder="自定义短码（可选，4–16 位字母数字及 - _）"
            />
          </div>
          <div className="link-form-side link-form-side--expiry">
            <label className="form-check expiry-panel-toggle">
              <input
                type="checkbox"
                checked={noExpiry}
                onChange={(e) => setNoExpiry(e.target.checked)}
              />
              不设过期时间
            </label>
            {noExpiry ? (
              <p className="expiry-panel-note">永不过期：不写入到期日，链接长期有效</p>
            ) : (
              <div className="field-stack expiry-datetime-stack">
                <span className="field-label">过期时间</span>
                <div className="form-datetime-pair">
                  <input
                    type="date"
                    value={expiresParts.date}
                    onChange={(e) =>
                      setExpiresParts((p) => ({
                        ...p,
                        date: e.target.value
                      }))
                    }
                    inputMode="none"
                    autoComplete="off"
                    className="input-calendar"
                    title="点击输入框或右侧图标打开日历"
                  />
                  <input
                    type="time"
                    step={60}
                    value={expiresParts.time}
                    onChange={(e) =>
                      setExpiresParts((p) => ({
                        ...p,
                        time: e.target.value
                      }))
                    }
                    inputMode="none"
                    autoComplete="off"
                    className="input-calendar"
                    title="点击输入框或右侧图标打开时钟"
                  />
                </div>
                <p className="form-hint form-hint-tight">默认过期时间约为 30 天后</p>
              </div>
            )}
          </div>
        </div>
        <label className="form-check">
          <input type="checkbox" checked={linkActive} onChange={(e) => setLinkActive(e.target.checked)} />
          创建后立即启用
        </label>
        <button type="submit" className="btn-primary" disabled={creating}>
          {creating ? "创建中..." : "创建"}
        </button>
        {message && <p className="ok">{message}</p>}
        {error && <p className="error">{error}</p>}
      </form>

      <div className="card table-wrap">
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>别名</th>
                <th>短码</th>
                <th>短链</th>
                <th>原始地址</th>
                <th>过期</th>
                <th>状态</th>
                <th className="th-actions">启停</th>
                <th className="th-actions">删除</th>
              </tr>
            </thead>
            <tbody>
              {!loading && items.length === 0 && (
                <tr>
                  <td colSpan={9} className="empty-cell">
                    暂无短链数据
                  </td>
                </tr>
              )}
              {items.map((i) => (
                <tr key={i.link_id}>
                  <td>{i.link_id}</td>
                  <td className="cell-ellipsis" title={i.alias || ""}>
                    {i.alias || "—"}
                  </td>
                  <td>
                    <code className="code-inline">{i.short_code}</code>
                  </td>
                  <td>
                    <a href={displayShortUrl(i.short_code)} target="_blank" rel="noreferrer" className="link-muted">
                      {displayShortUrl(i.short_code)}
                    </a>
                  </td>
                  <td className="cell-ellipsis" title={i.original_url}>
                    {i.original_url}
                  </td>
                  <td>
                    <div className="cell-expires">
                      {!i.expires_at || String(i.expires_at).trim() === "" ? (
                        <span className="expiry-never">永不过期</span>
                      ) : (
                        <>
                          <span>{formatExpiresAtDisplay(i.expires_at)}</span>
                          {isExpiresAtInPast(i.expires_at) && (
                            <span className="badge badge-warn" title="当前时间已超过该过期时间">
                              已过期
                            </span>
                          )}
                        </>
                      )}
                    </div>
                  </td>
                  <td>
                    {i.is_active === false ? (
                      <span className="badge">停用</span>
                    ) : (
                      <span className="badge badge-ok">启用</span>
                    )}
                  </td>
                  <td className="td-actions">
                    {i.is_active === false ? (
                      <button
                        type="button"
                        className="btn-primary btn-sm"
                        disabled={togglingId === i.link_id}
                        onClick={() => void setActiveForRow(i.link_id, true)}
                      >
                        启用
                      </button>
                    ) : (
                      <button
                        type="button"
                        className="btn-danger btn-sm"
                        disabled={togglingId === i.link_id}
                        onClick={() => void setActiveForRow(i.link_id, false)}
                      >
                        停用
                      </button>
                    )}
                  </td>
                  <td className="td-actions">
                    <button
                      type="button"
                      className="btn-danger btn-sm"
                      onClick={() => void remove(i.link_id)}
                    >
                      删除
                    </button>
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
