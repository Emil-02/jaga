import Link from "next/link";
import { getReport, getUser, getUserRepos, mapWithLimit, type GithubRepo } from "@/lib/github";
import { averageScore, findingCount, grade, scoreColor, type JagaReport } from "@/lib/report";

type ScoredRepo = GithubRepo & { report: JagaReport | null };

export default async function ProfilePage({ params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  const [user, repos] = await Promise.all([getUser(username), getUserRepos(username)]);
  const selected = repos.filter((repo) => !repo.fork).slice(0, 24);
  const scored = await mapWithLimit(selected, 8, async (repo): Promise<ScoredRepo> => ({
    ...repo,
    report: await getReport(repo.owner.login, repo.name),
  }));
  const withReport = scored.filter((repo): repo is ScoredRepo & { report: JagaReport } => repo.report !== null);
  const scores = withReport.map((repo) => repo.report.score);
  const overall = averageScore(scores);
  const color = overall === null ? "#6b7280" : scoreColor(overall);
  const displayName = user?.name || username;

  return (
    <main className="mx-auto max-w-5xl px-5 py-12">
      <div className="grid gap-8 lg:grid-cols-[1fr_360px]">
        <section>
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-good">profil github</p>
          <h1 className="mt-2 text-3xl font-bold tracking-tight">{displayName}</h1>
          <p className="mt-1 text-muted">@{username}</p>

          <div className="mt-8 flex items-center gap-6">
            <div
              className="grid size-28 shrink-0 place-items-center rounded-full"
              style={{ background: overall === null ? "#1e293b" : `conic-gradient(${color} ${overall}%, #1e293b 0)` }}
            >
              <div className="grid size-[92px] place-items-center rounded-full bg-bg text-center">
                <div>
                  <p className="text-3xl font-bold leading-none" style={{ color }}>
                    {overall ?? "–"}
                  </p>
                  <p className="mt-1 text-[10px] text-muted">SKOR</p>
                </div>
              </div>
            </div>
            <div>
              <p className="text-sm text-muted">Skor rata-rata</p>
              <p className="mt-1 text-2xl font-semibold">
                {overall === null ? "Belum ada laporan" : `Grade ${grade(overall)}`}
              </p>
              <p className="mt-2 text-sm text-muted">
                {withReport.length} dari {selected.length} repo punya <code className="text-link">jaga-report.json</code>
              </p>
            </div>
          </div>

          <div className="mt-8">
            <h2 className="text-sm font-semibold uppercase tracking-widest text-muted">Repo</h2>
            {withReport.length === 0 ? (
              <div className="mt-4 rounded-xl border border-dashed border-border bg-surface p-6 text-sm text-muted">
                Belum ada repo dengan laporan jaga. Tambahkan Action jaga ke repo, commit{" "}
                <code className="text-link">jaga-report.json</code>, lalu muat ulang halaman ini.
              </div>
            ) : (
              <ul className="mt-4 grid gap-3 sm:grid-cols-2">
                {withReport.map((repo) => (
                  <li key={repo.name}>
                    <Link
                      href={`/r/${encodeURIComponent(repo.owner.login)}/${encodeURIComponent(repo.name)}`}
                      className="flex items-center justify-between gap-3 rounded-xl border border-border bg-surface p-4 transition hover:border-good"
                    >
                      <span className="min-w-0">
                        <span className="block truncate font-medium">{repo.name}</span>
                        <span className="block text-xs text-muted">{findingCount(repo.report)} temuan</span>
                      </span>
                      <span className="shrink-0 text-xl font-bold" style={{ color: scoreColor(repo.report.score) }}>
                        {repo.report.score}
                      </span>
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>

        <aside className="space-y-5">
          <div
            role="img"
            aria-label={`Kartu skor jaga ${username}`}
            className="aspect-[560/150] w-full rounded-xl border border-border bg-cover bg-center"
            style={{ backgroundImage: `url(/api/card?user=${encodeURIComponent(username)})` }}
          />
          <div className="rounded-xl border border-border bg-surface p-5">
            <h2 className="text-sm font-semibold">Pasang di README</h2>
            <pre className="mt-3 overflow-x-auto rounded-lg bg-bg p-3 text-xs text-muted">{`[![jaga](https://jaga-card.vercel.app/api/badge?user=${username})](https://jaga-card.vercel.app/u/${username})`}</pre>
            <p className="mt-3 text-xs text-muted">Skor dihitung dari repo yang sudah punya laporan jaga.</p>
          </div>
        </aside>
      </div>
    </main>
  );
}
