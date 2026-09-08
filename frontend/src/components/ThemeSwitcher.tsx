import { useEffect, useState } from 'react'

// daisyUI標準テーマの一覧(v5で追加されたcaramellatte, abyss, silkを含む35種)。
// 増減させる場合はsrc/index.cssの@plugin "daisyui" { themes: ... } と合わせて変更すること。
const THEMES = [
    'light', 'dark', 'cupcake', 'bumblebee', 'emerald', 'corporate',
  'synthwave', 'retro', 'cyberpunk', 'valentine', 'halloween', 'garden',
  'forest', 'aqua', 'lofi', 'pastel', 'fantasy', 'wireframe', 'black',
  'luxury', 'dracula', 'cmyk', 'autumn', 'business', 'acid', 'lemonade',
  'night', 'coffee', 'winter', 'dim', 'nord', 'sunset',
  'caramellatte', 'abyss', 'silk',
] as const

const STORAGE_KEY = 'theme'

function getInitialTheme(): string {
    return localStorage.getItem(STORAGE_KEY) ?? 'silk'
}

export default function ThemeSwitcher() {
    const [theme, setTheme] = useState(getInitialTheme)

    // index.htmlの初期化スクリプトで初回描画時のちらつきは防いでいるが、
  // ユーザーが選択を変えた際にもここで<html>のdata-theme属性を更新する。
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem(STORAGE_KEY, theme)
  }, [theme])

  return (
    <select
        className="select select-bordered select-sm"
        value={theme}
        onChange={(e) => setTheme(e.target.value)}
        aria-label="select theme"
    >
        {THEMES.map((t) => (
            <option key={t} value={t}>
                {t}
            </option>
        ))}
    </select>
  )
}