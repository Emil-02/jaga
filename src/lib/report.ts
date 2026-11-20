export type Finding = {
  file: string;
  line: number;
  message: string;
  severity: string;
};

export type CheckResult = {
  name: string;
  status: string;
  severity: string;
  findings: Finding[];
  error?: string;
};

export type JagaReport = {
  root?: string;
  score: number;
  generated_at?: string;
  results: CheckResult[];
};

export function isReport(value: unknown): value is JagaReport {
  if (typeof value !== "object" || value === null) return false;
  const v = value as Record<string, unknown>;
  return typeof v.score === "number" && Array.isArray(v.results);
}

export function grade(score: number): string {
  if (score >= 90) return "A";
  if (score >= 80) return "B";
  if (score >= 60) return "C";
  if (score >= 40) return "D";
  return "F";
}

export function scoreColor(score: number): string {
  if (score < 50) return "#f87171";
  if (score < 80) return "#fbbf24";
  return "#4ade80";
}

export function findingCount(report: JagaReport): number {
  return report.results.reduce((total, result) => total + result.findings.length, 0);
}

export function warnCount(report: JagaReport): number {
  return report.results.filter((r) => r.status === "warn" || r.error).length;
}

export function averageScore(scores: number[]): number | null {
  if (scores.length === 0) return null;
  return Math.round(scores.reduce((a, b) => a + b, 0) / scores.length);
}
