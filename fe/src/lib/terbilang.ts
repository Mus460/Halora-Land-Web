const ANGKA = [
  "",
  "satu",
  "dua",
  "tiga",
  "empat",
  "lima",
  "enam",
  "tujuh",
  "delapan",
  "sembilan",
  "sepuluh",
  "sebelas",
];

function words(n: number): string {
  if (n < 12) return ANGKA[n];
  if (n < 20) return `${ANGKA[n - 10]} belas`;
  if (n < 100) return `${ANGKA[Math.floor(n / 10)]} puluh ${ANGKA[n % 10]}`.trim();
  if (n < 200) return `seratus ${words(n - 100)}`.trim();
  if (n < 1000) return `${ANGKA[Math.floor(n / 100)]} ratus ${words(n % 100)}`.trim();
  if (n < 2000) return `seribu ${words(n - 1000)}`.trim();
  if (n < 1000000) return `${words(Math.floor(n / 1000))} ribu ${words(n % 1000)}`.trim();
  if (n < 1000000000) {
    return `${words(Math.floor(n / 1000000))} juta ${words(n % 1000000)}`.trim();
  }
  return `${words(Math.floor(n / 1000000000))} miliar ${words(n % 1000000000)}`.trim();
}

export function terbilang(amount: number): string {
  const value = Math.round(Math.abs(amount));
  if (value === 0) return "Nol Rupiah";
  const w = words(value).trim();
  return w.charAt(0).toUpperCase() + w.slice(1) + " Rupiah";
}