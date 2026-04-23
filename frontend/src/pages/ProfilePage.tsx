import { FormEvent, useEffect, useState } from "react";
import { api } from "../lib/api";
import { getToken } from "../lib/auth";
import { formatRoleLabel } from "../lib/role";
import type { UserProfileResponse } from "../lib/types";

export function ProfilePage() {
  const token = getToken()!;
  const [profile, setProfile] = useState<UserProfileResponse | null>(null);
  const [email, setEmail] = useState("");
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [error, setError] = useState("");
  const [ok, setOk] = useState("");
  const [savingProfile, setSavingProfile] = useState(false);
  const [savingPassword, setSavingPassword] = useState(false);

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
    setSavingProfile(true);
    try {
      const res = await api.updateProfile(token, email);
      setProfile(res);
      setOk(res.message || "更新成功");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setSavingProfile(false);
    }
  }

  async function updatePassword(e: FormEvent) {
    e.preventDefault();
    setError("");
    setOk("");
    setSavingPassword(true);
    try {
      const res = await api.updatePassword(token, oldPassword, newPassword);
      setOk(res.message || "密码修改成功");
      setOldPassword("");
      setNewPassword("");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setSavingPassword(false);
    }
  }

  return (
    <section className="page">
      <header className="page-header">
        <h2>个人中心</h2>
        <p className="page-lead">账号信息与安全设置</p>
      </header>
      {profile && (
        <div className="card card-profile">
          <dl className="stat-list">
            <div>
              <dt>用户名</dt>
              <dd>{profile.username}</dd>
            </div>
            <div>
              <dt>角色</dt>
              <dd>{formatRoleLabel(profile.role)}</dd>
            </div>
            <div>
              <dt>状态</dt>
              <dd>{profile.status}</dd>
            </div>
          </dl>
        </div>
      )}
      <form className="card card-form" onSubmit={updateProfile}>
        <h3 className="card-title">资料</h3>
        <input
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="邮箱（可选）"
          type="email"
        />
        <button type="submit" className="btn-primary" disabled={savingProfile}>
          {savingProfile ? "保存中..." : "保存"}
        </button>
      </form>
      <form className="card card-form" onSubmit={updatePassword}>
        <h3 className="card-title">密码</h3>
        <div className="form-row two">
          <input
            type="password"
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            placeholder="当前密码"
          />
          <input
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            placeholder="新密码"
          />
        </div>
        <button type="submit" className="btn-primary" disabled={savingPassword}>
          {savingPassword ? "更新中..." : "更新密码"}
        </button>
      </form>
      {ok && <p className="ok">{ok}</p>}
      {error && <p className="error">{error}</p>}
    </section>
  );
}
