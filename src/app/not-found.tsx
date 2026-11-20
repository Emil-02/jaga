import Link from "next/link";

export default function NotFound() {
  return (
    <main className="mx-auto max-w-3xl px-5 py-20 text-center">
      <p className="text-xs font-mono text-good">404</p>
      <h1 className="mt-3 text-3xl font-bold">Halaman tidak ditemukan</h1>
      <p className="mt-3 text-muted">Coba cari username GitHub atau owner/repo dari halaman utama.</p>
      <Link href="/" className="mt-6 inline-block rounded-lg bg-good px-5 py-2.5 font-semibold text-bg">
        Kembali
      </Link>
    </main>
  );
}
