import type { NextRequest } from "next/server";
import { getReport, getUserRepos, mapWithLimit } from "@/lib/github";
import { averageScore } from "@/lib/report";
import { cardSVG } from "@/lib/svg";

export async function GET(request: NextRequest) {
  const username = (request.nextUrl.searchParams.get("user") ?? "").trim() || "github";
  const repos = (await getUserRepos(username)).filter((r) => !r.fork).slice(0, 12);
  const reports = await mapWithLimit(repos, 8, (r) => getReport(r.owner.login, r.name));
  const score = averageScore(reports.filter((r) => r !== null).map((r) => r!.score));

  return new Response(cardSVG({ username, score, repos: reports.filter((r) => r !== null).length }), {
    headers: {
      "Content-Type": "image/svg+xml; charset=utf-8",
      "Cache-Control": "public, max-age=0, s-maxage=3600, stale-while-revalidate=86400",
    },
  });
}
