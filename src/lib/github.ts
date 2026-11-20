import { isReport, type JagaReport } from "./report";

const API = "https://api.github.com";
const RAW = "https://raw.githubusercontent.com";
const CACHE_SECONDS = 600;

export type GithubRepo = {
  name: string;
  description: string | null;
  html_url: string;
  updated_at: string;
  fork: boolean;
  stargazers_count: number;
  owner: { login: string };
};

function headers(): HeadersInit {
  const h: Record<string, string> = {
    Accept: "application/vnd.github+json",
    "User-Agent": "jaga-card",
    "X-GitHub-Api-Version": "2022-11-28",
  };
  const token = process.env.GITHUB_TOKEN;
  if (token) h.Authorization = `Bearer ${token}`;
  return h;
}

export async function fetchJSON<T>(path: string): Promise<T | null> {
  try {
    const res = await fetch(`${API}${path}`, {
      headers: headers(),
      next: { revalidate: CACHE_SECONDS },
    });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch {
    return null;
  }
}

export async function getUser(username: string): Promise<{ login: string; name: string | null; avatar_url: string } | null> {
  return fetchJSON(`/users/${encodeURIComponent(username)}`);
}

export async function getUserRepos(username: string): Promise<GithubRepo[]> {
  const data = await fetchJSON<GithubRepo[]>(
    `/users/${encodeURIComponent(username)}/repos?per_page=100&sort=updated&type=owner`,
  );
  return Array.isArray(data) ? data : [];
}

export async function getRepo(owner: string, repo: string): Promise<GithubRepo | null> {
  return fetchJSON<GithubRepo>(`/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`);
}

export async function getReport(owner: string, repo: string): Promise<JagaReport | null> {
  try {
    const url = `${RAW}/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/HEAD/jaga-report.json`;
    const res = await fetch(url, {
      headers: { "User-Agent": "jaga-card" },
      next: { revalidate: CACHE_SECONDS },
    });
    if (!res.ok) return null;
    const data: unknown = await res.json();
    return isReport(data) ? data : null;
  } catch {
    return null;
  }
}

export async function mapWithLimit<T, R>(items: T[], limit: number, fn: (item: T) => Promise<R>): Promise<R[]> {
  const out = new Array<R>(items.length);
  let cursor = 0;
  const workers = Array.from({ length: Math.min(limit, items.length) }, async () => {
    while (cursor < items.length) {
      const index = cursor++;
      out[index] = await fn(items[index]);
    }
  });
  await Promise.all(workers);
  return out;
}
