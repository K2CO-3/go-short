import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../lib/api";

export function RegisterPage() {
  const nav = useNavigate();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await api.register({ username, email, password });
      nav("/login", { replace: true, state: { fromRegister: true } });
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <main className="auth">
      <div className="auth-header">
        <h1>注册</h1>
        <p className="auth-sub">创建新账号以使用短链服务</p>
      </div>
      <form onSubmit={onSubmit} className="card">
        <input
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="用户名（必填）"
          required
        />
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
          placeholder="密码（必填，至少 6 位）"
          required
          minLength={6}
        />
        {error && <p className="error">{error}</p>}
        <button type="submit" className="btn-primary">
          注册
        </button>
      </form>
      <p className="auth-footer">
        已有账号？<Link to="/login">登录</Link>
      </p>
    </main>
  );
}
