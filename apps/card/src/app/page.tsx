import Link from "next/link";

export default function Home() {
  return (
    <main className="mx-auto max-w-5xl px-5 py-16">
      <section className="max-w-2xl">
        <p className="mb-3 text-xs font-semibold uppercase tracking-[0.2em] text-good">satu kartu untuk semua repo kamu</p>
        <h1 className="text-4xl font-bold tracking-tight sm:text-5xl">Skor kepercayaan, bukan cuma badge.</h1>
        <p className="mt-5 text-muted">
          Jaga Card mengumpulkan hasil <code className="text-link">jaga check</code> dari repo GitHub kamu — secret,
          data pribadi, dan hygiene — lalu menampilkannya sebagai kartu yang bisa ditempel di README atau dibagikan.
        </p>

        <form action="/cari" method="get" className="mt-8 flex flex-col gap-3 sm:flex-row">
          <input
            name="q"
            required
            autoComplete="off"
            placeholder="Emil-02 atau Emil-02/jaga"
            className="w-full rounded-lg border border-border bg-surface px-4 py-3 text-ink outline-none placeholder:text-muted focus:border-good"
          />
          <button
            type="submit"
            className="rounded-lg bg-good px-6 py-3 font-semibold text-bg transition hover:brightness-110"
          >
            Lihat kartu
          </button>
        </form>
        <p className="mt-3 text-sm text-muted">
          Contoh: <Link href="/u/Emil-02" className="text-link hover:underline">profil Emil-02</Link> ·{" "}
          <Link href="/r/Emil-02/jaga" className="text-link hover:underline">repo jaga</Link>
        </p>
      </section>

      <section className="mt-16 grid gap-4 sm:grid-cols-3">
        {[
          ["01", "Skor yang jelas", "Ringkasan 0–100 dari tiga pemeriksaan: secret, PII, dan hygiene repo."],
          ["02", "Badge siap pakai", "Endpoint SVG untuk skor repo maupun skor rata-rata profil."],
          ["03", "Tanpa server sendiri", "Cukup commit jaga-report.json dari CI; Jaga Card yang merender."],
        ].map(([n, title, body]) => (
          <div key={n} className="rounded-xl border border-border bg-surface p-5">
            <p className="text-xs font-mono text-good">{n}</p>
            <h2 className="mt-3 font-semibold">{title}</h2>
            <p className="mt-2 text-sm text-muted">{body}</p>
          </div>
        ))}
      </section>
    </main>
  );
}
