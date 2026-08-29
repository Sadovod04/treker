// Small guarded localStorage helper for per-user UI preferences.

const KEY = {
  token: 'st.token',
  lastPlayer: 'st.lastPlayer',
  theme: 'st.theme',
};

function get(key, fallback = null) {
  try {
    const v = localStorage.getItem(key);
    return v === null ? fallback : v;
  } catch {
    return fallback;
  }
}

function set(key, value) {
  try {
    if (value === null || value === undefined) localStorage.removeItem(key);
    else localStorage.setItem(key, String(value));
  } catch {
    /* private mode / disabled storage — ignore */
  }
}

export const storage = {
  getToken: () => get(KEY.token, ''),
  setToken: (t) => set(KEY.token, t),
  getLastPlayer: () => {
    const v = get(KEY.lastPlayer);
    return v ? Number(v) : null;
  },
  setLastPlayer: (id) => set(KEY.lastPlayer, id),
  getTheme: () => get(KEY.theme, 'light'),
  setTheme: (t) => set(KEY.theme, t),
};
