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
      nav("/login");
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <main className="auth">
      <h1>注册</h1>
      <form onSubmit={onSubmit} className="card">
        <input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="用户名" />
        <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="邮箱" />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="密码"
        />
        {error && <p className="error">{error}</p>}
        <button type="submit">注册</button>
      </form>
      <p>
        已有账号？<Link to="/login">登录</Link>
      </p>
    </main>
  );
}
