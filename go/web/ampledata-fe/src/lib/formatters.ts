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
