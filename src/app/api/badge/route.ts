import type { NextRequest } from "next/server";
import { getReport, mapWithLimit, getUserRepos } from "@/lib/github";
import { averageScore } from "@/lib/report";
import { badgeSVG } from "@/lib/svg";

export async function GET(request: NextRequest) {
  const q = request.nextUrl.searchParams;
  const owner = q.get("owner");
  const repo = q.get("repo");
  const user = q.get("user");

  let score: number | null = null;
  if (owner && repo) {
    const report = await getReport(owner, repo);
    score = report?.score ?? null;
  } else if (user) {
    const repos = (await getUserRepos(user)).filter((r) => !r.fork).slice(0, 12);
    const reports = await mapWithLimit(repos, 8, (r) => getReport(r.owner.login, r.name));
    score = averageScore(reports.filter((r) => r !== null).map((r) => r!.score));
  }

  return new Response(badgeSVG(score), {
    headers: {
      "Content-Type": "image/svg+xml; charset=utf-8",
      "Cache-Control": "public, max-age=0, s-maxage=3600, stale-while-revalidate=86400",
    },
  });
}
