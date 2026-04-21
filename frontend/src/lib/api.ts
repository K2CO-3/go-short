import type {
  AdminListAccessLogsResponse,
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
  adminCreateUser(
    token: string,
    payload: { username: string; password: string; email: string; role: string; status: string }
  ) {
    return request<BaseResponse & { user_id?: string }>(
      "/admin/createUser",
      "POST",
      token,
      payload
    );
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
