import { FormEvent, useEffect, useState } from "react";
import { api } from "../lib/api";
import { getToken } from "../lib/auth";
import { formatRoleLabel } from "../lib/role";
import type { AdminUserItem } from "../lib/types";

export function AdminUsersPage() {
  const token = getToken()!;
  const [users, setUsers] = useState<AdminUserItem[]>([]);
  const [error, setError] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"user" | "admin">("user");
  const [status, setStatus] = useState<"active" | "inactive">("active");
  const [realName, setRealName] = useState("");
  const [phone, setPhone] = useState("");
  const [remark, setRemark] = useState("");
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);

  async function load() {
    setLoading(true);
    try {
      const res = await api.adminGetUsers(token, 1, 20);
      setUsers(res.users);
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
    setCreating(true);
    try {
      await api.adminCreateUser(token, {
        username,
        email,
        password,
        role,
        status,
        realName: realName.trim() || undefined,
        phone: phone.trim() || undefined,
        remark: remark.trim() || undefined
      });
      setUsername("");
      setEmail("");
      setPassword("");
      setRole("user");
      setStatus("active");
      setRealName("");
      setPhone("");
      setRemark("");
      await load();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setCreating(false);
    }
  }

  async function toggle(user: AdminUserItem) {
    try {
      await api.adminSetUserActive(token, user.user_id, !user.is_active);
      await load();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <section className="page">
      <header className="page-header">
        <h2>用户</h2>
        <p className="page-lead">管理员：创建与启用/禁用用户</p>
      </header>
      <form className="card card-form" onSubmit={create}>
        <h3 className="card-title">新建用户</h3>
        <div className="form-row three">
          <input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="用户名（必填）" required />
          <input
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="邮箱（必填）"
            type="email"
            required
          />
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="密码（必填）"
            required
            minLength={6}
          />
        </div>
        <div className="form-row two">
          <div className="field-stack">
            <span className="field-label">角色</span>
            <select className="input-select" value={role} onChange={(e) => setRole(e.target.value as "user" | "admin")}>
              <option value="user">普通用户</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          <div className="field-stack">
            <span className="field-label">状态</span>
            <select
              className="input-select"
              value={status}
              onChange={(e) => setStatus(e.target.value as "active" | "inactive")}
            >
              <option value="active">active</option>
              <option value="inactive">inactive</option>
            </select>
          </div>
        </div>
        <div className="form-row three">
          <input value={realName} onChange={(e) => setRealName(e.target.value)} placeholder="真实姓名（可选）" />
          <input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="手机号（可选）" />
          <input value={remark} onChange={(e) => setRemark(e.target.value)} placeholder="备注（可选）" />
        </div>
        <button type="submit" className="btn-primary" disabled={creating}>
          {creating ? "创建中..." : "创建"}
        </button>
      </form>
      {error && <p className="error">{error}</p>}
      <div className="card table-wrap">
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>用户名</th>
                <th>邮箱</th>
                <th>角色</th>
                <th>状态</th>
                <th className="th-actions" />
              </tr>
            </thead>
            <tbody>
              {!loading && users.length === 0 && (
                <tr>
                  <td colSpan={6} className="empty-cell">
                    暂无用户数据
                  </td>
                </tr>
              )}
              {users.map((u) => (
                <tr key={u.user_id}>
                  <td>{u.user_id}</td>
                  <td>{u.username}</td>
                  <td>{u.email}</td>
                  <td>
                    <span
                      className={
                        u.role === "admin" ? "badge badge-role-admin" : "badge badge-role-user"
                      }
                    >
                      {formatRoleLabel(u.role)}
                    </span>
                  </td>
                  <td>
                    <span className={u.is_active ? "badge badge-ok" : "badge"}>
                      {u.is_active ? "active" : "inactive"}
                    </span>
                  </td>
                  <td className="td-actions">
                    <button
                      type="button"
                      className={u.is_active ? "btn-danger btn-sm" : "btn-primary btn-sm"}
                      onClick={() => void toggle(u)}
                    >
                      {u.is_active ? "禁用" : "启用"}
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
