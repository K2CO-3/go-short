/** 与后端 role 字段对应，用于界面展示 */
export function formatRoleLabel(role: string | undefined | null): string {
  if (!role) return "—";
  switch (role) {
    case "admin":
      return "管理员";
    case "user":
      return "普通用户";
    default:
      return role;
  }
}
