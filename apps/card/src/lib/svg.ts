import { grade, scoreColor } from "./report";

export function escapeXml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}

export function badgeSVG(score: number | null): string {
  const color = score === null ? "#6b7280" : scoreColor(score);
  const label = score === null ? "–/100" : `${score}/100`;
  return `<svg xmlns="http://www.w3.org/2000/svg" width="106" height="20" role="img" aria-label="jaga: ${label}">
  <linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient>
  <clipPath id="r"><rect width="106" height="20" rx="3" fill="#fff"/></clipPath>
  <g clip-path="url(#r)"><rect width="42" height="20" fill="#555"/><rect x="42" width="64" height="20" fill="${color}"/><rect width="106" height="20" fill="url(#s)"/></g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="11">
    <text x="21" y="15">jaga</text><text x="74" y="15">${label}</text>
  </g>
</svg>`;
}

export function cardSVG(opts: {
  username: string;
  score: number | null;
  repos: number;
}): string {
  const name = escapeXml(opts.username);
  const color = opts.score === null ? "#6b7280" : scoreColor(opts.score);
  const value = opts.score === null ? "–" : String(opts.score);
  const badge = opts.score === null ? "?" : grade(opts.score);
  const caption = opts.repos === 1 ? "1 repo terpindai" : `${opts.repos} repo terpindai`;

  return `<svg xmlns="http://www.w3.org/2000/svg" width="560" height="150" role="img" aria-label="Skor jaga ${name}: ${value}">
  <rect width="560" height="150" rx="14" fill="#0b1120"/>
  <rect x="0.5" y="0.5" width="559" height="149" rx="13.5" fill="none" stroke="#1e293b"/>
  <circle cx="28" cy="30" r="6" fill="#4ade80"/>
  <text x="44" y="35" fill="#e2e8f0" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="15" font-weight="bold">jaga</text>
  <text x="532" y="35" fill="#64748b" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="10" text-anchor="end">KARTU KEPERCAYAAN</text>
  <text x="28" y="82" fill="#e2e8f0" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="24" font-weight="bold">${name}</text>
  <text x="420" y="98" fill="${color}" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="42" font-weight="bold" text-anchor="end">${value}</text>
  <rect x="436" y="62" width="42" height="24" rx="12" fill="${color}" fill-opacity="0.15" stroke="${color}" stroke-opacity="0.5"/>
  <text x="457" y="79" fill="${color}" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="13" font-weight="bold" text-anchor="middle">${badge}</text>
  <text x="28" y="116" fill="#94a3b8" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="12">${escapeXml(caption)}</text>
  <text x="28" y="136" fill="#475569" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="10">github.com/Emil-02/jaga-card</text>
</svg>`;
}
