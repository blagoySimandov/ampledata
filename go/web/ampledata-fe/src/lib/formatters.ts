export function formatDollars(cents: number) {
  return `$${(cents / 100).toFixed(0)}`;
}

export function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
    year: "numeric",
  });
}

export function formatOveragePrice(centsDecimal: string) {
  return `$${(parseFloat(centsDecimal) / 100).toFixed(3)}`;
}

export function usePct(used: number, included: number) {
  return Math.min((used / included) * 100, 100);
}
