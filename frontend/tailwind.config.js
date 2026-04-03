/** @type {import('tailwindcss').Config} */
// Configuración de TailwindCSS para backup-smart
// Modo oscuro activado por clase para compatibilidad con el tema dark de SysAdmins
module.exports = {
  darkMode: 'class',
  content: [
    './src/**/*.{html,ts}',
  ],
  theme: {
    extend: {
      colors: {
        // Paleta de colores personalizada para el tema terminal/dark
        terminal: {
          bg: '#0d1117',
          surface: '#161b22',
          border: '#30363d',
          text: '#c9d1d9',
          muted: '#8b949e',
          green: '#3fb950',
          blue: '#58a6ff',
          yellow: '#d29922',
          red: '#f85149',
          purple: '#bc8cff',
        },
      },
      fontFamily: {
        // Fuente monospace para la consola de logs
        mono: ['"JetBrains Mono"', '"Fira Code"', 'Consolas', 'monospace'],
      },
    },
  },
  plugins: [],
}

