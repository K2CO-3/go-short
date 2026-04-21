import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import { setToken } from "../lib/auth";

export function LoginPage() {
  const nav = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    try {
      const res = await api.login({ username, password });
      setToken(res.access_token);
      nav("/links");
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <main className="auth">
      <h1>登录</h1>
      <form onSubmit={onSubmit} className="card">
        <input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="用户名" />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="密码"
        />
        {error && <p className="error">{error}</p>}
        <button type="submit">登录</button>
      </form>
      <p>
        没有账号？<Link to="/register">注册</Link>
      </p>
    </main>
  );
}
