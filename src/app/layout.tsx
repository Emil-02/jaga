import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "Jaga Card — skor kepercayaan repo kamu",
  description:
    "Lihat skor keamanan, privasi, dan hygiene repo GitHub dari hasil jaga dalam satu kartu yang bisa dibagikan.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="id">
      <body className="min-h-screen bg-bg text-ink antialiased">
        <header className="border-b border-border">
          <div className="mx-auto flex max-w-5xl items-center justify-between px-5 py-4">
            <Link href="/" className="flex items-center gap-2 text-lg font-bold tracking-tight">
              <span className="inline-block size-2.5 rounded-full bg-good" />
              jaga<span className="text-muted">/card</span>
            </Link>
            <a
              href="https://github.com/Emil-02/jaga"
              target="_blank"
              rel="noreferrer"
              className="text-sm text-muted hover:text-link"
            >
              tentang jaga
            </a>
          </div>
        </header>
        {children}
        <footer className="mx-auto max-w-5xl px-5 py-10 text-xs text-muted">
          Dibuat oleh Emil-02 · tanpa database, tanpa telemetri · data diambil dari GitHub API.
        </footer>
      </body>
    </html>
  );
}
