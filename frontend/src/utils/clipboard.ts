/**
 * 复制文本到剪贴板
 * 优先使用 Clipboard API（需要 HTTPS 或 localhost）
 * fallback 使用 execCommand（兼容 HTTP 环境）
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  // 优先尝试 Clipboard API
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch {
      // Clipboard API 失败，尝试 fallback
    }
  }
  // Fallback: 使用 textarea + execCommand
  try {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.left = '-9999px'
    textarea.style.top = '-9999px'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.focus()
    textarea.select()
    const result = document.execCommand('copy')
    document.body.removeChild(textarea)
    return result
  } catch {
    return false
  }
}
