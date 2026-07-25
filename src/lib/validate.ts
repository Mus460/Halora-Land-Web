export function validateEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

export function validatePassword(password: string): string | null {
  if (password.length < 8) return 'Minimal 8 karakter'
  if (!/[A-Z]/.test(password)) return 'Harus ada huruf besar'
  if (!/[a-z]/.test(password)) return 'Harus ada huruf kecil'
  if (!/[0-9]/.test(password)) return 'Harus ada angka'
  return null
}

export function parseIntParam(id: string | string[] | undefined): number | null {
  if (!id || Array.isArray(id)) return null
  const parsed = parseInt(id, 10)
  return isNaN(parsed) ? null : parsed
}
