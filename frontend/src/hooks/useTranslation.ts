import { useState, useEffect } from 'react';
import zhCN from '../locales/zh-CN.json';
import en from '../locales/en.json';

type Translations = typeof zhCN;

const translationsMap = {
  'zh-CN': zhCN,
  'en': en
};

export interface UseTranslationReturn {
  t: (key: string) => string;
  language: string;
  setLanguage: (newLanguage: string) => void;
}

const useTranslation = (): UseTranslationReturn => {
  const [translations, setTranslations] = useState<Translations>(zhCN);
  const [language, setLanguage] = useState('zh-CN');

  // 从localStorage加载语言设置
  useEffect(() => {
    const savedLanguage = localStorage.getItem('language');
    if (savedLanguage && translationsMap[savedLanguage as keyof typeof translationsMap]) {
      setLanguage(savedLanguage);
      setTranslations(translationsMap[savedLanguage as keyof typeof translationsMap]);
    }
  }, []);

  const changeLanguage = (newLanguage: string) => {
    if (translationsMap[newLanguage as keyof typeof translationsMap]) {
      setLanguage(newLanguage);
      setTranslations(translationsMap[newLanguage as keyof typeof translationsMap]);
      localStorage.setItem('language', newLanguage);
    }
  };

  const t = (key: string): string => {
    // 简单的键路径解析
    const keys = key.split('.');
    let result: any = translations;

    for (const k of keys) {
      if (result && typeof result === 'object' && k in result) {
        result = result[k];
      } else {
        return key; // 如果找不到翻译，返回原始键
      }
    }

    return result as string;
  };

  return { t, language, setLanguage: changeLanguage };
};

export default useTranslation;