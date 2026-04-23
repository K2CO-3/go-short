import { NavLink, Navigate, Route, Routes, useNavigate } from "react-router-dom";
import { clearToken, getToken } from "./lib/auth";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { LinksPage } from "./pages/LinksPage";
import { ProfilePage } from "./pages/ProfilePage";
import { AdminUsersPage } from "./pages/AdminUsersPage";
import { AdminLogsPage } from "./pages/AdminLogsPage";
import { AdminLinksPage } from "./pages/AdminLinksPage";

function Protected({ children }: { children: JSX.Element }) {
  const token = getToken();
  if (!token) return <Navigate to="/login" replace />;
  return children;
}

function Nav() {
  const navigate = useNavigate();
  return (
    <header className="nav">
      <div className="nav-inner">
        <div className="brand">GoShort</div>
        <nav className="nav-links">
          <NavLink to="/links" end>
            短链
          </NavLink>
          <NavLink to="/profile">个人</NavLink>
          <NavLink to="/admin/users">用户</NavLink>
          <NavLink to="/admin/links">全站短链</NavLink>
          <NavLink to="/admin/logs">日志</NavLink>
        </nav>
        <button
          type="button"
          className="btn-ghost"
          onClick={() => {
            clearToken();
            navigate("/login");
          }}
        >
          退出
        </button>
      </div>
    </header>
  );
}

export default function App() {
  return (
    <div className="app-root">
      <Routes>
        <Route
          path="/login"
          element={
            <div className="auth-shell">
              <LoginPage />
            </div>
          }
        />
        <Route
          path="/register"
          element={
            <div className="auth-shell">
              <RegisterPage />
            </div>
          }
        />
        <Route
          path="*"
          element={
            <Protected>
              <div className="app-shell">
                <Nav />
                <main className="container">
                  <Routes>
                    <Route path="/" element={<Navigate to="/links" replace />} />
                    <Route path="/links" element={<LinksPage />} />
                    <Route path="/profile" element={<ProfilePage />} />
                    <Route path="/admin/users" element={<AdminUsersPage />} />
                    <Route path="/admin/links" element={<AdminLinksPage />} />
                    <Route path="/admin/logs" element={<AdminLogsPage />} />
                  </Routes>
                </main>
              </div>
            </Protected>
          }
        />
      </Routes>
    </div>
  );
}
