import { NavLink } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { storage } from '../services/storage.js';

const links = [
  { to: '/', label: 'Dashboard', end: true },
  { to: '/compare', label: 'Сравнение' },
  { to: '/admin', label: 'Админ' },
];

export default function Header() {
  const [theme, setTheme] = useState(storage.getTheme());

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    storage.setTheme(theme);
  }, [theme]);

  return (
    <header className="app-header">
      <div className="brand">⚽ Спортивный трекер</div>
      <nav>
        {links.map((l) => (
          <NavLink key={l.to} to={l.to} end={l.end}
            className={({ isActive }) => (isActive ? 'active' : undefined)}>
            {l.label}
          </NavLink>
        ))}
      </nav>
      <button className="ghost" onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
        {theme === 'dark' ? '☀️' : '🌙'}
      </button>
    </header>
  );
}
