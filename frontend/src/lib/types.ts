export type BaseResponse = {
  success: boolean;
  message?: string;
};

export type ApiError = {
  success?: false;
  code?: string;
  message?: string;
  details?: string;
  error?: string;
};

export type RegisterRequest = {
  username: string;
  password: string;
  email: string;
};

export type RegisterResponse = BaseResponse & {
  user_id: string;
  username: string;
  email: string;
};

export type LoginRequest = {
  username: string;
  password: string;
};

export type LoginResponse = BaseResponse & {
  access_token: string;
  role: string;
  expires_in: number;
  token_type: string;
  refresh_token?: string;
};

export type CreateLinkRequest = {
  url: string;
  alias?: string | null;
  expires_at?: string | null;
  status?: boolean | null;
  short_code?: string | null;
};

/** 对应 internal/handler/admin/request.go CreateUserRequest */
export type AdminCreateUserRequest = {
  username: string;
  password: string;
  email: string;
  role: string;
  status: string;
  realName?: string;
  phone?: string;
  remark?: string;
};

export type LinkItem = {
  link_id: number;
  short_code: string;
  original_url: string;
  short_url: string;
  alias: string;
  is_active?: boolean;
  expires_at?: string | null;
  created_at: string;
};

export type LinkResponse = BaseResponse & Partial<LinkItem>;

export type ListLinksResponse = BaseResponse & {
  links: LinkItem[];
  total: number;
  page: number;
  limit: number;
};

export type AdminLinkItem = {
  link_id: string;
  short_code: string;
  original_url: string;
  is_active?: boolean;
  created_by?: string;
  created_at?: string;
};

export type AdminListLinksResponse = BaseResponse & {
  links: AdminLinkItem[];
  total: number;
  page: number;
  limit: number;
};

export type UserProfileResponse = BaseResponse & {
  user_id: string;
  username: string;
  email?: string | null;
  role: string;
  status: string;
  link_count: number;
  created_at?: string | null;
  updated_at?: string | null;
};

export type AdminUserItem = {
  user_id: string;
  username: string;
  email: string;
  role?: string;
  is_active?: boolean;
  created_at?: string;
  updated_at?: string;
};

export type AdminListUsersResponse = BaseResponse & {
  users: AdminUserItem[];
  total: number;
  page: number;
  limit: number;
};

export type AdminAccessLogItem = {
  id: number;
  link_id: number;
  short_code: string;
  ip_address: string;
  user_agent: string;
  visited_at: string;
};

export type AdminListAccessLogsResponse = BaseResponse & {
  logs: AdminAccessLogItem[];
  total: number;
};
