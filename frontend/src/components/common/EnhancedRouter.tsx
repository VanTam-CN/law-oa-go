import React from 'react';
import {
  BrowserRouter,
  BrowserRouterProps
} from 'react-router-dom';

interface EnhancedRouterProps extends BrowserRouterProps {
  useHistoryRouter?: boolean;
}

const EnhancedRouter: React.FC<EnhancedRouterProps> = ({
  children,
  ...props
}) => {
  // 正确配置future flags来消除警告
  const routerProps = {
    ...props,
    future: {
      v7_startTransition: true,
      v7_relativeSplatPath: true
    }
  };

  return (
    <BrowserRouter {...routerProps}>
      {children}
    </BrowserRouter>
  );
};

export default EnhancedRouter;