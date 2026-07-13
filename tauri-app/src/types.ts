// Mirror of the Rust structs returned by the `list_projects` command.

export interface EntryPoint {
  name: string;
  type: string;
  description?: string | null;
  url?: string | null;
}

export interface Worktree {
  path: string;
  name: string;
  description?: string | null;
  special: boolean;
  registered_at?: string | null;
  updated_at?: string | null;
  entry_points: EntryPoint[];
}

export interface Project {
  project_path: string;
  display_name: string;
  worktrees: Worktree[];
}

// Mirror of the Rust `StatusCheck` returned by `check_dependencies`.
export interface StatusCheck {
  id: string;
  label: string;
  ok: boolean;
  detail: string;
}

// Mirror of the Rust `Opener` returned by `list_openers`.
export interface Opener {
  id: string;
  label: string;
  available: boolean;
}
