import { useEffect, useState } from "react";
import { api } from "../lib/api";
import { formatExpiresAtDisplay, isExpiresAtInPast } from "../lib/datetime";
import { getToken } from "../lib/auth";
import type { AdminLinkItem } from "../lib/types";

export function AdminLinksPage() {
  const token = getToken()!;
  const [items, setItems] = useState<AdminLinkItem[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [togglingId, setTogglingId] = useState<string | null>(null);

  async function load() {
    setLoading(true);
    setError("");
    try {
      const res = await api.adminGetLinks(token, 1, 30);
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

  async function toggleRow(row: AdminLinkItem) {
    const id = parseInt(row.link_id, 10);
    if (Number.isNaN(id)) return;
    const currentlyActive = row.is_active !== false;
    setTogglingId(row.link_id);
    setError("");
    try {
      await api.setLinkActive(token, id, !currentlyActive);
      await load();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setTogglingId(null);
    }
  }

  return (
    <section className="page">
      <header className="page-header">
        <h2>全站短链</h2>
        <p className="page-lead">管理员可启用或禁用任意用户的短码</p>
      </header>
      {error && <p className="error">{error}</p>}
      <div className="card table-wrap">
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>短码</th>
                <th>创建者</th>
                <th>原始地址</th>
                <th>过期</th>
                <th>状态</th>
                <th className="th-actions" />
              </tr>
            </thead>
            <tbody>
              {!loading && items.length === 0 && !error && (
                <tr>
                  <td colSpan={7} className="empty-cell">
                    暂无数据
                  </td>
                </tr>
              )}
              {items.map((i) => (
                <tr key={i.link_id}>
                  <td>{i.link_id}</td>
                  <td>
                    <code className="code-inline">{i.short_code}</code>
                  </td>
                  <td className="cell-ellipsis" title={i.creator_username || ""}>
                    {i.creator_username || "—"}
                  </td>
                  <td className="cell-ellipsis" title={i.original_url}>
                    {i.original_url}
                  </td>
                  <td>
                    <div className="cell-expires">
                      <span>{formatExpiresAtDisplay(i.expires_at)}</span>
                      {isExpiresAtInPast(i.expires_at) && (
                        <span className="badge badge-warn" title="当前时间已超过该过期时间">
                          已过期
                        </span>
                      )}
                    </div>
                  </td>
                  <td>
                    {i.is_active === false ? <span className="badge">停用</span> : <span className="badge badge-ok">启用</span>}
                  </td>
                  <td className="td-actions">
                    <button
                      type="button"
                      className={i.is_active !== false ? "btn-danger btn-sm" : "btn-primary btn-sm"}
                      disabled={togglingId === i.link_id}
                      onClick={() => void toggleRow(i)}
                    >
                      {i.is_active !== false ? "停用" : "启用"}
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
