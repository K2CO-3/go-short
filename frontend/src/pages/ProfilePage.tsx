import { FormEvent, useEffect, useState } from "react";
import { api } from "../lib/api";
import { getToken } from "../lib/auth";
import type { UserProfileResponse } from "../lib/types";

export function ProfilePage() {
  const token = getToken()!;
  const [profile, setProfile] = useState<UserProfileResponse | null>(null);
  const [email, setEmail] = useState("");
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [error, setError] = useState("");
  const [ok, setOk] = useState("");

  async function load() {
    try {
      const res = await api.getProfile(token);
      setProfile(res);
      setEmail(res.email || "");
    } catch (err) {
      setError((err as Error).message);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function updateProfile(e: FormEvent) {
    e.preventDefault();
    setError("");
    setOk("");
    try {
      const res = await api.updateProfile(token, email);
      setProfile(res);
      setOk(res.message || "更新成功");
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function updatePassword(e: FormEvent) {
    e.preventDefault();
    setError("");
    setOk("");
    try {
      const res = await api.updatePassword(token, oldPassword, newPassword);
      setOk(res.message || "密码修改成功");
      setOldPassword("");
      setNewPassword("");
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <section>
      <h2>个人中心</h2>
      {profile && (
        <div className="card">
          <p>用户名：{profile.username}</p>
          <p>角色：{profile.role}</p>
          <p>状态：{profile.status}</p>
        </div>
      )}
      <form className="card" onSubmit={updateProfile}>
        <h3>更新资料</h3>
        <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="邮箱" />
        <button type="submit">保存资料</button>
      </form>
      <form className="card" onSubmit={updatePassword}>
        <h3>修改密码</h3>
        <input
          type="password"
          value={oldPassword}
          onChange={(e) => setOldPassword(e.target.value)}
          placeholder="旧密码"
        />
        <input
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          placeholder="新密码"
        />
        <button type="submit">修改密码</button>
      </form>
      {ok && <p className="ok">{ok}</p>}
      {error && <p className="error">{error}</p>}
    </section>
  );
}
