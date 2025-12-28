const THEME_KEY = 'file-locker-theme';

export const getTheme = () => {
  return localStorage.getItem(THEME_KEY) || 'light';
};

export const saveTheme = (theme) => {
  localStorage.setItem(THEME_KEY, theme);
  applyTheme(theme);
};

export const toggleTheme = () => {
  const currentTheme = getTheme();
  const newTheme = currentTheme === 'light' ? 'dark' : 'light';
  saveTheme(newTheme);
  return newTheme;
};

export const applyTheme = (theme) => {
  document.documentElement.setAttribute('data-theme', theme);
};

export const initTheme = () => {
  const theme = getTheme();
  applyTheme(theme);
  return theme;
};
