import { NextResponse, type NextRequest } from "next/server";

export async function GET(request: NextRequest) {
  const raw = (request.nextUrl.searchParams.get("q") ?? "").trim();
  const parts = raw
    .replace(/^https?:\/\/github\.com\//, "")
    .replace(/\.git$/, "")
    .split("/")
    .filter(Boolean);

  if (parts.length >= 2) {
    return NextResponse.redirect(
      new URL(`/r/${encodeURIComponent(parts[0])}/${encodeURIComponent(parts[1])}`, request.url),
    );
  }
  if (parts.length === 1) {
    return NextResponse.redirect(new URL(`/u/${encodeURIComponent(parts[0])}`, request.url));
  }
  return NextResponse.redirect(new URL("/", request.url));
}
