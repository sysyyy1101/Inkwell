/** 服务端时间可能是 unix 秒(列表接口)或 RFC3339 字符串(详情接口), 统一转成毫秒时间戳 */
export function toTimestamp(value: number | string | null | undefined): number {
  if (value === null || value === undefined) {
    return 0
  }
  if (typeof value === 'number') {
    // 秒级时间戳大约 1e9, 毫秒级大约 1e12
    return value < 1e12 ? value * 1000 : value
  }
  const parsed = Date.parse(value)
  return Number.isNaN(parsed) ? 0 : parsed
}

/** 相对时间: 刚刚 / 12 分钟前 / 3 小时前 / 昨天 / 5 天前 / 2026-03-01 */
export function formatRelativeTime(value: number | string | null | undefined): string {
  const timestamp = toTimestamp(value)
  if (!timestamp) {
    return ''
  }
  const diff = Date.now() - timestamp
  const minute = 60 * 1000
  const hour = 60 * minute
  const day = 24 * hour

  if (diff < 0) {
    return '刚刚'
  }
  if (diff < minute) {
    return '刚刚'
  }
  if (diff < hour) {
    return `${Math.floor(diff / minute)} 分钟前`
  }
  if (diff < day) {
    return `${Math.floor(diff / hour)} 小时前`
  }
  if (diff < 2 * day) {
    return '昨天'
  }
  if (diff < 7 * day) {
    return `${Math.floor(diff / day)} 天前`
  }
  return formatDate(timestamp)
}

/** 绝对时间: 2026-03-01 14:05 */
export function formatDateTime(value: number | string | null | undefined): string {
  const timestamp = toTimestamp(value)
  if (!timestamp) {
    return ''
  }
  const date = new Date(timestamp)
  return `${formatDate(timestamp)} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

export function formatDate(value: number | string | null | undefined): string {
  const timestamp = toTimestamp(value)
  if (!timestamp) {
    return ''
  }
  const date = new Date(timestamp)
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

function pad(value: number): string {
  return value < 10 ? `0${value}` : String(value)
}

/** 计数展示: 1234 -> 1,234; 23456 -> 2.3 万 */
export function formatCount(value: number | null | undefined): string {
  const count = value ?? 0
  if (count >= 10000) {
    return `${(count / 10000).toFixed(1).replace(/\.0$/, '')} 万`
  }
  return count.toLocaleString('zh-CN')
}

/** 取文本摘要, 超长时截断并补省略号 */
export function truncateText(text: string, max = 160): string {
  const normalized = text.replace(/\s+/g, ' ').trim()
  return normalized.length > max ? `${normalized.slice(0, max)}…` : normalized
}

/** 头像占位文字: 中文取前 1 个字, 英文取首字母 */
export function avatarText(name: string): string {
  const trimmed = (name || '').trim()
  if (!trimmed) {
    return '?'
  }
  return /[\u4e00-\u9fa5]/.test(trimmed[0]) ? trimmed[0] : trimmed[0].toUpperCase()
}

/**
 * 头像与色点用的低饱和色相。按名字哈希稳定取一个,
 * 具体颜色交给 CSS: 深浅主题下透明度/明度不同, 在这里写死颜色会不好维护。
 */
const AVATAR_TONES = [
  { hue: 214, sat: 46 },
  { hue: 158, sat: 40 },
  { hue: 24, sat: 48 },
  { hue: 266, sat: 42 },
  { hue: 190, sat: 44 },
  { hue: 42, sat: 46 },
]

export function avatarTone(name: string): Record<string, string> {
  let hash = 0
  for (let i = 0; i < name.length; i += 1) {
    hash = (hash * 31 + name.charCodeAt(i)) % 100003
  }
  const tone = AVATAR_TONES[hash % AVATAR_TONES.length]
  return {
    '--tone-hue': String(tone.hue),
    '--tone-sat': `${tone.sat}%`,
  }
}

/** 后端限制: 发帖超过一周的帖子不能再投票(dao/redis 里的 OneWeekInSeconds) */
export const VOTE_WINDOW_MS = 7 * 24 * 60 * 60 * 1000

/** 投票窗口是否还开着, 用来在按钮上提前给出提示, 不用等接口报错 */
export function isVoteWindowOpen(createTime: number | string | null | undefined): boolean {
  const timestamp = toTimestamp(createTime)
  if (!timestamp) {
    return true
  }
  return Date.now() - timestamp <= VOTE_WINDOW_MS
}
