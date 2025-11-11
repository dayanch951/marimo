import { useEffect, useState } from 'react';
import { Theme } from '../types/theme.types';

const THEME_KEY = 'marimo-theme';

export const useTheme = () => {
  const [theme, setThemeState] = useState<Theme>(() => {
    const saved = localStorage.getItem(THEME_KEY);
    return (saved as Theme) || 'system';
  });

  const [effectiveTheme, setEffectiveTheme] = useState<'light' | 'dark'>('light');

  useEffect(() => {
    const root = document.documentElement;
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

    const applyTheme = () => {
      let themeToApply: 'light' | 'dark';

      if (theme === 'system') {
        themeToApply = mediaQuery.matches ? 'dark' : 'light';
      } else {
        themeToApply = theme;
      }

      setEffectiveTheme(themeToApply);
      root.classList.remove('light', 'dark');
      root.classList.add(themeToApply);
      root.setAttribute('data-theme', themeToApply);
    };

    applyTheme();

    const handleChange = () => {
      if (theme === 'system') {
        applyTheme();
      }
    };

    mediaQuery.addEventListener('change', handleChange);

    return () => {
      mediaQuery.removeEventListener('change', handleChange);
    };
  }, [theme]);

  const setTheme = (newTheme: Theme) => {
    setThemeState(newTheme);
    localStorage.setItem(THEME_KEY, newTheme);
  };

  const toggleTheme = () => {
    const newTheme = effectiveTheme === 'light' ? 'dark' : 'light';
    setTheme(newTheme);
  };

  return {
    theme,
    effectiveTheme,
    setTheme,
    toggleTheme,
  };
};
