export function creditLabel(score: number): { text: string; color: string } {
  if (score >= 180) return { text: '金牌守约者', color: '#faad14' }
  if (score >= 150) return { text: '银牌守约者', color: '#bfbfbf' }
  if (score >= 100) return { text: '普通用户', color: '#52c41a' }
  if (score >= 60) return { text: '信用警告', color: '#fa8c16' }
  return { text: '信用受限', color: '#f5222d' }
}

export function levelXp(level: number): { current: number; next: number } {
  const current = (level - 1) * (level - 1) * 100
  const next = level * level * 100
  return { current, next }
}

export function statusLabel(status: string): { text: string; color: string } {
  const map: Record<string, { text: string; color: string }> = {
    pending: { text: '待开始', color: 'blue' },
    active: { text: '进行中', color: 'green' },
    completed: { text: '已完成', color: 'gold' },
    failed: { text: '已失败', color: 'red' },
    cancelled: { text: '已取消', color: 'default' },
  }
  return map[status] || { text: status, color: 'default' }
}

export function categoryIcon(category: string): string {
  const map: Record<string, string> = {
    fitness: '💪',
    study: '📚',
    habit: '🎯',
    health: '🍎',
    social: '🤝',
    other: '📌',
  }
  return map[category] || '📌'
}
