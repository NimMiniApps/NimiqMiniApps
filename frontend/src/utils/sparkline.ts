export function buildSparklinePoints(
  daily: { date: string; value: number }[],
  width: number,
  height: number,
): string {
  if (daily.length === 0) return ''
  const values = daily.map((d) => d.value)
  const max = Math.max(...values, 1)
  const pad = 2
  const innerW = width - pad * 2
  const innerH = height - pad * 2
  if (daily.length === 1) {
    const y = pad + innerH / 2
    return `${pad},${y} ${pad + innerW},${y}`
  }
  return daily
    .map((d, i) => {
      const x = pad + (i / (daily.length - 1)) * innerW
      const y = pad + innerH - (d.value / max) * innerH
      return `${x},${y}`
    })
    .join(' ')
}
