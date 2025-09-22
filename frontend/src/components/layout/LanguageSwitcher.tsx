import React from 'react';
import { Nav } from 'react-bootstrap';
import useTranslation from '../../hooks/useTranslation';

const LanguageSwitcher: React.FC = () => {
  const { language, setLanguage } = useTranslation();

  const languages = [
    { code: 'zh-CN', name: '中文', flag: '🇨🇳' },
    { code: 'en', name: 'English', flag: '🇺🇸' }
  ];

  const handleLanguageChange = (langCode: string) => {
    setLanguage(langCode);
  };

  return (
    <Nav className="language-switcher">
      {languages.map((lang) => (
        <Nav.Link
          key={lang.code}
          className={`language-option ${language === lang.code ? 'active' : ''}`}
          onClick={() => handleLanguageChange(lang.code)}
          title={lang.name}
        >
          <span className="language-flag">{lang.flag}</span>
          <span className="language-name ms-1">{lang.name}</span>
        </Nav.Link>
      ))}
    </Nav>
  );
};

export default LanguageSwitcher;