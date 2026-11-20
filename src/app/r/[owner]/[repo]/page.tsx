import Link from "next/link";
import { getReport } from "@/lib/github";
import { findingCount, grade, scoreColor, warnCount } from "@/lib/report";

export default async function RepoPage({ params }: { params: Promise<{ owner: string; repo: string }> }) {
  const { owner, repo } = await params;
  const report = await getReport(owner, repo);
  const fullName = `${owner}/${repo}`;

  if (!report) {
    return (
      <main className="mx-auto max-w-3xl px-5 py-16">
        <h1 className="text-3xl font-bold tracking-tight">{fullName}</h1>
        <div className="mt-8 rounded-xl border border-dashed border-border bg-surface p-6">
          <h2 className="font-semibold">Repo ini belum punya laporan jaga</h2>
          <p className="mt-2 text-sm text-muted">
            Jalankan <code className="text-link">jaga check . --format json -o jaga-report.json</code> di repo,
            commit file-nya, lalu muat ulang halaman ini.
          </p>
          <pre className="mt-4 overflow-x-auto rounded-lg bg-bg p-4 text-xs text-muted">{`jaga check . --format json -o jaga-report.json
git add jaga-report.json && git commit -m "chore: add jaga report"`}</pre>
        </div>
        <Link href={`/u/${encodeURIComponent(owner)}`} className="mt-6 inline-block text-sm text-link hover:underline">
          ← lihat profil {owner}
        </Link>
      </main>
    );
  }

  const color = scoreColor(report.score);
  const scanned = report.generated_at
    ? new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeStyle: "short" }).format(new Date(report.generated_at))
    : null;

  return (
    <main className="mx-auto max-w-5xl px-5 py-12">
      <p className="text-xs font-semibold uppercase tracking-[0.2em] text-good">laporan jaga</p>
      <h1 className="mt-2 text-3xl font-bold tracking-tight break-all">{fullName}</h1>

      <div className="mt-8 grid gap-6 lg:grid-cols-[220px_1fr]">
        <div className="space-y-5">
          <div
            className="grid size-44 place-items-center rounded-full"
            style={{ background: `conic-gradient(${color} ${report.score}%, #1e293b 0)` }}
          >
            <div className="grid size-[136px] place-items-center rounded-full bg-bg text-center">
              <div>
                <p className="text-5xl font-bold leading-none" style={{ color }}>
                  {report.score}
                </p>
                <p className="mt-2 text-xs text-muted">DARI 100</p>
              </div>
            </div>
          </div>
          <div
            role="img"
            aria-label={`Badge skor jaga ${fullName}`}
            className="h-5 w-[106px] rounded bg-cover bg-center"
            style={{ backgroundImage: `url(/api/badge?owner=${encodeURIComponent(owner)}&repo=${encodeURIComponent(repo)})` }}
          />
          <pre className="overflow-x-auto rounded-lg border border-border bg-surface p-3 text-[11px] text-muted">{`[![jaga](https://jaga-card.vercel.app/api/badge?owner=${owner}&repo=${repo})](https://jaga-card.vercel.app/r/${owner}/${repo})`}</pre>
        </div>

        <section>
          <div className="grid grid-cols-3 gap-3">
            {[
              ["Grade", grade(report.score)],
              ["Temuan", String(findingCount(report))],
              ["Perlu perhatian", String(warnCount(report))],
            ].map(([label, value]) => (
              <div key={label} className="rounded-xl border border-border bg-surface p-4">
                <p className="text-xs text-muted">{label}</p>
                <p className="mt-1 text-2xl font-semibold">{value}</p>
              </div>
            ))}
          </div>
          {scanned && <p className="mt-3 text-xs text-muted">Terakhir dipindai: {scanned}</p>}

          <h2 className="mt-8 text-sm font-semibold uppercase tracking-widest text-muted">Pemeriksaan</h2>
          <div className="mt-4 space-y-3">
            {report.results.map((result) => (
              <article key={result.name} className="rounded-xl border border-border bg-surface p-5">
                <div className="flex items-center gap-3">
                  <h3 className="font-mono font-semibold">{result.name}</h3>
                  <span className="rounded-full border border-border px-2 py-0.5 text-[10px] uppercase tracking-wider text-muted">
                    {result.severity}
                  </span>
                  <span className="ml-auto text-xs text-muted">
                    {result.status === "warn" ? "peringatan" : result.error ? "error" : "lolos"}
                  </span>
                </div>
                {result.findings.length > 0 ? (
                  <ul className="mt-4 space-y-2">
                    {result.findings.map((finding, index) => (
                      <li key={`${finding.file}-${finding.line}-${index}`} className="text-sm">
                        <code className="mr-2 text-link">
                          {finding.file}
                          {finding.line > 0 ? `:${finding.line}` : ""}
                        </code>
                        <span className="text-muted">{finding.message}</span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="mt-4 text-sm text-muted">Tidak ada temuan.</p>
                )}
                {result.error && <p className="mt-2 text-sm text-bad">{result.error}</p>}
              </article>
            ))}
          </div>
        </section>
      </div>

      <a
        href={`https://github.com/${owner}/${repo}`}
        target="_blank"
        rel="noreferrer"
        className="mt-10 inline-block text-sm text-link hover:underline"
      >
        Buka repo di GitHub ↗
      </a>
    </main>
  );
}
