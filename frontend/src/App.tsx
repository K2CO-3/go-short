import { NavLink, Navigate, Route, Routes, useNavigate } from "react-router-dom";
import { clearToken, getToken } from "./lib/auth";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { LinksPage } from "./pages/LinksPage";
import { ProfilePage } from "./pages/ProfilePage";
import { AdminUsersPage } from "./pages/AdminUsersPage";
import { AdminLogsPage } from "./pages/AdminLogsPage";

function Protected({ children }: { children: JSX.Element }) {
  const token = getToken();
  if (!token) return <Navigate to="/login" replace />;
  return children;
}

function Nav() {
  const navigate = useNavigate();
  return (
    <header className="nav">
      <div className="brand">GoShort Console</div>
      <nav>
        <NavLink to="/links">短链</NavLink>
        <NavLink to="/profile">个人</NavLink>
        <NavLink to="/admin/users">管理用户</NavLink>
        <NavLink to="/admin/logs">管理日志</NavLink>
      </nav>
      <button
        onClick={() => {
          clearToken();
          navigate("/login");
        }}
      >
        退出
      </button>
    </header>
  );
}

export default function App() {
  return (
    <div>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route
          path="*"
          element={
            <Protected>
              <div>
                <Nav />
                <main className="container">
                  <Routes>
                    <Route path="/" element={<Navigate to="/links" replace />} />
                    <Route path="/links" element={<LinksPage />} />
                    <Route path="/profile" element={<ProfilePage />} />
                    <Route path="/admin/users" element={<AdminUsersPage />} />
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
