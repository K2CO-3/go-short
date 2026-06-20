import { clearToken } from "./auth";
import type {
  AdminCreateUserRequest,
  AdminListAccessLogsResponse,
  AdminListLinksResponse,
  AdminListUsersResponse,
  ApiError,
  BaseResponse,
  CreateLinkRequest,
  LinkResponse,
  ListLinksResponse,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  UserProfileResponse
} from "./types";

const API_BASE =
  (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim() || "/api/v1";

type Method = "GET" | "POST" | "PUT" | "DELETE";

async function request<T>(
  path: string,
  method: Method,
  token?: string | null,
  body?: unknown
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json"
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body)
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    if (res.status === 401 && token) {
      clearToken();
      window.location.assign("/login");
    }
    const err = data as ApiError;
    throw new Error(err.message || err.error || `HTTP ${res.status}`);
  }
  return data as T;
}

export const api = {
  register(payload: RegisterRequest) {
    return request<RegisterResponse>("/auth/register", "POST", null, payload);
  },
  login(payload: LoginRequest) {
    return request<LoginResponse>("/auth/login", "POST", null, payload);
  },
  createLink(token: string, payload: CreateLinkRequest) {
    return request<LinkResponse>("/links", "POST", token, payload);
  },
  getLinks(token: string, page = 1) {
    return request<ListLinksResponse>(`/links?page=${page}`, "GET", token);
  },
  getLinksByAlias(token: string, alias: string, page = 1) {
    const q = `?alias=${encodeURIComponent(alias)}&page=${page}`;
    return request<ListLinksResponse>(`/links/GetLinksByAlias${q}`, "GET", token);
  },
  deleteLink(token: string, id: number) {
    return request<BaseResponse>(`/links/${id}`, "DELETE", token);
  },
  /** 启用/禁用短链：本人可管理自己的短链；角色为 admin 时也可管理他人短链 */
  setLinkActive(token: string, id: number, active: boolean) {
    const path = active ? "activate" : "unactivate";
    return request<LinkResponse>(`/links/${id}/${path}`, "PUT", token);
  },
  adminGetLinks(token: string, page = 1, size = 20) {
    return request<AdminListLinksResponse>(`/admin/getLinkList?page=${page}&size=${size}`, "GET", token);
  },
  getProfile(token: string) {
    return request<UserProfileResponse>("/user/profile", "GET", token);
  },
  updateProfile(token: string, email?: string) {
    return request<UserProfileResponse>("/user/profile", "PUT", token, { email });
  },
  updatePassword(token: string, old_password: string, new_password: string) {
    return request<BaseResponse>("/user/password", "PUT", token, {
      old_password,
      new_password
    });
  },
  adminGetUsers(token: string, page = 1, size = 10) {
    return request<AdminListUsersResponse>(
      `/admin/getUserList?page=${page}&size=${size}`,
      "GET",
      token
    );
  },
  adminCreateUser(token: string, payload: AdminCreateUserRequest) {
    const body: Record<string, unknown> = {
      username: payload.username,
      password: payload.password,
      email: payload.email,
      role: payload.role,
      status: payload.status
    };
    if (payload.realName?.trim()) body.realName = payload.realName.trim();
    if (payload.phone?.trim()) body.phone = payload.phone.trim();
    if (payload.remark?.trim()) body.remark = payload.remark.trim();
    return request<BaseResponse & { user_id?: string }>("/admin/createUser", "POST", token, body);
  },
  adminSetUserActive(token: string, userID: string, active: boolean) {
    const endpoint = active ? "activateUser" : "unactivateUser";
    return request<BaseResponse>(`/admin/${endpoint}/${userID}`, "PUT", token);
  },
  adminGetLogs(
    token: string,
    query: {
      start_time?: string;
      end_time?: string;
      ip_address?: string;
      short_code?: string;
      original_url?: string;
      limit?: number;
    }
  ) {
    const params = new URLSearchParams();
    Object.entries(query).forEach(([k, v]) => {
      if (v !== undefined && v !== null && `${v}`.trim() !== "") params.set(k, `${v}`);
    });
    const suffix = params.toString() ? `?${params.toString()}` : "";
    return request<AdminListAccessLogsResponse>(`/admin/logs${suffix}`, "GET", token);
  }
};
