import { FormEvent, useEffect, useState } from "react";
import { api } from "../lib/api";
import { getToken } from "../lib/auth";
import type { AdminUserItem } from "../lib/types";

export function AdminUsersPage() {
  const token = getToken()!;
  const [users, setUsers] = useState<AdminUserItem[]>([]);
  const [error, setError] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  async function load() {
    try {
      const res = await api.adminGetUsers(token, 1, 20);
      setUsers(res.users);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function create(e: FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await api.adminCreateUser(token, { username, email, password, role: "user", status: "active" });
      setUsername("");
      setEmail("");
      setPassword("");
      await load();
    } catch (err) {
      setError((err as Error).message);
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
    <section>
      <h2>管理员 - 用户管理</h2>
      <form className="card" onSubmit={create}>
        <h3>创建用户</h3>
        <input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="用户名" />
        <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="邮箱" />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="密码"
        />
        <button type="submit">创建</button>
      </form>
      {error && <p className="error">{error}</p>}
      <div className="card">
        <table>
          <thead>
            <tr>
              <th>UserID</th>
              <th>用户名</th>
              <th>邮箱</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.user_id}>
                <td>{u.user_id}</td>
                <td>{u.username}</td>
                <td>{u.email}</td>
                <td>{u.is_active ? "active" : "inactive"}</td>
                <td>
                  <button onClick={() => void toggle(u)}>{u.is_active ? "禁用" : "启用"}</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
