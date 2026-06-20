import { FormEvent, useEffect, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import { setToken } from "../lib/auth";

export function LoginPage() {
  const nav = useNavigate();
  const location = useLocation();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [notice] = useState(() => {
    const st = location.state as { fromRegister?: boolean } | null;
    return st?.fromRegister ? "注册成功，请登录" : "";
  });

  useEffect(() => {
    const st = location.state as { fromRegister?: boolean } | null;
    if (st?.fromRegister) {
      nav(location.pathname, { replace: true, state: {} });
    }
  }, [location, nav]);

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
      <div className="auth-header">
        <h1>登录</h1>
        <p className="auth-sub">使用账号进入控制台</p>
      </div>
      <form onSubmit={onSubmit} className="card">
        {notice && <p className="ok">{notice}</p>}
        <input
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="用户名（必填）"
          required
        />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="密码（必填）"
          required
        />
        {error && <p className="error">{error}</p>}
        <button type="submit" className="btn-primary">
          登录
        </button>
      </form>
      <p className="auth-footer">
        没有账号？<Link to="/register">注册</Link>
      </p>
    </main>
  );
}
